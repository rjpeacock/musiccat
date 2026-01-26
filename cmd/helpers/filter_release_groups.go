package helpers

import (
	"strings"

	"musiccat/external/musicbrainz"
)

// FilterReleaseGroups applies filters to release groups
func FilterReleaseGroups(releaseGroups []musicbrainz.ReleaseGroup, albumOnly, singleOnly bool, afterYear, beforeYear *int, titleFilter string) []musicbrainz.ReleaseGroup {
	var filtered []musicbrainz.ReleaseGroup

	for _, rg := range releaseGroups {
		// Album/Single only filters
		if albumOnly && rg.PrimaryType != "Album" {
			continue
		}
		if singleOnly && rg.PrimaryType != "Single" {
			continue
		}

		// Year filters
		year := ParseYear(rg.FirstReleaseDate)
		if year != nil {
			if afterYear != nil && *year <= *afterYear {
				continue
			}
			if beforeYear != nil && *year >= *beforeYear {
				continue
			}
		}

		// Title filter (partial match, case-insensitive)
		if titleFilter != "" {
			if !strings.Contains(strings.ToLower(rg.Title), strings.ToLower(titleFilter)) {
				continue
			}
		}

		filtered = append(filtered, rg)
	}

	return filtered
}
