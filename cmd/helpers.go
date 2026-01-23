package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	index int
	promo bool
}

func selectReleasesWithPagination(releaseGroups []ReleaseGroup, pageSize int) ([]SelectionItem, error) {
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
		fmt.Printf("Page %d of %d:\n", currentPage+1, (len(releaseGroups)+pageSize-1)/pageSize)
		for i := start; i < end; i++ {
			rg := releaseGroups[i]
			num := i + 1
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
		prompt := fmt.Sprintf("Select releases (numbers, comma-separated, suffix 'p' for promo, %d for more): ", moreNumber)
		input := promptString(prompt)
		if input == strconv.Itoa(moreNumber) {
			currentPage++
			continue
		}
		// Parse selections
		selectedItems, err := parseSelections(input, releaseGroups)
		if err != nil {
			fmt.Println("Invalid input:", err)
			continue
		}
		return selectedItems, nil
	}
	return nil, fmt.Errorf("no releases selected")
}

func parseSelections(input string, releaseGroups []ReleaseGroup) ([]SelectionItem, error) {
	parts := strings.Split(input, ",")
	var selected []SelectionItem
	for _, part := range parts {
		part = strings.TrimSpace(part)
		promo := false
		if strings.HasSuffix(part, "p") {
			promo = true
			part = part[:len(part)-1]
		}
		num, err := strconv.Atoi(part)
		if err != nil || num < 1 || num > len(releaseGroups) {
			return nil, fmt.Errorf("invalid selection %s", part)
		}
		selected = append(selected, SelectionItem{index: num, promo: promo})
	}
	return selected, nil
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
