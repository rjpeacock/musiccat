package helpers

import "musiccat/external/musicbrainz"

// CompareByDefaultOrder implements the default sorting logic
func CompareByDefaultOrder(a, b musicbrainz.ReleaseGroup, desc bool) bool {
	// Primary: release type (with pure albums before albums with secondary types)
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

	if aTypeOrder != bTypeOrder {
		if desc {
			return aTypeOrder > bTypeOrder
		}
		return aTypeOrder < bTypeOrder
	}

	// If both are the same type, prioritize pure types over secondary-typed releases
	// (e.g., pure Album before Album+Compilation)
	if aTypeOrder == bTypeOrder {
		aHasSecondary := len(a.SecondaryTypes) > 0
		bHasSecondary := len(b.SecondaryTypes) > 0
		
		if aHasSecondary != bHasSecondary {
			if desc {
				return aHasSecondary // With secondary comes first when desc
			}
			return !aHasSecondary // Pure comes first when ascending
		}
	}

	// Secondary: year
	aYear := ParseYear(a.FirstReleaseDate)
	bYear := ParseYear(b.FirstReleaseDate)

	if aYear != nil && bYear != nil && *aYear != *bYear {
		if desc {
			return *aYear > *bYear
		}
		return *aYear < *bYear
	} else if aYear != nil && bYear == nil {
		return !desc // Years come before null years when ascending
	} else if aYear == nil && bYear != nil {
		return desc // Null years come after years when ascending
	}

	// Tertiary: title (alphabetical)
	if desc {
		return a.Title > b.Title
	}
	return a.Title < b.Title
}
