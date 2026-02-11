package auth

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/api/googleapi"
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

func TestIsInsufficientScopeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "should return true for 403 with insufficientPermissions",
			err: &googleapi.Error{
				Code:   403,
				Errors: []googleapi.ErrorItem{{Reason: "insufficientPermissions"}},
			},
			want: true,
		},
		{
			name: "should return false for 403 with different reason",
			err: &googleapi.Error{
				Code:   403,
				Errors: []googleapi.ErrorItem{{Reason: "forbidden"}},
			},
			want: false,
		},
		{
			name: "should return false for 401 error",
			err: &googleapi.Error{
				Code: 401,
			},
			want: false,
		},
		{
			name: "should return false for 404 error",
			err: &googleapi.Error{
				Code: 404,
			},
			want: false,
		},
		{
			name: "should return false for nil error",
			err:  nil,
			want: false,
		},
		{
			name: "should return false for non-googleapi error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isInsufficientScopeError(tt.err)
			if got != tt.want {
				t.Errorf("isInsufficientScopeError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleCalendarError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		context        string
		wantNil        bool
		wantErrContain string
	}{
		{
			name:    "should return nil for nil error",
			err:     nil,
			context: "test",
			wantNil: true,
		},
		{
			name:           "should suggest login for 401",
			err:            &googleapi.Error{Code: 401},
			context:        "list events",
			wantErrContain: "gsuite login",
		},
		{
			name: "should mention calendar permission for 403 insufficient scope",
			err: &googleapi.Error{
				Code:   403,
				Errors: []googleapi.ErrorItem{{Reason: "insufficientPermissions"}},
			},
			context:        "list events",
			wantErrContain: "calendar permission",
		},
		{
			name:           "should say not found for 404",
			err:            &googleapi.Error{Code: 404},
			context:        "get event",
			wantErrContain: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := HandleCalendarError(tt.err, tt.context)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(got.Error(), tt.wantErrContain) {
				t.Errorf("error %q does not contain %q", got.Error(), tt.wantErrContain)
			}
		})
	}
}
