package helpers

import "testing"

func TestInferFormatDetail(t *testing.T) {
	tests := []struct {
		name           string
		releaseType    string
		formatCategory string
		want           string
	}{
		// CD tests
		{"CD Single", "Single", "CD", "Single"},
		{"CD EP", "EP", "CD", "EP"},
		{"CD Album", "Album", "CD", "Album"},
		{"CD other", "Other", "CD", "CD"},
		
		// Vinyl tests
		{"Vinyl Single", "Single", "Vinyl", "7\""},
		{"Vinyl EP", "EP", "Vinyl", "12\""},
		{"Vinyl Album", "Album", "Vinyl", "Album"},
		{"Vinyl other", "Other", "Vinyl", "Vinyl"},
		
		// Cassette tests
		{"Cassette Single", "Single", "Cassette", "Single"},
		{"Cassette Album", "Album", "Cassette", "Album"},
		{"Cassette EP", "EP", "Cassette", "EP"},
		{"Cassette other", "Other", "Cassette", "Cassette"},
		
		// Digital tests
		{"Digital Single", "Single", "Digital", "Single"},
		{"Digital EP", "EP", "Digital", "EP"},
		{"Digital Album", "Album", "Digital", "Album"},
		{"Digital other", "Other", "Digital", "Digital"},
		
		// Edge cases
		{"empty category", "Album", "", ""},
		{"unknown category", "Album", "Unknown", "Album"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferFormatDetail(tt.releaseType, tt.formatCategory)
			if got != tt.want {
				t.Errorf("InferFormatDetail(%q, %q) = %q, want %q",
					tt.releaseType, tt.formatCategory, got, tt.want)
			}
		})
	}
}
