package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const (
	// redirectURL is the local callback URL for the OAuth2 flow.
	redirectURL = "http://localhost:8089/callback"
	// callbackAddr is the address the local HTTP server listens on.
	callbackAddr = ":8089"
	// authTimeout is the maximum time to wait for the user to complete authentication.
	authTimeout = 2 * time.Minute
)

// OAuth2Config holds the configuration for OAuth2 authorization code flow with PKCE.
type OAuth2Config struct {
	config *oauth2.Config
}

// NewOAuth2Config creates a new OAuth2Config with the given client credentials.
// ClientID and ClientSecret should come from a Google OAuth2 client configuration;
// they must not be hardcoded.
func NewOAuth2Config(clientID, clientSecret string) *OAuth2Config {
	return &OAuth2Config{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes:       []string{gmail.GmailModifyScope},
		},
	}
}

// Authenticate performs the full OAuth2 authorization code flow with PKCE.
// It starts a local HTTP server, opens the browser for user consent, and
// exchanges the authorization code for a token.
func (c *OAuth2Config) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	// Generate PKCE code verifier (32 random bytes, base64url no padding)
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE code verifier: %w", err)
	}

	// Generate code challenge (SHA256 of verifier, base64url no padding)
	challenge := generateCodeChallenge(verifier)

	// Generate random state parameter (16 bytes, hex-encoded)
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state parameter: %w", err)
	}

	// Build authorization URL with PKCE parameters
	authURL := c.config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	// Channel to receive the authorization code from the callback
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Set up callback handler
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		queryState := r.URL.Query().Get("state")
		if queryState != state {
			errCh <- fmt.Errorf("state mismatch: expected %s, got %s", state, queryState)
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code in callback")
			http.Error(w, "No authorization code", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authentication successful!</h1><p>You can close this tab.</p></body></html>")
		codeCh <- code
	})

	// Start local HTTP server
	listener, err := net.Listen("tcp", callbackAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to start local server on %s: %w", callbackAddr, err)
	}

	server := &http.Server{Handler: mux}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- fmt.Errorf("local server error: %w", serveErr)
		}
	}()

	// Ensure server is shut down when we return
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx) //nolint:errcheck
	}()

	// Open browser or print URL
	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If the browser does not open, visit this URL:\n%s\n\n", authURL)
	openBrowser(authURL)

	// Wait for authorization code with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
		// Success — got the code
	case err := <-errCh:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("timed out waiting for authentication callback (timeout: %s)", authTimeout)
	}

	// Exchange authorization code for token with PKCE verifier
	token, err := c.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for token: %w", err)
	}

	return token, nil
}

// NewGmailService creates an authenticated Gmail service from an existing OAuth2 token.
func (c *OAuth2Config) NewGmailService(ctx context.Context, token *oauth2.Token) (*gmail.Service, error) {
	tokenSource := c.config.TokenSource(ctx, token)
	client := oauth2.NewClient(ctx, tokenSource)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return service, nil
}

// generateCodeVerifier generates a PKCE code verifier from 32 random bytes,
// encoded as base64url without padding.
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge computes the PKCE code challenge as the SHA256 hash
// of the verifier, encoded as base64url without padding.
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateState generates a random state parameter from 16 bytes, hex-encoded.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// openBrowser attempts to open a URL in the user's default browser.
// It tries platform-specific commands and silently falls back to doing nothing
// (the URL is already printed to the terminal).
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	// Fire and forget — if it fails, user already has the URL printed
	cmd.Start() //nolint:errcheck
}
