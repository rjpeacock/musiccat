package cmd

import (
	"database/sql"
	"testing"

	"musiccat/internal/db"

	_ "modernc.org/sqlite"
)

func TestMultiVariantOwnership(t *testing.T) {
	tests := []struct {
		name             string
		releaseArtist    string
		releaseTitle     string
		year             *int
		ownershipEntries []OwnershipEntry
	}{
		{
			name:          "single ownership entry",
			releaseArtist: "Test Artist",
			releaseTitle:  "Test Album",
			year:          intPtr(2023),
			ownershipEntries: []OwnershipEntry{
				{
					formatCategory: "CD",
					formatDetail:   stringPtr("Album"),
					isPromo:        false,
					notes:          "",
				},
			},
		},
		{
			name:          "multiple ownership entries for same release",
			releaseArtist: "Test Artist",
			releaseTitle:  "Variant Album",
			year:          intPtr(2023),
			ownershipEntries: []OwnershipEntry{
				{
					formatCategory: "CD",
					formatDetail:   stringPtr("Album"),
					isPromo:        false,
					notes:          "",
				},
				{
					formatCategory: "Vinyl",
					formatDetail:   stringPtr("LP"),
					isPromo:        true,
					notes:          "Colored vinyl variant",
				},
				{
					formatCategory: "Vinyl",
					formatDetail:   stringPtr("Picture Disc"),
					isPromo:        false,
					notes:          "Limited edition",
				},
			},
		},
		{
			name:          "ownership with null format detail and notes",
			releaseArtist: "Minimal Artist",
			releaseTitle:  "Minimal Album",
			year:          nil,
			ownershipEntries: []OwnershipEntry{
				{
					formatCategory: "Cassette",
					formatDetail:   nil,
					isPromo:        false,
					notes:          "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup in-memory database
			testDB, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatalf("Failed to open test database: %v", err)
			}
			defer testDB.Close()

			// Bootstrap database
			if err := db.BootstrapDB(testDB); err != nil {
				t.Fatalf("Failed to bootstrap test database: %v", err)
			}

			// Insert release
			releaseID, err := db.InsertRelease(testDB, tt.releaseArtist, tt.releaseTitle, tt.year, nil)
			if err != nil {
				t.Fatalf("Failed to insert release: %v", err)
			}
			if releaseID == 0 {
				t.Fatal("Expected non-zero release ID")
			}

			// Insert ownership entries
			var ownershipIDs []int64
			for _, entry := range tt.ownershipEntries {
				ownershipID, err := db.InsertOwnership(testDB, releaseID, entry.formatCategory, entry.formatDetail, nil, nil, nil, entry.notesPtr(), entry.isPromo)
				if err != nil {
					t.Fatalf("Failed to insert ownership: %v", err)
				}
				if ownershipID == 0 {
					t.Fatal("Expected non-zero ownership ID")
				}
				ownershipIDs = append(ownershipIDs, ownershipID)
			}

			// Verify ownership entries were inserted correctly
			for i, expectedEntry := range tt.ownershipEntries {
				var id int
				var formatCategory, formatDetail, notes string
				var formatDetailNull, notesNull sql.NullString
				var isPromo bool

				err = testDB.QueryRow(`
					SELECT id, format_category, format_detail, is_promo, notes 
					FROM ownership 
					WHERE id = ?
				`, ownershipIDs[i]).Scan(&id, &formatCategory, &formatDetailNull, &isPromo, &notesNull)

				if err != nil {
					t.Fatalf("Failed to query ownership entry %d: %v", i, err)
				}

				// Handle nullable fields
				if formatDetailNull.Valid {
					formatDetail = formatDetailNull.String
				}
				if notesNull.Valid {
					notes = notesNull.String
				}

				// Verify format category
				if formatCategory != expectedEntry.formatCategory {
					t.Errorf("Expected format_category %s, got %s", expectedEntry.formatCategory, formatCategory)
				}

				// Verify format detail
				if expectedEntry.formatDetail == nil {
					if formatDetailNull.Valid {
						t.Errorf("Expected NULL format_detail, got %s", formatDetail)
					}
				} else {
					if !formatDetailNull.Valid || formatDetail != *expectedEntry.formatDetail {
						t.Errorf("Expected format_detail %s, got %s", *expectedEntry.formatDetail, formatDetail)
					}
				}

				// Verify promo status
				if isPromo != expectedEntry.isPromo {
					t.Errorf("Expected is_promo %t, got %t", expectedEntry.isPromo, isPromo)
				}

				// Verify notes
				if expectedEntry.notes == "" {
					if notesNull.Valid {
						t.Errorf("Expected NULL notes, got %s", notes)
					}
				} else {
					if !notesNull.Valid || notes != expectedEntry.notes {
						t.Errorf("Expected notes %s, got %s", expectedEntry.notes, notes)
					}
				}
			}

			// Test filtering by format category
			countByFormat, err := testDB.QueryRow(`
				SELECT COUNT(*) 
				FROM ownership o
				JOIN releases r ON o.release_id = r.id
				WHERE r.artist = ? AND o.format_category = ?
			`, tt.releaseArtist, "Vinyl").Scan()
			if err != nil {
				t.Fatalf("Failed to count by format: %v", err)
			}

			expectedVinylCount := 0
			for _, entry := range tt.ownershipEntries {
				if entry.formatCategory == "Vinyl" {
					expectedVinylCount++
				}
			}

			if countByFormat != expectedVinylCount {
				t.Errorf("Expected %d Vinyl entries, got %v", expectedVinylCount, countByFormat)
			}
		})
	}
}

func TestFormatDetailSuggestions(t *testing.T) {
	tests := []struct {
		name           string
		formatCategory string
		expectedCount  int
		expectedValues []string
	}{
		{
			name:           "CD format details",
			formatCategory: "CD",
			expectedCount:  6,
			expectedValues: []string{"Album", "Single", "EP", "Maxi-Single", "Promo", "Digipak", "Jewel Case"},
		},
		{
			name:           "Vinyl format details",
			formatCategory: "Vinyl",
			expectedCount:  7,
			expectedValues: []string{"LP", "12\"", "10\"", "7\"", "Single", "EP", "Picture Disc", "Colored Vinyl"},
		},
		{
			name:           "Cassette format details",
			formatCategory: "Cassette",
			expectedCount:  4,
			expectedValues: []string{"Album", "Single", "Tape", "Cassette"},
		},
		{
			name:           "unknown format category",
			formatCategory: "Unknown",
			expectedCount:  0,
			expectedValues: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, exists := FormatDetailSuggestions[tt.formatCategory]

			if !exists && tt.expectedCount > 0 {
				t.Errorf("Expected format category %s to exist in suggestions", tt.formatCategory)
				return
			}

			if exists && len(suggestions) != tt.expectedCount {
				t.Errorf("Expected %d suggestions for %s, got %d", tt.expectedCount, tt.formatCategory, len(suggestions))
				return
			}

			if tt.expectedValues != nil {
				for _, expected := range tt.expectedValues {
					found := false
					for _, actual := range suggestions {
						if actual == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected suggestion %s not found in %v", expected, suggestions)
					}
				}
			}
		})
	}
}

// Helper types and functions for testing
type OwnershipEntry struct {
	formatCategory string
	formatDetail   *string
	isPromo        bool
	notes          string
}

func (e OwnershipEntry) notesPtr() *string {
	if e.notes == "" {
		return nil
	}
	return &e.notes
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
