package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCredentials(t *testing.T) {
	const testJSON = `{"installed":{"client_id":"test"}}`

	tests := []struct {
		name           string
		googleCreds    string
		googleAppCreds string
		writeFile      bool
		wantErr        bool
		wantBytes      string
	}{
		{
			name:        "should load from GOOGLE_CREDENTIALS env var",
			googleCreds: testJSON,
			wantBytes:   testJSON,
		},
		{
			name:      "should load from GOOGLE_APPLICATION_CREDENTIALS file",
			writeFile: true,
			wantBytes: testJSON,
		},
		{
			name:        "should prioritize GOOGLE_CREDENTIALS over GOOGLE_APPLICATION_CREDENTIALS",
			googleCreds: testJSON,
			writeFile:   true,
			wantBytes:   testJSON,
		},
		{
			name:    "should error when neither env var is set",
			wantErr: true,
		},
		{
			name:           "should error when GOOGLE_APPLICATION_CREDENTIALS file not found",
			googleAppCreds: "/nonexistent/path/creds.json",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOOGLE_CREDENTIALS", "")
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

			if tt.googleCreds != "" {
				t.Setenv("GOOGLE_CREDENTIALS", tt.googleCreds)
			}

			if tt.writeFile {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "creds.json")
				if err := os.WriteFile(filePath, []byte(testJSON), 0600); err != nil {
					t.Fatalf("failed to write temp creds file: %v", err)
				}
				t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filePath)
			} else if tt.googleAppCreds != "" {
				t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.googleAppCreds)
			}

			got, err := LoadCredentials()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.wantBytes {
				t.Errorf("got %q, want %q", string(got), tt.wantBytes)
			}
		})
	}
}
