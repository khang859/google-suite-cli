package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestOutputJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantErr   bool
		contained string
	}{
		{
			name: "should produce JSON for valid struct",
			input: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{Name: "test", Count: 42},
			wantErr:   false,
			contained: `"name": "test"`,
		},
		{
			name:    "should error for unmarshalable input",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := outputJSON(tt.input)

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout

			if tt.wantErr {
				if err == nil {
					t.Error("outputJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("outputJSON() unexpected error: %v", err)
			}

			output := buf.String()
			if tt.contained != "" && !strings.Contains(output, tt.contained) {
				t.Errorf("outputJSON() output = %q, want it to contain %q", output, tt.contained)
			}
		})
	}
}
