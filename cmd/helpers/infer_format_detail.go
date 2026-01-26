package helpers

// InferFormatDetail determines suggested format detail based on release type and current format
func InferFormatDetail(releaseType, formatCategory string) string {
	if formatCategory == "" {
		return ""
	}

	// Base suggestions by format category
	switch formatCategory {
	case "CD":
		switch releaseType {
		case "Single":
			return "CD Single"
		case "EP":
			return "CD EP"
		case "Album":
			return "CD Album"
		default:
			return "CD"
		}
	case "Vinyl":
		switch releaseType {
		case "Single":
			return "7\""
		case "EP":
			return "10\""
		case "Album":
			return "LP"
		default:
			return "Vinyl"
		}
	case "Cassette":
		switch releaseType {
		case "Single":
			return "Cassette Single"
		case "Album":
			return "Cassette Album"
		default:
			return "Cassette"
		}
	case "Digital":
		switch releaseType {
		case "Single":
			return "Digital Single"
		case "EP":
			return "Digital EP"
		case "Album":
			return "Digital Album"
		default:
			return "Digital"
		}
	default:
		return ""
	}
}
