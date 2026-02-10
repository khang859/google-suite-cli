package auth

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	t.Parallel()

	t.Run("should return non-empty result", func(t *testing.T) {
		t.Parallel()

		v, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v == "" {
			t.Fatal("expected non-empty verifier")
		}
	})

	t.Run("should return correct length", func(t *testing.T) {
		t.Parallel()

		v, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 32 bytes base64url-encoded without padding = 43 characters
		if len(v) != 43 {
			t.Errorf("verifier length = %d, want 43", len(v))
		}
	})

	t.Run("should produce different results on two calls", func(t *testing.T) {
		t.Parallel()

		v1, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v2, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v1 == v2 {
			t.Errorf("two calls produced identical verifiers: %q", v1)
		}
	})

	t.Run("should produce valid base64url encoding", func(t *testing.T) {
		t.Parallel()

		v, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = base64.RawURLEncoding.DecodeString(v)
		if err != nil {
			t.Errorf("verifier is not valid base64url: %v", err)
		}
	})
}

func TestGenerateCodeChallenge(t *testing.T) {
	t.Parallel()

	t.Run("should produce deterministic output", func(t *testing.T) {
		t.Parallel()

		verifier := "test-verifier-value"
		c1 := generateCodeChallenge(verifier)
		c2 := generateCodeChallenge(verifier)
		if c1 != c2 {
			t.Errorf("same input produced different challenges: %q vs %q", c1, c2)
		}
	})

	t.Run("should match RFC 7636 test vector", func(t *testing.T) {
		t.Parallel()

		// From RFC 7636 Appendix B
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

		got := generateCodeChallenge(verifier)
		if got != want {
			t.Errorf("generateCodeChallenge(%q) = %q, want %q", verifier, got, want)
		}
	})
}

func TestGenerateState(t *testing.T) {
	t.Parallel()

	t.Run("should return 32 hex chars", func(t *testing.T) {
		t.Parallel()

		s, err := generateState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 16 bytes hex-encoded = 32 characters
		if len(s) != 32 {
			t.Errorf("state length = %d, want 32", len(s))
		}
	})

	t.Run("should produce different results on two calls", func(t *testing.T) {
		t.Parallel()

		s1, err := generateState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s2, err := generateState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s1 == s2 {
			t.Errorf("two calls produced identical states: %q", s1)
		}
	})

	t.Run("should produce valid hex encoding", func(t *testing.T) {
		t.Parallel()

		s, err := generateState()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = hex.DecodeString(s)
		if err != nil {
			t.Errorf("state is not valid hex: %v", err)
		}
	})
}
