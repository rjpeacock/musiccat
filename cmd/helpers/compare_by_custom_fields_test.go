package helpers

import (
	"testing"

	"musiccat/external/musicbrainz"
)

func TestCompareByCustomFields(t *testing.T) {
	album1997 := musicbrainz.ReleaseGroup{
		Title:            "Album A",
		FirstReleaseDate: "1997-05-21",
		PrimaryType:      "Album",
	}
	album2000 := musicbrainz.ReleaseGroup{
		Title:            "Album B",
		FirstReleaseDate: "2000-10-02",
		PrimaryType:      "Album",
	}
	single1997 := musicbrainz.ReleaseGroup{
		Title:            "Single A",
		FirstReleaseDate: "1997-05-26",
		PrimaryType:      "Single",
	}

	tests := []struct {
		name       string
		a          musicbrainz.ReleaseGroup
		b          musicbrainz.ReleaseGroup
		sortFields []string
		desc       bool
		want       bool
	}{
		{
			"sort by type ascending - album before single",
			album1997,
			single1997,
			[]string{"type"},
			false,
			true,
		},
		{
			"sort by type descending - single before album",
			album1997,
			single1997,
			[]string{"type"},
			true,
			false,
		},
		{
			"sort by year ascending",
			album1997,
			album2000,
			[]string{"year"},
			false,
			true,
		},
		{
			"sort by year descending",
			album1997,
			album2000,
			[]string{"year"},
			true,
			false,
		},
		{
			"sort by title ascending",
			album2000, // "Album B"
			album1997, // "Album A"
			[]string{"title"},
			false,
			false, // B > A, so B should not come before A
		},
		{
			"sort by title descending",
			album2000, // "Album B"
			album1997, // "Album A"
			[]string{"title"},
			true,
			true, // B > A, so B should come before A in descending
		},
		{
			"multi-field sort - type then year",
			album2000, // Album 2000
			album1997, // Album 1997
			[]string{"type", "year"},
			false,
			false, // Same type, 2000 > 1997
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareByCustomFields(tt.a, tt.b, tt.sortFields, tt.desc)
			if got != tt.want {
				t.Errorf("CompareByCustomFields(%v, %v, %v, desc=%v) = %v, want %v",
					tt.a.Title, tt.b.Title, tt.sortFields, tt.desc, got, tt.want)
			}
		})
	}
}
