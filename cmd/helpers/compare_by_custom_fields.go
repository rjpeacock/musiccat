package helpers

import "musiccat/external/musicbrainz"

// CompareByCustomFields implements custom sorting based on specified fields
func CompareByCustomFields(a, b musicbrainz.ReleaseGroup, sortFields []string, desc bool) bool {
	for _, field := range sortFields {
		var cmp int
		switch field {
		case "type":
			typeOrder := map[string]int{
				"Album":  1,
				"EP":     2,
				"Single": 3,
			}

			aTypeOrder := 4 // "Other"
			bTypeOrder := 4 // "Other"

			if order, exists := typeOrder[a.PrimaryType]; exists {
				aTypeOrder = order
			}
			if order, exists := typeOrder[b.PrimaryType]; exists {
				bTypeOrder = order
			}

			cmp = aTypeOrder - bTypeOrder

		case "year":
			aYear := ParseYear(a.FirstReleaseDate)
			bYear := ParseYear(b.FirstReleaseDate)

			if aYear != nil && bYear != nil {
				cmp = *aYear - *bYear
			} else if aYear != nil && bYear == nil {
				cmp = -1
			} else if aYear == nil && bYear != nil {
				cmp = 1
			}

		case "title":
			if a.Title < b.Title {
				cmp = -1
			} else if a.Title > b.Title {
				cmp = 1
			} else {
				cmp = 0
			}
		}

		if cmp != 0 {
			if desc {
				return cmp > 0
			}
			return cmp < 0
		}
	}

	// If all fields are equal, maintain original order
	return false
}
