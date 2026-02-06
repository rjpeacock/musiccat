package helpers

import (
	"musiccat/external/musicbrainz"
	"testing"
)

func TestSortingFixes(t *testing.T) {
	tests := []struct {
		name     string
		input    []musicbrainz.ReleaseGroup
		expected []int // Indices in the order they should appear
	}{
		{
			name: "Case-insensitive sorting",
			input: []musicbrainz.ReleaseGroup{
				{Title: "apple", PrimaryType: "Single"},
				{Title: "Banana", PrimaryType: "Single"},
				{Title: "cherry", PrimaryType: "Single"},
			},
			expected: []int{0, 1, 2}, // apple, Banana, cherry
		},
		{
			name: "The prefix handling",
			input: []musicbrainz.ReleaseGroup{
				{Title: "The Beatles", PrimaryType: "Album"},
				{Title: "Beatles", PrimaryType: "Album"},
				{Title: "The Rolling Stones", PrimaryType: "Album"},
			},
			expected: []int{0, 1, 2}, // The Beatles, Beatles, The Rolling Stones (stable sort)
		},
		{
			name: "The vs These - only remove The prefix",
			input: []musicbrainz.ReleaseGroup{
				{Title: "These Animal Men", PrimaryType: "Album"},
				{Title: "The Beatles", PrimaryType: "Album"},
			},
			expected: []int{1, 0}, // The Beatles (as Beatles), These Animal Men
		},
		{
			name: "Empty years at end",
			input: []musicbrainz.ReleaseGroup{
				{Title: "Album A", PrimaryType: "Album", FirstReleaseDate: "2023-01-01"},
				{Title: "Album B", PrimaryType: "Album", FirstReleaseDate: ""},
				{Title: "Album C", PrimaryType: "Album", FirstReleaseDate: "2021-01-01"},
				{Title: "Album D", PrimaryType: "Album", FirstReleaseDate: ""},
			},
			expected: []int{2, 0, 1, 3}, // 2021, 2023, (empty), (empty)
		},
		{
			name: "Mixed case and The",
			input: []musicbrainz.ReleaseGroup{
				{Title: "The Apple", PrimaryType: "Single"},
				{Title: "banana", PrimaryType: "Single"},
				{Title: "Cherry", PrimaryType: "Single"},
			},
			expected: []int{0, 1, 2}, // The Apple (as Apple), banana, Cherry
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test default order
			sorted := SortReleaseGroups(tt.input, []string{"type", "year", "title"}, false)

			// Verify order matches expected
			for i, expectedIndex := range tt.expected {
				if i < len(sorted) && sorted[i].Title != tt.input[expectedIndex].Title {
					t.Errorf("Default order: at position %d, expected %s but got %s",
						i, tt.input[expectedIndex].Title, sorted[i].Title)
				}
			}
		})
	}
}
