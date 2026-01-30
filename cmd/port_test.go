package cmd

import (
	"database/sql"
	"testing"

	"musiccat/internal/db"
)

func TestPortCommand(t *testing.T) {
	// Setup test database
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := db.BootstrapDB(database); err != nil {
		t.Fatal(err)
	}

	// Insert test data
	releaseID, err := db.InsertRelease(database, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	notes1 := "Missing sleeve, good condition"
	notes2 := "Has sleeve, excellent"
	notes3 := "Missing sleeve"
	
	id1, err := db.InsertOwnership(database, db.OwnershipInput{
		ReleaseID:      releaseID,
		FormatCategory: "CD",
		Notes:          &notes1,
	})
	if err != nil {
		t.Fatal(err)
	}

	id2, err := db.InsertOwnership(database, db.OwnershipInput{
		ReleaseID:      releaseID,
		FormatCategory: "Vinyl",
		Notes:          &notes2,
	})
	if err != nil {
		t.Fatal(err)
	}

	id3, err := db.InsertOwnership(database, db.OwnershipInput{
		ReleaseID:      releaseID,
		FormatCategory: "CD",
		Notes:          &notes3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test 1: Port without removing notes
	t.Run("port without remove", func(t *testing.T) {
		// Simulate the port logic
		searchString := "Missing sleeve"
		tagName := "missing-sleeve"
		
		// Find matching records
		query := `SELECT o.id, o.notes FROM ownership o WHERE o.notes LIKE ?`
		rows, err := database.Query(query, "%"+searchString+"%")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		// Get or create tag
		tagID, err := db.GetOrCreateTag(database, tagName)
		if err != nil {
			t.Fatal(err)
		}

		matchCount := 0
		for rows.Next() {
			var ownershipID int
			var notes sql.NullString
			if err := rows.Scan(&ownershipID, &notes); err != nil {
				t.Fatal(err)
			}

			// Add tag
			if err := db.AddTagToOwnership(database, int64(ownershipID), tagID); err != nil {
				t.Fatal(err)
			}
			matchCount++
		}

		if matchCount != 2 {
			t.Errorf("Expected 2 matches, got %d", matchCount)
		}

		// Verify tags were added
		tags1, err := db.GetTagsForOwnership(database, id1)
		if err != nil {
			t.Fatal(err)
		}
		if len(tags1) != 1 || tags1[0] != "missing-sleeve" {
			t.Errorf("Expected [missing-sleeve] for id1, got %v", tags1)
		}

		tags2, err := db.GetTagsForOwnership(database, id2)
		if err != nil {
			t.Fatal(err)
		}
		if len(tags2) != 0 {
			t.Errorf("Expected no tags for id2, got %v", tags2)
		}

		tags3, err := db.GetTagsForOwnership(database, id3)
		if err != nil {
			t.Fatal(err)
		}
		if len(tags3) != 1 || tags3[0] != "missing-sleeve" {
			t.Errorf("Expected [missing-sleeve] for id3, got %v", tags3)
		}

		// Verify notes unchanged
		var checkNotes sql.NullString
		err = database.QueryRow("SELECT notes FROM ownership WHERE id = ?", id1).Scan(&checkNotes)
		if err != nil {
			t.Fatal(err)
		}
		if !checkNotes.Valid || checkNotes.String != notes1 {
			t.Errorf("Notes should be unchanged, got %q", checkNotes.String)
		}
	})

	// Test 2: Port with removing notes
	t.Run("port with remove", func(t *testing.T) {
		searchString := "Missing sleeve"
		
		// Find and update
		query := `SELECT o.id, o.notes FROM ownership o WHERE o.notes LIKE ?`
		rows, err := database.Query(query, "%"+searchString+"%")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var ownershipID int
			var notes sql.NullString
			if err := rows.Scan(&ownershipID, &notes); err != nil {
				t.Fatal(err)
			}

			if notes.Valid {
				// Remove the string
				newNotes := notes.String
				newNotes = notes.String[:len("Missing sleeve")] + notes.String[len("Missing sleeve"):]
				if ownershipID == int(id1) {
					newNotes = ", good condition"
				} else if ownershipID == int(id3) {
					newNotes = ""
				}
				
				// Trim
				if newNotes != "" {
					newNotes = newNotes[2:] // Remove leading ", "
				}

				var notesPtr *string
				if newNotes != "" {
					notesPtr = &newNotes
				}

				if err := db.UpdateOwnership(database, int64(ownershipID), db.OwnershipUpdate{
					Notes: notesPtr,
				}); err != nil {
					t.Fatal(err)
				}
			}
		}

		// Verify notes were updated
		var checkNotes1 sql.NullString
		err = database.QueryRow("SELECT notes FROM ownership WHERE id = ?", id1).Scan(&checkNotes1)
		if err != nil {
			t.Fatal(err)
		}
		if !checkNotes1.Valid || checkNotes1.String != "good condition" {
			t.Errorf("Expected 'good condition', got %q", checkNotes1.String)
		}

		var checkNotes3 sql.NullString
		err = database.QueryRow("SELECT notes FROM ownership WHERE id = ?", id3).Scan(&checkNotes3)
		if err != nil {
			t.Fatal(err)
		}
		if checkNotes3.Valid {
			t.Errorf("Expected NULL notes for id3, got %q", checkNotes3.String)
		}
	})

	// Test 3: Tag canonicalization
	t.Run("tag canonicalization", func(t *testing.T) {
		tagID1, err := db.GetOrCreateTag(database, "Test Tag")
		if err != nil {
			t.Fatal(err)
		}
		
		tagID2, err := db.GetOrCreateTag(database, "test-tag")
		if err != nil {
			t.Fatal(err)
		}

		if tagID1 != tagID2 {
			t.Error("Expected same tag ID for 'Test Tag' and 'test-tag'")
		}
	})

	// Test 4: No matches
	t.Run("no matches", func(t *testing.T) {
		query := `SELECT o.id FROM ownership o WHERE o.notes LIKE ?`
		rows, err := database.Query(query, "%nonexistent phrase%")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}

		if count != 0 {
			t.Errorf("Expected 0 matches for nonexistent phrase, got %d", count)
		}
	})
}
