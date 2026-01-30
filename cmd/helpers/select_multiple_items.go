package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func SelectMultipleItems[T any](prompt string, items []T) ([]T, error) {
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
