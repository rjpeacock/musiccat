package helpers

import (
	"reflect"
	"testing"
)

func TestParseSortFields(t *testing.T) {
	tests := []struct {
		name    string
		sortStr string
		want    []string
	}{
		{"empty string", "", nil},
		{"single field", "type", []string{"type"}},
		{"multiple fields", "type,year,title", []string{"type", "year", "title"}},
		{"with spaces", "type, year, title", []string{"type", "year", "title"}},
		{"invalid field ignored", "type,invalid,year", []string{"type", "year"}},
		{"all invalid", "invalid,bad", nil},
		{"duplicate fields", "type,year,type", []string{"type", "year", "type"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSortFields(tt.sortStr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSortFields(%q) = %v, want %v", tt.sortStr, got, tt.want)
			}
		})
	}
}
