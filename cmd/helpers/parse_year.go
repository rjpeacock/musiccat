package helpers

import "strconv"

func ParseYear(date string) *int {
	if len(date) < 4 {
		return nil
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return nil
	}
	return &year
}
