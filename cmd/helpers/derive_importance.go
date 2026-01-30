package helpers

// DeriveImportance derives importance score based on pirate status, promo status, and format detail
// Returns: 1 for pirate, else 5 for albums, 4 for singles/other, minus 2 if promo
func DeriveImportance(isPirate, isPromo bool, formatDetail *string) int {
	if isPirate {
		return 1
	}

	importance := 4 // Default for singles/other

	if formatDetail != nil && *formatDetail == "Album" {
		importance = 5
	}

	if isPromo {
		importance -= 2
	}

	// Ensure minimum of 1
	if importance < 1 {
		importance = 1
	}

	return importance
}
