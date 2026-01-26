package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPromptOptionalString(t *testing.T) {
	tests := []struct {
		name       string
		prompt     string
		current    string
		input      string
		wantResult *string
		wantNil    bool
	}{
		{
			"empty input returns nil",
			"Notes",
			"Current note",
			"\n",
			nil,
			true,
		},
		{
			"update value",
			"Notes",
			"Old note",
			"New note\n",
			stringPtr("New note"),
			false,
		},
		{
			"set new value from empty",
			"Notes",
			"",
			"First note\n",
			stringPtr("First note"),
			false,
		},
		{
			"any text sets value",
			"Notes",
			"Old",
			"Updated\n",
			stringPtr("Updated"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write test input
			go func() {
				w.Write([]byte(tt.input))
				w.Close()
			}()

			// Capture stdout
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			result := PromptOptionalString(tt.prompt, tt.current)

			// Restore stdout and read captured output
			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)

			// Restore stdin
			os.Stdin = oldStdin

			if tt.wantNil {
				if result != nil {
					t.Errorf("PromptOptionalString() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("PromptOptionalString() = nil, want %q", *tt.wantResult)
				} else if *result != *tt.wantResult {
					t.Errorf("PromptOptionalString() = %q, want %q", *result, *tt.wantResult)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
