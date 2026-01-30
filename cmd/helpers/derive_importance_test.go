package helpers

import "testing"

func TestDeriveImportance(t *testing.T) {
	tests := []struct {
		name         string
		isPirate     bool
		isPromo      bool
		formatDetail *string
		want         int
	}{
		{"pirate always 1", true, false, strPtr("Album"), 1},
		{"pirate promo also 1", true, true, strPtr("Album"), 1},
		{"album non-promo", false, false, strPtr("Album"), 5},
		{"album promo", false, true, strPtr("Album"), 3},
		{"single non-promo", false, false, strPtr("Single"), 4},
		{"single promo", false, true, strPtr("Single"), 2},
		{"other non-promo", false, false, strPtr("EP"), 4},
		{"other promo", false, true, strPtr("EP"), 2},
		{"nil format non-promo", false, false, nil, 4},
		{"nil format promo", false, true, nil, 2},
		{"empty string non-promo", false, false, strPtr(""), 4},
		{"promo minimum cap", false, true, strPtr(""), 2}, // 4-2=2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveImportance(tt.isPirate, tt.isPromo, tt.formatDetail)
			if got != tt.want {
				t.Errorf("DeriveImportance(pirate=%v, promo=%v, detail=%v) = %d, want %d",
					tt.isPirate, tt.isPromo, formatDetailStr(tt.formatDetail), got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func formatDetailStr(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
