package helpers

import (
	"testing"

	"musiccat/external/musicbrainz"
)

func TestSortReleaseGroups(t *testing.T) {
	album1997 := musicbrainz.ReleaseGroup{
		ID:               "1",
		Title:            "OK Computer",
		FirstReleaseDate: "1997-05-21",
		PrimaryType:      "Album",
	}
	album2000 := musicbrainz.ReleaseGroup{
		ID:               "2",
		Title:            "Kid A",
		FirstReleaseDate: "2000-10-02",
		PrimaryType:      "Album",
	}
	single1997 := musicbrainz.ReleaseGroup{
		ID:               "3",
		Title:            "Paranoid Android",
		FirstReleaseDate: "1997-05-26",
		PrimaryType:      "Single",
	}
	ep1998 := musicbrainz.ReleaseGroup{
		ID:               "4",
		Title:            "Airbag",
		FirstReleaseDate: "1998-01-01",
		PrimaryType:      "EP",
	}

	tests := []struct {
		name       string
		input      []musicbrainz.ReleaseGroup
		sortFields []string
		desc       bool
		wantOrder  []string // Expected IDs in order
	}{
		{
			"default sort ascending",
			[]musicbrainz.ReleaseGroup{single1997, album2000, ep1998, album1997},
			nil,
			false,
			[]string{"1", "2", "4", "3"}, // Albums first (by year), then EP, then Single
		},
		{
			"default sort descending",
			[]musicbrainz.ReleaseGroup{single1997, album2000, ep1998, album1997},
			nil,
			true,
			[]string{"3", "4", "2", "1"}, // Reverse order
		},
		{
			"custom sort by year",
			[]musicbrainz.ReleaseGroup{album2000, single1997, ep1998, album1997},
			[]string{"year"},
			false,
			[]string{"3", "1", "4", "2"}, // 1997 (single), 1997 (album), 1998, 2000 - maintains input order for ties
		},
		{
			"custom sort by title",
			[]musicbrainz.ReleaseGroup{album2000, single1997, ep1998, album1997},
			[]string{"title"},
			false,
			[]string{"4", "2", "1", "3"}, // Airbag, Kid A, OK Computer, Paranoid Android
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortReleaseGroups(tt.input, tt.sortFields, tt.desc)
			if len(got) != len(tt.wantOrder) {
				t.Errorf("SortReleaseGroups() returned %d items, want %d", len(got), len(tt.wantOrder))
				return
			}
			for i, wantID := range tt.wantOrder {
				if got[i].ID != wantID {
					t.Errorf("Position %d: got ID %s (%s), want ID %s",
						i, got[i].ID, got[i].Title, wantID)
				}
			}
		})
	}
}
