package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPromptOptionalFloat(t *testing.T) {
	tests := []struct {
		name       string
		current    float64
		input      string
		wantResult *float64
		wantNil    bool
	}{
		{
			"empty returns nil",
			19.99,
			"\n",
			nil,
			true,
		},
		{
			"update value",
			10.00,
			"25.50\n",
			floatPtr(25.50),
			false,
		},
		{
			"new value",
			0.00,
			"12.99\n",
			floatPtr(12.99),
			false,
		},
		{
			"invalid input returns nil",
			20.00,
			"invalid\n",
			nil,
			true,
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

			// Capture stdout to suppress prompts
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut

			result := PromptOptionalFloat("Cost", tt.current)

			// Restore stdout
			wOut.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, rOut)

			// Restore stdin
			os.Stdin = oldStdin

			if tt.wantNil {
				if result != nil {
					t.Errorf("PromptOptionalFloat() = %v, want nil", *result)
				}
			} else {
				if result == nil {
					t.Errorf("PromptOptionalFloat() = nil, want %v", *tt.wantResult)
				} else if *result != *tt.wantResult {
					t.Errorf("PromptOptionalFloat() = %v, want %v", *result, *tt.wantResult)
				}
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
