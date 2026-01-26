package helpers

import (
	"testing"

	"musiccat/external/musicbrainz"
)

func TestParseSelections(t *testing.T) {
	releaseGroups := []musicbrainz.ReleaseGroup{
		{ID: "1", Title: "Album 1", PrimaryType: "Album"},
		{ID: "2", Title: "Single 1", PrimaryType: "Single"},
		{ID: "3", Title: "Album 2", PrimaryType: "Album"},
	}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
		checks    func(t *testing.T, items []SelectionItem)
	}{
		{
			"single selection",
			"1",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if items[0].ReleaseGroup.ID != "1" {
					t.Errorf("Expected ID 1, got %s", items[0].ReleaseGroup.ID)
				}
				if items[0].Promo || items[0].Pirate {
					t.Error("Expected non-promo, non-pirate")
				}
				if items[0].Quantity != 1 {
					t.Errorf("Expected quantity 1, got %d", items[0].Quantity)
				}
			},
		},
		{
			"multiple selections",
			"1,2,3",
			3,
			false,
			func(t *testing.T, items []SelectionItem) {
				if len(items) != 3 {
					t.Errorf("Expected 3 items, got %d", len(items))
				}
			},
		},
		{
			"promo suffix",
			"1p",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if !items[0].Promo {
					t.Error("Expected promo to be true")
				}
			},
		},
		{
			"pirate suffix",
			"2i",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if !items[0].Pirate {
					t.Error("Expected pirate to be true")
				}
			},
		},
		{
			"both promo and pirate",
			"1ip",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if !items[0].Promo || !items[0].Pirate {
					t.Error("Expected both promo and pirate to be true")
				}
			},
		},
		{
			"quantity in parentheses",
			"1(3)",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if items[0].Quantity != 3 {
					t.Errorf("Expected quantity 3, got %d", items[0].Quantity)
				}
			},
		},
		{
			"quantity with promo",
			"2(2)p",
			1,
			false,
			func(t *testing.T, items []SelectionItem) {
				if items[0].Quantity != 2 || !items[0].Promo {
					t.Errorf("Expected quantity 2 and promo, got quantity=%d promo=%v",
						items[0].Quantity, items[0].Promo)
				}
			},
		},
		{
			"invalid selection out of range",
			"5",
			0,
			true,
			nil,
		},
		{
			"invalid selection zero",
			"0",
			0,
			true,
			nil,
		},
		{
			"invalid quantity format",
			"1(abc)",
			0,
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSelections(tt.input, releaseGroups)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSelections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.wantCount {
					t.Errorf("ParseSelections() returned %d items, want %d", len(got), tt.wantCount)
				}
				if tt.checks != nil {
					tt.checks(t, got)
				}
			}
		})
	}
}
