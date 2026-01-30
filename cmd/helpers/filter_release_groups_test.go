package helpers

import (
	"testing"

	"musiccat/external/musicbrainz"
)

func TestFilterReleaseGroups(t *testing.T) {
	// Test data
	album1997 := musicbrainz.ReleaseGroup{
		ID:               "1",
		Title:            "OK Computer",
		FirstReleaseDate: "1997-05-21",
		PrimaryType:      "Album",
	}
	single1997 := musicbrainz.ReleaseGroup{
		ID:               "2",
		Title:            "Paranoid Android",
		FirstReleaseDate: "1997-05-26",
		PrimaryType:      "Single",
	}
	album2000 := musicbrainz.ReleaseGroup{
		ID:               "3",
		Title:            "Kid A",
		FirstReleaseDate: "2000-10-02",
		PrimaryType:      "Album",
	}
	single2003 := musicbrainz.ReleaseGroup{
		ID:               "4",
		Title:            "There There",
		FirstReleaseDate: "2003-05-21",
		PrimaryType:      "Single",
	}

	releaseGroups := []musicbrainz.ReleaseGroup{album1997, single1997, album2000, single2003}

	tests := []struct {
		name        string
		albumOnly   bool
		singleOnly  bool
		afterYear   *int
		beforeYear  *int
		titleFilter string
		wantCount   int
		wantTitles  []string
	}{
		{"no filters", false, false, nil, nil, "", 4, []string{"OK Computer", "Paranoid Android", "Kid A", "There There"}},
		{"album only", true, false, nil, nil, "", 2, []string{"OK Computer", "Kid A"}},
		{"single only", false, true, nil, nil, "", 2, []string{"Paranoid Android", "There There"}},
		{"after 1997", false, false, intPtr(1997), nil, "", 2, []string{"Kid A", "There There"}},
		{"before 2003", false, false, nil, intPtr(2003), "", 3, []string{"OK Computer", "Paranoid Android", "Kid A"}},
		{"title filter", false, false, nil, nil, "kid", 1, []string{"Kid A"}},
		{"title filter case insensitive", false, false, nil, nil, "PARANOID", 1, []string{"Paranoid Android"}},
		{"albums after 1997", true, false, intPtr(1997), nil, "", 1, []string{"Kid A"}},
		{"no matches", true, false, intPtr(2005), nil, "", 0, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterReleaseGroups(releaseGroups, tt.albumOnly, tt.singleOnly, tt.afterYear, tt.beforeYear, tt.titleFilter)
			if len(got) != tt.wantCount {
				t.Errorf("FilterReleaseGroups() returned %d results, want %d", len(got), tt.wantCount)
			}
			for i, rg := range got {
				if i >= len(tt.wantTitles) {
					break
				}
				if rg.Title != tt.wantTitles[i] {
					t.Errorf("Result[%d].Title = %q, want %q", i, rg.Title, tt.wantTitles[i])
				}
			}
		})
	}
}
