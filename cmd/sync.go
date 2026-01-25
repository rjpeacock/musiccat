package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"musiccat/external/musicbrainz"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <id|range>",
	Short: "Sync release data from MusicBrainz",
	Long:  `Fetch release data from MusicBrainz and update missing format details based on release type. Accepts a single ID or a range (e.g., 50-65).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr := args[0]
		
		// Parse ID or range
		ids, err := parseIDRange(idStr)
		if err != nil {
			return err
		}

		// Open DB
		database, err := db.OpenDB()
		if err != nil {
			return err
		}
		defer database.Close()

		// Bootstrap if needed
		if err := db.BootstrapDB(database); err != nil {
			return err
		}

		// Process each ID
		updated := 0
		skipped := 0
		errors := 0

		for _, id := range ids {
			err := syncSingleID(database, id)
			if err != nil {
				if strings.Contains(err.Error(), "already set") || strings.Contains(err.Error(), "No auto-format") {
					skipped++
					fmt.Println(err.Error())
				} else {
					errors++
					fmt.Printf("Error syncing ID %d: %v\n", id, err)
				}
			} else {
				updated++
			}
		}

		fmt.Printf("\nSync complete: %d updated, %d skipped, %d errors\n", updated, skipped, errors)
		return nil
	},
}

func parseIDRange(idStr string) ([]int, error) {
	// Check if it's a range (contains hyphen)
	if strings.Contains(idStr, "-") {
		parts := strings.Split(idStr, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s (expected format: 50-65)", idStr)
		}

		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start ID: %s", parts[0])
		}

		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end ID: %s", parts[1])
		}

		if start > end {
			return nil, fmt.Errorf("start ID must be less than or equal to end ID")
		}

		// Build range
		ids := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			ids = append(ids, i)
		}
		return ids, nil
	}

	// Single ID
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %s", idStr)
	}
	return []int{id}, nil
}

func syncSingleID(database *sql.DB, id int) error {
	// Get ownership entry with release info
	var releaseID int64
	var formatCategory string
	var formatDetail sql.NullString
	var mbid sql.NullString
	var artist, title string

	err := database.QueryRow(`
		SELECT o.release_id, o.format_category, o.format_detail, r.musicbrainz_release_group_id, r.artist, r.title
		FROM ownership o
		JOIN releases r ON o.release_id = r.id
		WHERE o.id = ?`, id).Scan(&releaseID, &formatCategory, &formatDetail, &mbid, &artist, &title)
	if err != nil {
		return fmt.Errorf("ownership with ID %d not found", id)
	}

	// Check if format detail is already set
	if formatDetail.Valid && formatDetail.String != "" {
		return fmt.Errorf("ID %d: Format detail already set to '%s' (%s - %s)", id, formatDetail.String, artist, title)
	}

	// Check if MusicBrainz ID exists
	if !mbid.Valid || mbid.String == "" {
		return fmt.Errorf("ID %d: No MusicBrainz ID found - cannot sync", id)
	}

	// Fetch release group from MusicBrainz
	fmt.Printf("Fetching data for ID %d: %s - %s...\n", id, artist, title)
	releaseGroup, err := musicbrainz.GetReleaseGroup(mbid.String)
	if err != nil {
		return fmt.Errorf("failed to fetch from MusicBrainz: %w", err)
	}

	// Determine format detail using auto logic
	detail := autoFormatDetailSync(formatCategory, releaseGroup.PrimaryType)
	if detail == "" {
		return fmt.Errorf("ID %d: No auto-format detail applicable for %s with type %s", id, formatCategory, releaseGroup.PrimaryType)
	}

	// Update ownership with format detail
	_, err = database.Exec("UPDATE ownership SET format_detail = ? WHERE id = ?", detail, id)
	if err != nil {
		return fmt.Errorf("failed to update format detail: %w", err)
	}

	fmt.Printf("Updated ID %d: Set format detail to '%s' (type: %s)\n", id, detail, releaseGroup.PrimaryType)
	return nil
}

func autoFormatDetailSync(formatCategory, releaseType string) string {
	upperFormat := strings.ToUpper(formatCategory)

	switch upperFormat {
	case "CD", "CASSETTE":
		// For CD and Cassette, auto-set to the release type
		return releaseType
	case "VINYL":
		// For Vinyl, only auto-set to Album if it's an Album
		// Singles require manual 7" or 12" specification
		if releaseType == "Album" {
			return "Album"
		}
		return ""
	default:
		return ""
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
