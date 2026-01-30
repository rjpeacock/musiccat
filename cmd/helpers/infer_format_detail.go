package helpers

// InferFormatDetail determines suggested format detail based on release type and current format
func InferFormatDetail(releaseType, formatCategory string) string {
	if formatCategory == "" {
		return ""
	}

	// Vinyl has special size-based format details
	if formatCategory == "Vinyl" {
		switch releaseType {
		case "Single":
			return "7\""
		case "EP":
			return "12\""
		case "Album":
			return "Album"
		default:
			return "Vinyl"
		}
	}

	// CD, Digital, Cassette all use release type as format detail
	switch releaseType {
	case "Single", "EP", "Album":
		return releaseType
	default:
		return formatCategory
	}
}
