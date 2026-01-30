package helpers

import (
	"fmt"
	"strings"

	"musiccat/external/musicbrainz"
)

func SelectReleasesWithPagination(releaseGroups []musicbrainz.ReleaseGroup, pageSize int, formatCategory string, albumOnly, singleOnly bool, year int, titleFilter string, startingPage int) ([]SelectionItem, bool, int, error) {
	currentPage := startingPage
	for {
		start := currentPage * pageSize
		end := start + pageSize
		if end > len(releaseGroups) {
			end = len(releaseGroups)
		}
		if start >= len(releaseGroups) {
			break
		}
		totalPages := (len(releaseGroups) + pageSize - 1) / pageSize
		isLastPage := currentPage+1 >= totalPages
		canFetchMore := len(releaseGroups)%100 == 0 && len(releaseGroups) > 0

		// Show applied filters for clarity
		var filterInfo []string
		if albumOnly {
			filterInfo = append(filterInfo, "albums only")
		}
		if singleOnly {
			filterInfo = append(filterInfo, "singles only")
		}
		if year > 0 {
			filterInfo = append(filterInfo, fmt.Sprintf("year %d", year))
		}
		if titleFilter != "" {
			filterInfo = append(filterInfo, fmt.Sprintf("title '%s'", titleFilter))
		}

		if len(filterInfo) > 0 {
			fmt.Printf("Displaying %d–%d of %d+ releases (filtered: %s):\n", start+1, end, len(releaseGroups), strings.Join(filterInfo, ", "))
		} else {
			fmt.Printf("Displaying %d–%d of %d+ releases:\n", start+1, end, len(releaseGroups))
		}

		for i := start; i < end; i++ {
			rg := releaseGroups[i]
			num := i + 1 // Display absolute number (1-based)
			year := ParseYear(rg.FirstReleaseDate)
			yearStr := ""
			if year != nil {
				yearStr = fmt.Sprintf(" (%d)", *year)
			}
			typeStr := ""
			if rg.PrimaryType != "" {
				typeStr = " [" + rg.PrimaryType
				if len(rg.SecondaryTypes) > 0 {
					typeStr += " + " + strings.Join(rg.SecondaryTypes, ", ")
				}
				typeStr += "]"
			}
			fmt.Printf("%d. %s%s%s\n", num, rg.Title, yearStr, typeStr)
		}

		var prompt string
		if isLastPage && !canFetchMore {
			prompt = "Select releases (numbers, comma-separated, suffix 'p' for promo, suffix 'i' for pirate, suffix '(n)' for quantity): "
		} else {
			prompt = "Select releases (numbers, comma-separated, suffix 'p' for promo, suffix 'i' for pirate, suffix '(n)' for quantity, 00 for more): "
		}
		input := PromptString(prompt)
		if input == "00" {
			if isLastPage && canFetchMore {
				// Signal caller to fetch more results, return current page
				return nil, true, currentPage, nil
			} else if !isLastPage {
				// Just advance to next page of existing results
				currentPage++
				continue
			} else {
				fmt.Println("No more results available.")
				continue
			}
		}
		// Parse selections - can select from any previously displayed items
		selectedItems, err := ParseSelections(input, releaseGroups)
		if err != nil {
			fmt.Println("Invalid input:", err)
			continue
		}
		return selectedItems, false, 0, nil
	}
	return nil, false, 0, fmt.Errorf("no releases selected")
}
