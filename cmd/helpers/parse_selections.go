package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"musiccat/external/musicbrainz"
)

func ParseSelections(input string, releaseGroups []musicbrainz.ReleaseGroup) ([]SelectionItem, error) {
	parts := strings.Split(input, ",")
	var selected []SelectionItem

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for promo suffix 'p'
		promo := strings.HasSuffix(part, "p")
		if promo {
			part = strings.TrimSuffix(part, "p")
		}

		// Check for pirate suffix 'i'
		pirate := strings.HasSuffix(part, "i")
		if pirate {
			part = strings.TrimSuffix(part, "i")
		}

		// Handle quantity with optional parentheses: "1(2)" or just "1"
		quantity := 1
		if strings.Contains(part, "(") && strings.HasSuffix(part, ")") {
			parenStart := strings.Index(part, "(")
			parenEnd := strings.Index(part, ")")
			quantityStr := part[parenStart+1 : parenEnd]
			q, err := strconv.Atoi(quantityStr)
			if err != nil {
				return nil, fmt.Errorf("invalid quantity in selection %s", part)
			}
			quantity = q
			part = part[:parenStart]
		}

		num, err := strconv.Atoi(part)
		if err != nil || num < 1 || num > len(releaseGroups) {
			return nil, fmt.Errorf("invalid selection %s (valid range: 1-%d)", part, len(releaseGroups))
		}

		// User enters absolute numbers (1-based), convert to 0-based index
		actualIndex := num - 1

		selected = append(selected, SelectionItem{ReleaseGroup: releaseGroups[actualIndex], Promo: promo, Pirate: pirate, Quantity: quantity})
	}

	return selected, nil
}
