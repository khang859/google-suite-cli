package auth

import (
	"strings"
	"testing"
)

func TestExtractOAuth2ClientCreds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		jsonData       string
		wantClientID   string
		wantClientSec  string
		wantErr        bool
		wantErrContain string
	}{
		{
			name:          "should parse installed format",
			jsonData:      `{"installed":{"client_id":"id123","client_secret":"sec456"}}`,
			wantClientID:  "id123",
			wantClientSec: "sec456",
		},
		{
			name:          "should parse web format",
			jsonData:      `{"web":{"client_id":"webid","client_secret":"websec"}}`,
			wantClientID:  "webid",
			wantClientSec: "websec",
		},
		{
			name:           "should error when both keys are missing",
			jsonData:       `{"other":{}}`,
			wantErr:        true,
			wantErrContain: "neither",
		},
		{
			name:           "should error when client_id is empty",
			jsonData:       `{"installed":{"client_id":"","client_secret":"sec"}}`,
			wantErr:        true,
			wantErrContain: "client_id is empty",
		},
		{
			name:           "should error when JSON is invalid",
			jsonData:       `not json`,
			wantErr:        true,
			wantErrContain: "parse",
		},
		{
			name:          "should allow empty client_secret",
			jsonData:      `{"installed":{"client_id":"id","client_secret":""}}`,
			wantClientID:  "id",
			wantClientSec: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clientID, clientSecret, err := extractOAuth2ClientCreds([]byte(tt.jsonData))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if clientID != tt.wantClientID {
				t.Errorf("clientID = %q, want %q", clientID, tt.wantClientID)
			}
			if clientSecret != tt.wantClientSec {
				t.Errorf("clientSecret = %q, want %q", clientSecret, tt.wantClientSec)
			}
		})
	}
}
