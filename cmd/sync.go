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
	Use:   "sync <id>",
	Short: "Sync release data from MusicBrainz",
	Long:  `Fetch release data from MusicBrainz and update missing format details based on release type.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr := args[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", idStr)
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

		// Get ownership entry with release info
		var releaseID int64
		var formatCategory string
		var formatDetail sql.NullString
		var mbid sql.NullString
		var artist, title string

		err = database.QueryRow(`
			SELECT o.release_id, o.format_category, o.format_detail, r.musicbrainz_release_group_id, r.artist, r.title
			FROM ownership o
			JOIN releases r ON o.release_id = r.id
			WHERE o.id = ?`, id).Scan(&releaseID, &formatCategory, &formatDetail, &mbid, &artist, &title)
		if err != nil {
			return fmt.Errorf("ownership with ID %d not found", id)
		}

		// Check if format detail is already set
		if formatDetail.Valid && formatDetail.String != "" {
			fmt.Printf("Format detail already set to '%s' for ID %d (%s - %s)\n", formatDetail.String, id, artist, title)
			return nil
		}

		// Check if MusicBrainz ID exists
		if !mbid.Valid || mbid.String == "" {
			return fmt.Errorf("no MusicBrainz ID found for this release - cannot sync")
		}

		// Fetch release group from MusicBrainz
		fmt.Printf("Fetching data for %s - %s...\n", artist, title)
		releaseGroup, err := musicbrainz.GetReleaseGroup(mbid.String)
		if err != nil {
			return fmt.Errorf("failed to fetch from MusicBrainz: %w", err)
		}

		// Determine format detail using auto logic
		detail := autoFormatDetailSync(formatCategory, releaseGroup.PrimaryType)
		if detail == "" {
			fmt.Printf("No auto-format detail applicable for %s with type %s\n", formatCategory, releaseGroup.PrimaryType)
			return nil
		}

		// Update ownership with format detail
		_, err = database.Exec("UPDATE ownership SET format_detail = ? WHERE id = ?", detail, id)
		if err != nil {
			return fmt.Errorf("failed to update format detail: %w", err)
		}

		fmt.Printf("Updated ID %d: Set format detail to '%s' (type: %s)\n", id, detail, releaseGroup.PrimaryType)
		return nil
	},
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
