package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPromptString(t *testing.T) {
	tests := []struct {
		name       string
		prompt     string
		input      string
		wantResult string
	}{
		{"simple input", "Enter name: ", "Test Name\n", "Test Name"},
		{"empty input", "Enter name: ", "\n", ""},
		{"input with spaces", "Enter text: ", "  spaced text  \n", "spaced text"},
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

			result := PromptString(tt.prompt)

			// Restore stdout and read captured output
			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)

			// Restore stdin
			os.Stdin = oldStdin

			if result != tt.wantResult {
				t.Errorf("PromptString() = %q, want %q", result, tt.wantResult)
			}
		})
	}
}
