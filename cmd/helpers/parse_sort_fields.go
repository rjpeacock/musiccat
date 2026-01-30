package helpers

import "strings"

// ParseSortFields parses sort field string into slice of fields
func ParseSortFields(sortStr string) []string {
	if sortStr == "" {
		return nil
	}

	fields := strings.Split(sortStr, ",")
	validFields := map[string]bool{"type": true, "year": true, "title": true}

	var result []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if validFields[field] {
			result = append(result, field)
		}
	}

	return result
}
