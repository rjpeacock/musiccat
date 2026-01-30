package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"musiccat/cmd/helpers"
	"musiccat/external/musicbrainz"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var missingCmd = &cobra.Command{
	Use:     `missing "<artist name>"`,
	Aliases: []string{"m"},
	Short:   "Find releases you don't own yet for an artist",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires artist name argument")
		}
		artistName := strings.Join(args, " ")
		
		exact, _ := cmd.Flags().GetBool("exact")
		albumOnly, _ := cmd.Flags().GetBool("album")
		singleOnly, _ := cmd.Flags().GetBool("single")
		epOnly, _ := cmd.Flags().GetBool("ep")
		compilationOnly, _ := cmd.Flags().GetBool("compilation")
		liveOnly, _ := cmd.Flags().GetBool("live")
		soundtrackOnly, _ := cmd.Flags().GetBool("soundtrack")
		year, _ := cmd.Flags().GetInt("year")
		titleFilter, _ := cmd.Flags().GetString("title")

		return findMissingReleases(artistName, exact, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly, year, titleFilter)
	},
}

func findMissingReleases(artistName string, exact bool, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly bool, year int, titleFilter string) error {
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

	// Search for artist on MusicBrainz (exact flag currently not implemented)
	artists, err := musicbrainz.SearchArtists(artistName)
	if err != nil {
		return err
	}
	if len(artists) == 0 {
		return fmt.Errorf("no artists found for '%s'", artistName)
	}

	// Display and select artist
	fmt.Println("Artists found:")
	for i, artist := range artists {
		disamb := ""
		if artist.Disambiguation != "" {
			disamb = " (" + artist.Disambiguation + ")"
		}
		fmt.Printf("%d. %s%s\n", i+1, artist.Name, disamb)
	}
	selectedArtist, err := helpers.SelectItem("Select artist (number): ", artists)
	if err != nil {
		return err
	}

	// Fetch all release groups from MusicBrainz
	allReleaseGroups, err := musicbrainz.SearchReleaseGroups(
		selectedArtist.ID,
		albumOnly,
		singleOnly,
		epOnly,
		compilationOnly,
		liveOnly,
		soundtrackOnly,
		year,
		titleFilter,
		0, // limit 0 = fetch all
	)
	if err != nil {
		return err
	}
	if len(allReleaseGroups) == 0 {
		return fmt.Errorf("no releases found for artist '%s' with specified filters", selectedArtist.Name)
	}

	// Get all owned release group IDs for this artist from the database
	ownedIDs, err := getOwnedReleaseGroupIDs(database, selectedArtist.Name)
	if err != nil {
		return err
	}

	// Filter out owned releases
	var missingReleases []musicbrainz.ReleaseGroup
	for _, rg := range allReleaseGroups {
		if !contains(ownedIDs, rg.ID) {
			missingReleases = append(missingReleases, rg)
		}
	}

	// Display results
	if len(missingReleases) == 0 {
		fmt.Printf("✓ You own all %d releases matching the filters for %s!\n", len(allReleaseGroups), selectedArtist.Name)
		return nil
	}

	fmt.Printf("\nFound %d/%d releases you don't own yet:\n\n", len(missingReleases), len(allReleaseGroups))

	// Sort missing releases using existing helper
	sortFieldsSlice := helpers.ParseSortFields("type,year,title")
	sortedReleases := helpers.SortReleaseGroups(missingReleases, sortFieldsSlice, false)

	for i, rg := range sortedReleases {
		year := "????"
		if rg.FirstReleaseDate != "" && len(rg.FirstReleaseDate) >= 4 {
			year = rg.FirstReleaseDate[:4]
		}
		
		typeDisplay := rg.PrimaryType
		if len(rg.SecondaryTypes) > 0 {
			typeDisplay = fmt.Sprintf("%s + %s", rg.PrimaryType, strings.Join(rg.SecondaryTypes, " + "))
		}
		
		fmt.Printf("%3d. %s - %s [%s]\n", i+1, year, rg.Title, typeDisplay)
	}

	return nil
}

// getOwnedReleaseGroupIDs returns all MusicBrainz release group IDs owned for an artist
func getOwnedReleaseGroupIDs(database *sql.DB, artistName string) ([]string, error) {
	query := `
		SELECT DISTINCT r.musicbrainz_release_group_id
		FROM releases r
		JOIN ownership o ON r.id = o.release_id
		WHERE r.artist = ? AND r.musicbrainz_release_group_id IS NOT NULL
	`
	
	rows, err := database.Query(query, artistName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	missingCmd.Flags().Bool("exact", false, "Exact artist name match")
	missingCmd.Flags().Bool("album", false, "Show only albums (excludes compilations, live, soundtracks)")
	missingCmd.Flags().Bool("single", false, "Show only singles")
	missingCmd.Flags().Bool("ep", false, "Show only EPs")
	missingCmd.Flags().Bool("compilation", false, "Show only compilations")
	missingCmd.Flags().Bool("live", false, "Show only live albums")
	missingCmd.Flags().Bool("soundtrack", false, "Show only soundtracks")
	missingCmd.Flags().Int("year", 0, "Filter by release year")
	missingCmd.Flags().String("title", "", "Filter by title (partial match)")

	rootCmd.AddCommand(missingCmd)
}
