package cmd

import (
	"fmt"
	"strings"

	"musiccat/external/musicbrainz"
	"musiccat/internal/config"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   `add "<artist name>"`,
	Short: "Search and add releases by artist",
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		if manual {
			if len(args) != 0 {
				return fmt.Errorf("--manual takes no arguments")
			}
			return addManual()
		}
		if len(args) == 0 {
			return fmt.Errorf("requires artist name argument")
		}
		artistName := strings.Join(args, " ")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		sortFields, _ := cmd.Flags().GetString("sort")
		desc, _ := cmd.Flags().GetBool("desc")
		albumOnly, _ := cmd.Flags().GetBool("album-only")
		singleOnly, _ := cmd.Flags().GetBool("single-only")
		year, _ := cmd.Flags().GetInt("year")
		titleFilter, _ := cmd.Flags().GetString("title")
		pirateFlag, _ := cmd.Flags().GetBool("pirate")

		return addFromMusicBrainz(artistName, pageSize, sortFields, desc, albumOnly, singleOnly, year, titleFilter, pirateFlag)
	},
}

func addFromMusicBrainz(artistName string, pageSize int, sortFields string, desc bool, albumOnly, singleOnly bool, year int, titleFilter string, pirateFlag bool) error {
	// Load config for current format
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.CurrentFormat == "" {
		return fmt.Errorf("no current format set; use 'musiccat set-format <FORMAT>' first")
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

	// Search artists
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
	selectedArtist, err := selectItem("Select artist (number): ", artists)
	if err != nil {
		return err
	}

	// Search release groups with filters applied via API query
	releaseGroups, err := musicbrainz.SearchReleaseGroups(selectedArtist.ID, albumOnly, singleOnly, year, titleFilter)
	if err != nil {
		return err
	}
	if len(releaseGroups) == 0 {
		return fmt.Errorf("no releases found for artist '%s' with specified filters", selectedArtist.Name)
	}

	// Apply sorting
	sortFieldsSlice := parseSortFields(sortFields)
	sortedGroups := SortReleaseGroups(releaseGroups, sortFieldsSlice, desc)

	selectedItems, err := selectReleasesWithPagination(sortedGroups, pageSize, cfg.CurrentFormat, albumOnly, singleOnly, year, titleFilter)
	if err != nil {
		return err
	}

	// Insert each selected release and ownership
	totalAdded := 0
	for _, item := range selectedItems {
		rg := item.releaseGroup
		var releaseYear *int = parseYear(rg.FirstReleaseDate)
		releaseID, err := db.InsertRelease(database, selectedArtist.Name, rg.Title, releaseYear, &rg.ID)
		if err != nil {
			return err
		}

		// Insert multiple ownership entries if quantity > 1
		for i := 0; i < item.quantity; i++ {
			var notes *string

			// Prompt for notes for each copy if quantity > 1
			if item.quantity > 1 {
				prompt := fmt.Sprintf("Notes for copy %d/%d of %s (optional): ", i+1, item.quantity, rg.Title)
				noteStr := promptString(prompt)
				if noteStr != "" {
					notes = &noteStr
				}
			} else if item.notes != "" {
				// Use existing notes for single copy
				notes = &item.notes
			}

			finalPirate := item.pirate || pirateFlag
			_, err = db.InsertOwnership(database, releaseID, cfg.CurrentFormat, nil, nil, nil, nil, notes, nil, item.promo, finalPirate)
			if err != nil {
				return err
			}
			totalAdded++
		}
	}

	fmt.Printf("Added %d ownership entries to collection.\n", totalAdded)
	return nil
}

func addManual() error {
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

	// Prompts
	artist := promptString("Artist: ")
	title := promptString("Title: ")
	manualYear := promptOptionalInt("Year (optional): ")
	formatCategory := promptValidFormat("Format category (CD, Vinyl, Cassette): ")
	formatDetailInput := promptFormatDetail(formatCategory)
	var formatDetail *string
	if formatDetailInput != "" {
		formatDetail = &formatDetailInput
	}

	// Insert release
	id, err := db.InsertRelease(database, artist, title, manualYear, nil)
	if err != nil {
		return err
	}

	// Insert ownership
	_, err = db.InsertOwnership(database, id, formatCategory, formatDetail, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		return err
	}

	fmt.Println("Added release to collection.")
	return nil
}

func init() {
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	addCmd.Flags().Int("page-size", 100, "Number of releases per page")
	addCmd.Flags().String("sort", "", "Sort by field(s): type, year, title (comma-separated)")
	addCmd.Flags().Bool("desc", false, "Reverse sort order")
	addCmd.Flags().Bool("album-only", false, "Show only albums")
	addCmd.Flags().Bool("single-only", false, "Show only singles")
	addCmd.Flags().Int("year", 0, "Filter by specific release year")
	addCmd.Flags().String("title", "", "Filter by release title (partial match)")
	addCmd.Flags().Bool("pirate", false, "Mark releases as pirate copies")
	rootCmd.AddCommand(addCmd)
}
