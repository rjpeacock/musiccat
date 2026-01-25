package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"musiccat/external/musicbrainz"
)

func parseYear(date string) *int {
	if len(date) < 4 {
		return nil
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return nil
	}
	return &year
}

func promptString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptOptionalInt(prompt string) *int {
	input := promptString(prompt)
	if input == "" {
		return nil
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptValidFormat(prompt string) string {
	for {
		input := promptString(prompt)
		for _, f := range ValidFormats {
			if strings.EqualFold(input, f) {
				return f
			}
		}
		fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	}
}

func promptFormatDetail(formatCategory string) string {
	suggestions, exists := formatDetailSuggestions[formatCategory]
	if !exists {
		return promptString("Format detail (optional): ")
	}

	suggestionStr := strings.Join(suggestions, ", ")
	fmt.Printf("Suggested format details: %s\n", suggestionStr)
	return promptString("Format detail (optional): ")
}

func selectItem[T any](prompt string, items []T) (T, error) {
	var zero T
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(items) {
			fmt.Printf("Invalid selection. Enter a number between 1 and %d: ", len(items))
			continue
		}
		return items[num-1], nil
	}
	return zero, scanner.Err()
}

func selectMultipleItems[T any](prompt string, items []T) ([]T, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		parts := strings.Split(input, ",")
		var selected []T
		valid := true
		for _, part := range parts {
			part = strings.TrimSpace(part)
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 || num > len(items) {
				valid = false
				break
			}
			selected = append(selected, items[num-1])
		}
		if !valid {
			fmt.Printf("Invalid selection. Enter numbers between 1 and %d, comma-separated: ", len(items))
			continue
		}
		return selected, nil
	}
	return nil, scanner.Err()
}

type SelectionItem struct {
	releaseGroup musicbrainz.ReleaseGroup
	promo        bool
	pirate       bool
	quantity     int
	notes        string
}

func selectReleasesWithPagination(releaseGroups []musicbrainz.ReleaseGroup, pageSize int, formatCategory string, albumOnly, singleOnly bool, year int, titleFilter string) ([]SelectionItem, error) {
	const moreNumber = 99
	currentPage := 0
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
			fmt.Printf("Displaying %d–%d of %d releases (filtered: %s):\n", start+1, end, len(releaseGroups), strings.Join(filterInfo, ", "))
		} else {
			fmt.Printf("Displaying %d–%d of %d releases:\n", start+1, end, len(releaseGroups))
		}

		for i := start; i < end; i++ {
			rg := releaseGroups[i]
			num := i - start + 1 // Display number relative to current page (1-based)
			year := parseYear(rg.FirstReleaseDate)
			yearStr := ""
			if year != nil {
				yearStr = fmt.Sprintf(" (%d)", *year)
			}
			typeStr := ""
			if rg.PrimaryType != "" {
				typeStr = " [" + rg.PrimaryType + "]"
			}
			fmt.Printf("%d. %s%s%s\n", num, rg.Title, yearStr, typeStr)
		}

		var prompt string
		if isLastPage {
			prompt = "Select releases (numbers, comma-separated, suffix 'p' for promo, suffix '(n)' for quantity): "
		} else {
			prompt = fmt.Sprintf("Select releases (numbers, comma-separated, suffix 'p' for promo, suffix '(n)' for quantity, %d for more): ", moreNumber)
		}
		input := promptString(prompt)
		if input == strconv.Itoa(moreNumber) {
			if isLastPage {
				fmt.Println("Already on last page.")
				continue
			}
			currentPage++
			continue
		}
		// Parse selections with page offset
		selectedItems, err := parseSelections(input, releaseGroups, start)
		if err != nil {
			fmt.Println("Invalid input:", err)
			continue
		}
		return selectedItems, nil
	}
	return nil, fmt.Errorf("no releases selected")
}

func parseSelections(input string, releaseGroups []musicbrainz.ReleaseGroup, pageOffset int) ([]SelectionItem, error) {
	parts := strings.Split(input, ",")
	var selected []SelectionItem

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for promo suffix 'p'
		promo := strings.HasSuffix(part, "p")
		if promo {
			part = strings.TrimSuffix(part, "p")
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
			return nil, fmt.Errorf("invalid selection %s", part)
		}

		// Calculate actual array index with page offset
		actualIndex := pageOffset + (num - 1)
		if actualIndex >= len(releaseGroups) {
			return nil, fmt.Errorf("invalid selection %s", part)
		}

		selected = append(selected, SelectionItem{releaseGroup: releaseGroups[actualIndex], promo: promo, pirate: false, quantity: quantity})
	}

	// Ask for variant notes for each selected release
	for i, item := range selected {
		if item.quantity > 1 {
			notes := promptString(fmt.Sprintf("Variant notes for %s (optional): ", item.releaseGroup.Title))
			selected[i].notes = notes
		}
	}

	return selected, nil
}

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
			return compareByCustomFields(sorted[i], sorted[j], sortFields, desc)
		}

		// Default sorting: type → year → title
		return compareByDefaultOrder(sorted[i], sorted[j], desc)
	})

	return sorted
}

// compareByDefaultOrder implements the default sorting logic
func compareByDefaultOrder(a, b musicbrainz.ReleaseGroup, desc bool) bool {
	// Primary: release type
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

	// Secondary: year
	aYear := parseYear(a.FirstReleaseDate)
	bYear := parseYear(b.FirstReleaseDate)

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

// compareByCustomFields implements custom sorting based on specified fields
func compareByCustomFields(a, b musicbrainz.ReleaseGroup, sortFields []string, desc bool) bool {
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
			aYear := parseYear(a.FirstReleaseDate)
			bYear := parseYear(b.FirstReleaseDate)

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

// parseSortFields parses sort field string into slice of fields
func parseSortFields(sortStr string) []string {
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
		year := parseYear(rg.FirstReleaseDate)
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

// inferFormatDetail determines suggested format detail based on release type and current format
func inferFormatDetail(releaseType, formatCategory string) string {
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

func promptOptionalString(prompt, current string) *string {
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	return &input
}

func promptOptionalIntUpdate(prompt string, current *int) *int {
	currentStr := ""
	if current != nil {
		currentStr = strconv.Itoa(*current)
	}
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptValidFormatUpdate(prompt, current string) *string {
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", current))
	if input == "" {
		return nil
	}
	for _, f := range ValidFormats {
		if strings.EqualFold(input, f) {
			return &f
		}
	}
	fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	return nil
}

func promptOptionalFloat(prompt string, current float64) *float64 {
	input := promptString(prompt + fmt.Sprintf(" (current: %.2f): ", current))
	if input == "" {
		return nil
	}
	num, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptOptionalBool(prompt string, current bool) *bool {
	currentStr := "no"
	if current {
		currentStr = "yes"
	}
	input := promptString(prompt + fmt.Sprintf(" (current: %s): ", currentStr))
	if input == "" {
		return nil
	}
	value := strings.EqualFold(input, "yes") || strings.EqualFold(input, "y") || strings.EqualFold(input, "true")
	return &value
}
