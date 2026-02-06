package cmd

import (
	"testing"
)

func TestFormatDetailOverride(t *testing.T) {
	tests := []struct {
		name           string
		formatCategory string
		releaseType    string
		configDetail   string // What's in config.CurrentFormatDetail
		expected       string
	}{
		{
			name:           "CD Album with 7\" config - conflict, musicbrainz wins",
			formatCategory: "CD",
			releaseType:    "Album",
			configDetail:   "7\"",
			expected:       "Album", // MusicBrainz wins
		},
		{
			name:           "CD Single with Single config - compatible, config wins",
			formatCategory: "CD",
			releaseType:    "Single",
			configDetail:   "Single",
			expected:       "Single", // Config wins (compatible)
		},
		{
			name:           "Vinyl Album with 7\" config - conflict, musicbrainz wins",
			formatCategory: "Vinyl",
			releaseType:    "Album",
			configDetail:   "7\"",
			expected:       "Album", // MusicBrainz wins
		},
		{
			name:           "Vinyl Single with 7\" config - compatible, config wins",
			formatCategory: "Vinyl",
			releaseType:    "Single",
			configDetail:   "7\"",
			expected:       "7\"", // Config wins (compatible)
		},
		{
			name:           "CD Album with Album config - compatible",
			formatCategory: "CD",
			releaseType:    "Album",
			configDetail:   "Album",
			expected:       "Album",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily mock config, so we'll test the logic directly
			inferred := inferFormatDetailInternal(tt.formatCategory, tt.releaseType)
			compatible := isCompatible(tt.formatCategory, tt.configDetail, tt.releaseType)

			var result string
			if inferred != "" && !compatible {
				result = inferred
			} else {
				result = tt.configDetail
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q (inferred: %q, compatible: %v)",
					tt.expected, result, inferred, compatible)
			}
		})
	}
}
