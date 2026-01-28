package db

import (
	"database/sql"
	"testing"
)

func TestGetOrCreateTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a new tag
	id1, err := GetOrCreateTag(db, "promo")
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Get the same tag again (should return the same ID)
	id2, err := GetOrCreateTag(db, "promo")
	if err != nil {
		t.Fatalf("Failed to get existing tag: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same tag ID, got %d and %d", id1, id2)
	}

	// Create a different tag
	id3, err := GetOrCreateTag(db, "signed")
	if err != nil {
		t.Fatalf("Failed to create second tag: %v", err)
	}

	if id3 == id1 {
		t.Error("Expected different tag IDs for different tags")
	}
}

func TestAddTagToOwnership(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Create a tag and add it to ownership
	tagID, err := GetOrCreateTag(db, "promo")
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	err = AddTagToOwnership(db, ownershipID, tagID)
	if err != nil {
		t.Fatalf("Failed to add tag to ownership: %v", err)
	}

	// Verify the tag was added
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0] != "promo" {
		t.Errorf("Expected tag 'promo', got '%s'", tags[0])
	}
}

func TestAddTagToOwnershipIdempotent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Create a tag
	tagID, err := GetOrCreateTag(db, "promo")
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	// Add the tag twice
	err = AddTagToOwnership(db, ownershipID, tagID)
	if err != nil {
		t.Fatalf("Failed to add tag to ownership first time: %v", err)
	}

	err = AddTagToOwnership(db, ownershipID, tagID)
	if err != nil {
		t.Fatalf("Failed to add tag to ownership second time: %v", err)
	}

	// Verify only one tag was added
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
}

func TestMultipleTagsOnOwnership(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Add multiple tags
	tagNames := []string{"promo", "signed", "missing-sleeve"}
	for _, name := range tagNames {
		tagID, err := GetOrCreateTag(db, name)
		if err != nil {
			t.Fatalf("Failed to create tag '%s': %v", name, err)
		}
		err = AddTagToOwnership(db, ownershipID, tagID)
		if err != nil {
			t.Fatalf("Failed to add tag '%s' to ownership: %v", name, err)
		}
	}

	// Verify all tags were added
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	// Tags should be sorted
	expectedTags := []string{"missing-sleeve", "promo", "signed"}
	for i, expected := range expectedTags {
		if tags[i] != expected {
			t.Errorf("Expected tag '%s' at index %d, got '%s'", expected, i, tags[i])
		}
	}
}

func TestRemoveTagFromOwnership(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Add two tags
	tagID1, _ := GetOrCreateTag(db, "promo")
	tagID2, _ := GetOrCreateTag(db, "signed")
	AddTagToOwnership(db, ownershipID, tagID1)
	AddTagToOwnership(db, ownershipID, tagID2)

	// Remove one tag
	err = RemoveTagFromOwnership(db, ownershipID, tagID1)
	if err != nil {
		t.Fatalf("Failed to remove tag: %v", err)
	}

	// Verify only one tag remains
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0] != "signed" {
		t.Errorf("Expected tag 'signed', got '%s'", tags[0])
	}

	// Verify the tag itself still exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tags WHERE id = ?", tagID1).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count tags: %v", err)
	}
	if count != 1 {
		t.Error("Tag should still exist in tags table")
	}
}

func TestRemoveAllTagsFromOwnership(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Add multiple tags
	for _, name := range []string{"promo", "signed", "missing-sleeve"} {
		tagID, _ := GetOrCreateTag(db, name)
		AddTagToOwnership(db, ownershipID, tagID)
	}

	// Remove all tags
	err = RemoveAllTagsFromOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to remove all tags: %v", err)
	}

	// Verify no tags remain
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}

func TestRenameTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Create a tag and add it to ownership
	tagID, _ := GetOrCreateTag(db, "promotional")
	AddTagToOwnership(db, ownershipID, tagID)

	// Rename the tag
	err = RenameTag(db, "promotional", "promo")
	if err != nil {
		t.Fatalf("Failed to rename tag: %v", err)
	}

	// Verify the tag was renamed
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0] != "promo" {
		t.Errorf("Expected tag 'promo', got '%s'", tags[0])
	}

	// Verify old tag name doesn't exist
	_, err = GetTagIDByName(db, "promotional")
	if err != sql.ErrNoRows {
		t.Error("Old tag name should not exist")
	}
}

func TestDeleteTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := BootstrapDB(db); err != nil {
		t.Fatal(err)
	}

	// Create a release and ownership
	releaseID, err := InsertRelease(db, "Test Artist", "Test Album", nil, nil)
	if err != nil {
		t.Fatalf("Failed to insert release: %v", err)
	}

	ownershipID, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatalf("Failed to insert ownership: %v", err)
	}

	// Create a tag and add it to ownership
	tagID, _ := GetOrCreateTag(db, "promo")
	AddTagToOwnership(db, ownershipID, tagID)

	// Delete the tag
	err = DeleteTag(db, "promo")
	if err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}

	// Verify the tag is gone
	tags, err := GetTagsForOwnership(db, ownershipID)
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}

	// Verify tag doesn't exist in tags table
	_, err = GetTagIDByName(db, "promo")
	if err != sql.ErrNoRows {
		t.Error("Tag should not exist")
	}
}
