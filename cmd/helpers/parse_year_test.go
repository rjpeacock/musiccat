package helpers

import "testing"

func TestParseYear(t *testing.T) {
	tests := []struct {
		name  string
		date  string
		want  *int
	}{
		{"valid full date", "1997-05-13", intPtr(1997)},
		{"valid year only", "2003", intPtr(2003)},
		{"valid long date", "1995-08-28T00:00:00", intPtr(1995)},
		{"empty string", "", nil},
		{"too short", "199", nil},
		{"invalid year", "abcd", nil},
		{"partial year", "20", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseYear(tt.date)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParseYear(%q) = %v, want %v", tt.date, got, tt.want)
				return
			}
			if got != nil && tt.want != nil && *got != *tt.want {
				t.Errorf("ParseYear(%q) = %d, want %d", tt.date, *got, *tt.want)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
