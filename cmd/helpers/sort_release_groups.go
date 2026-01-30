package helpers

import (
	"sort"

	"musiccat/external/musicbrainz"
)

// SortReleaseGroups sorts release groups according to the default specification:
// Primary: release type (Album → EP → Single → Other)
// Secondary: first release year (ascending)
// Tertiary: title (alphabetical)
func SortReleaseGroups(releaseGroups []musicbrainz.ReleaseGroup, sortFields []string, desc bool) []musicbrainz.ReleaseGroup {
	sorted := make([]musicbrainz.ReleaseGroup, len(releaseGroups))
	copy(sorted, releaseGroups)

	sort.Slice(sorted, func(i, j int) bool {
		// If custom sort fields are specified, use them
		if len(sortFields) > 0 {
			return CompareByCustomFields(sorted[i], sorted[j], sortFields, desc)
		}

		// Default sorting: type → year → title
		return CompareByDefaultOrder(sorted[i], sorted[j], desc)
	})

	return sorted
}
