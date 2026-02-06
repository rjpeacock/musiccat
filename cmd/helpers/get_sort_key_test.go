package helpers

import (
	"testing"
)

func TestGetSortKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Beatles", "Beatles"},
		{"The Rolling Stones", "Rolling Stones"},
		{"These Animal Men", "These Animal Men"},
		{"Beatles", "Beatles"},
		{"the beatles", "beatles"},
		{"THE BEATLES", "BEATLES"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getSortKey(tt.input)
			if result != tt.expected {
				t.Errorf("getSortKey(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
