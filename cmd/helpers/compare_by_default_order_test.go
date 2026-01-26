package helpers

import (
	"testing"

	"musiccat/external/musicbrainz"
)

func TestCompareByDefaultOrder(t *testing.T) {
	album1997 := musicbrainz.ReleaseGroup{
		Title:            "OK Computer",
		FirstReleaseDate: "1997-05-21",
		PrimaryType:      "Album",
	}
	album2000 := musicbrainz.ReleaseGroup{
		Title:            "Kid A",
		FirstReleaseDate: "2000-10-02",
		PrimaryType:      "Album",
	}
	single1997 := musicbrainz.ReleaseGroup{
		Title:            "Paranoid Android",
		FirstReleaseDate: "1997-05-26",
		PrimaryType:      "Single",
	}
	ep1998 := musicbrainz.ReleaseGroup{
		Title:            "Airbag",
		FirstReleaseDate: "1998-01-01",
		PrimaryType:      "EP",
	}
	other := musicbrainz.ReleaseGroup{
		Title:            "Other Release",
		FirstReleaseDate: "1999-01-01",
		PrimaryType:      "Other",
	}

	tests := []struct {
		name string
		a    musicbrainz.ReleaseGroup
		b    musicbrainz.ReleaseGroup
		desc bool
		want bool
	}{
		// Type ordering (Album < EP < Single < Other)
		{"album before single asc", album1997, single1997, false, true},
		{"album before single desc", album1997, single1997, true, false},
		{"album before ep asc", album1997, ep1998, false, true},
		{"ep before single asc", ep1998, single1997, false, true},
		{"single before other asc", single1997, other, false, true},
		
		// Year ordering (same type)
		{"same type older first asc", album1997, album2000, false, true},
		{"same type older first desc", album1997, album2000, true, false},
		
		// Title ordering (same type and year)
		{"same type/year title asc", album1997, album1997, false, false}, // Same title
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareByDefaultOrder(tt.a, tt.b, tt.desc)
			if got != tt.want {
				t.Errorf("CompareByDefaultOrder(%v, %v, desc=%v) = %v, want %v",
					tt.a.Title, tt.b.Title, tt.desc, got, tt.want)
			}
		})
	}
}
