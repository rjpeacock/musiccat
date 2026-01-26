package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func SelectItem[T any](prompt string, items []T) (T, error) {
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
