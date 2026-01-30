package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPromptOptionalBool(t *testing.T) {
	tests := []struct {
		name       string
		current    bool
		input      string
		wantResult *bool
		wantNil    bool
	}{
		{"yes input", false, "y\n", boolPtr(true), false},
		{"no input", true, "n\n", boolPtr(false), false},
		{"empty returns nil", true, "\n", nil, true},
		{"empty returns nil from false", false, "\n", nil, true},
		{"yes uppercase", false, "Y\n", boolPtr(true), false},
		{"no uppercase", true, "N\n", boolPtr(false), false},
		{"true keyword", false, "true\n", boolPtr(true), false},
		{"anything else is false", true, "maybe\n", boolPtr(false), false},
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

			result := PromptOptionalBool("Promo", tt.current)

			// Restore stdout
			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)

			// Restore stdin
			os.Stdin = oldStdin

			if tt.wantNil {
				if result != nil {
					t.Errorf("PromptOptionalBool() = %v, want nil", *result)
				}
			} else {
				if result == nil {
					t.Errorf("PromptOptionalBool() = nil, want %v", *tt.wantResult)
				} else if *result != *tt.wantResult {
					t.Errorf("PromptOptionalBool() = %v, want %v", *result, *tt.wantResult)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
