package tags

import "testing"

func TestCanonicalizeTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "promo",
			expected: "promo",
		},
		{
			name:     "uppercase to lowercase",
			input:    "PROMO",
			expected: "promo",
		},
		{
			name:     "mixed case",
			input:    "Promo",
			expected: "promo",
		},
		{
			name:     "whitespace to hyphen",
			input:    "Missing Sleeve",
			expected: "missing-sleeve",
		},
		{
			name:     "multiple words",
			input:    "sleeve missing",
			expected: "sleeve-missing",
		},
		{
			name:     "leading/trailing whitespace",
			input:    "  promo  ",
			expected: "promo",
		},
		{
			name:     "multiple spaces",
			input:    "missing  sleeve",
			expected: "missing-sleeve",
		},
		{
			name:     "tabs to hyphens",
			input:    "missing\tsleeve",
			expected: "missing-sleeve",
		},
		{
			name:     "repeated hyphens",
			input:    "missing--sleeve",
			expected: "missing-sleeve",
		},
		{
			name:     "complex example",
			input:    "  Missing   Sleeve  ",
			expected: "missing-sleeve",
		},
		{
			name:     "already canonical",
			input:    "missing-sleeve",
			expected: "missing-sleeve",
		},
		{
			name:     "single word with spaces",
			input:    "  signed  ",
			expected: "signed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanonicalizeTag(tt.input)
			if result != tt.expected {
				t.Errorf("CanonicalizeTag(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
