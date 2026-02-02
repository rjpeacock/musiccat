package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"musiccat/cmd/helpers"
	"musiccat/external/musicbrainz"
	"musiccat/internal/config"
	"musiccat/internal/db"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     `add "<artist name>" | add --release-id <ID>`,
	Aliases: []string{"a"},
	Short:   "Search and add releases by artist or add another copy by release ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		releaseID, _ := cmd.Flags().GetInt("release-id")
		
		if manual {
			if len(args) != 0 {
				return fmt.Errorf("--manual takes no arguments")
			}
			return addManual()
		}
		
		if releaseID > 0 {
			if len(args) != 0 {
				return fmt.Errorf("--release-id takes no artist name argument")
			}
			return addByReleaseID(releaseID)
		}
		
		if len(args) == 0 {
			return fmt.Errorf("requires artist name argument or --release-id flag")
		}
		artistName := strings.Join(args, " ")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		sortFields, _ := cmd.Flags().GetString("sort")
		desc, _ := cmd.Flags().GetBool("desc")
		albumOnly, _ := cmd.Flags().GetBool("album")
		singleOnly, _ := cmd.Flags().GetBool("single")
		epOnly, _ := cmd.Flags().GetBool("ep")
		compilationOnly, _ := cmd.Flags().GetBool("compilation")
		liveOnly, _ := cmd.Flags().GetBool("live")
		soundtrackOnly, _ := cmd.Flags().GetBool("soundtrack")
		year, _ := cmd.Flags().GetInt("year")
		titleFilter, _ := cmd.Flags().GetString("title")
		pirateFlag, _ := cmd.Flags().GetBool("pirate")

		return addFromMusicBrainz(artistName, pageSize, sortFields, desc, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly, year, titleFilter, pirateFlag)
	},
}

func addFromMusicBrainz(artistName string, pageSize int, sortFields string, desc bool, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly bool, year int, titleFilter string, pirateFlag bool) error {
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

	// Search and select artist with pagination
	selectedArtist, err := helpers.SelectArtist(artistName)
	if err != nil {
		return err
	}

	// Search release groups with filters applied via API query
	allReleaseGroups, err := musicbrainz.SearchReleaseGroups(selectedArtist.ID, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly, year, titleFilter, 0)
	if err != nil {
		return err
	}
	if len(allReleaseGroups) == 0 {
		return fmt.Errorf("no releases found for artist '%s' with specified filters", selectedArtist.Name)
	}

	// Apply sorting
	sortFieldsSlice := helpers.ParseSortFields(sortFields)
	allReleaseGroups = helpers.SortReleaseGroups(allReleaseGroups, sortFieldsSlice, desc)

	// Display and select releases, fetching more if needed
	var selectedItems []helpers.SelectionItem
	currentPage := 0
	for {
		items, needMore, _, err := helpers.SelectReleasesWithPagination(allReleaseGroups, pageSize, cfg.CurrentFormat, albumOnly, singleOnly, year, titleFilter, currentPage)
		if err != nil {
			return err
		}
		if needMore {
			// Fetch more results from API
			fmt.Println("Fetching more results...")
			offset := len(allReleaseGroups)
			moreGroups, err := musicbrainz.SearchReleaseGroups(selectedArtist.ID, albumOnly, singleOnly, epOnly, compilationOnly, liveOnly, soundtrackOnly, year, titleFilter, offset)
			if err != nil {
				fmt.Printf("Error fetching more results: %v\n", err)
				continue
			}
			if len(moreGroups) == 0 {
				fmt.Println("No more results available.")
				continue
			}
			// Sort the new batch and append (don't re-sort everything to keep stable numbering)
			moreGroups = helpers.SortReleaseGroups(moreGroups, sortFieldsSlice, desc)
			allReleaseGroups = append(allReleaseGroups, moreGroups...)
			// Continue from the first page of newly fetched results
			currentPage = offset / pageSize
			continue
		}
		selectedItems = items
		break
	}

	// Insert each selected release and ownership
	totalAdded := 0
	for _, item := range selectedItems {
		rg := item.ReleaseGroup
		var releaseYear *int = helpers.ParseYear(rg.FirstReleaseDate)
		releaseID, err := db.InsertRelease(database, selectedArtist.Name, rg.Title, releaseYear, &rg.ID)
		if err != nil {
			return err
		}

		// Insert multiple ownership entries if quantity > 1
		for i := 0; i < item.Quantity; i++ {
			var notes *string

			// Prompt for notes for each copy if quantity > 1
			if item.Quantity > 1 {
				prompt := fmt.Sprintf("Notes for copy %d/%d of %s (optional): ", i+1, item.Quantity, rg.Title)
				noteStr := helpers.PromptString(prompt)
				if noteStr != "" {
					notes = &noteStr
				}
			} else if item.Notes != "" {
				// Use existing notes for single copy
				notes = &item.Notes
			}

			// Auto-set format detail based on type and format category
			var formatDetail *string
			if rg.PrimaryType != "" {
				detail := autoFormatDetail(cfg.CurrentFormat, rg.PrimaryType)
				if detail != "" {
					formatDetail = &detail
				}
			}

			finalPirate := item.Pirate || pirateFlag
			_, err = db.InsertOwnership(database, db.OwnershipInput{
				ReleaseID:      releaseID,
				FormatCategory: cfg.CurrentFormat,
				FormatDetail:   formatDetail,
				Notes:          notes,
				IsPromo:        item.Promo,
				IsPirate:       finalPirate,
			})
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

	lastArtist := ""
	
	for {
		// Prompts
		var artist string
		if lastArtist == "" {
			artist = helpers.PromptString("Artist: ")
		} else {
			fmt.Printf("Artist (Enter to repeat '%s'): ", lastArtist)
			artist = helpers.PromptString("")
			if artist == "" {
				artist = lastArtist
			}
		}
		
		if artist == "" {
			fmt.Println("Artist is required.")
			continue
		}
		
		lastArtist = artist
		
		title := helpers.PromptString("Title: ")
		if title == "" {
			fmt.Println("Title is required.")
			continue
		}
		
		manualYear := helpers.PromptOptionalInt("Year (optional): ")
		formatCategory := helpers.PromptValidFormat("Format category (CD, Vinyl, Cassette): ")
		formatDetailInput := helpers.PromptFormatDetail(formatCategory)
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
		_, err = db.InsertOwnership(database, db.OwnershipInput{
			ReleaseID:      id,
			FormatCategory: formatCategory,
			FormatDetail:   formatDetail,
		})
		if err != nil {
			return err
		}

		fmt.Println("Added release to collection.")
		
		// Ask if they want to add another
		fmt.Print("\nAdd another? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			break
		}
		fmt.Println()
	}

	return nil
}

func addByReleaseID(releaseID int) error {
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

	// Verify release exists and get details
	var artist, title string
	var yearNull sql.NullInt32
	err = database.QueryRow(`
		SELECT artist, title, year 
		FROM releases 
		WHERE id = ?
	`, releaseID).Scan(&artist, &title, &yearNull)

	if err == sql.ErrNoRows {
		return fmt.Errorf("no release found with ID %d", releaseID)
	}
	if err != nil {
		return err
	}

	yearStr := "????"
	if yearNull.Valid {
		yearStr = fmt.Sprintf("%d", yearNull.Int32)
	}

	fmt.Printf("Adding another copy of:\n  %s - %s (%s)\n\n", artist, title, yearStr)

	// Prompt for format detail
	formatDetailInput := helpers.PromptFormatDetail(cfg.CurrentFormat)
	var formatDetail *string
	if formatDetailInput != "" {
		formatDetail = &formatDetailInput
	}

	// Prompt for notes
	notesInput := helpers.PromptString("Notes (optional): ")
	var notes *string
	if notesInput != "" {
		notes = &notesInput
	}

	// Prompt for promo
	isPromo := helpers.PromptOptionalBool("Is this a promo? (y/n, optional): ", false)
	promoVal := false
	if isPromo != nil {
		promoVal = *isPromo
	}

	// Prompt for pirate
	isPirate := helpers.PromptOptionalBool("Is this a pirate copy? (y/n, optional): ", false)
	pirateVal := false
	if isPirate != nil {
		pirateVal = *isPirate
	}

	// Insert ownership
	_, err = db.InsertOwnership(database, db.OwnershipInput{
		ReleaseID:      int64(releaseID),
		FormatCategory: cfg.CurrentFormat,
		FormatDetail:   formatDetail,
		Notes:          notes,
		IsPromo:        promoVal,
		IsPirate:       pirateVal,
	})
	if err != nil {
		return err
	}

	fmt.Println("Added ownership entry.")
	return nil
}

// autoFormatDetail determines the format detail based on format category and release type
func autoFormatDetail(formatCategory, releaseType string) string {
	// Check if there's a saved format detail setting that overrides inference
	cfg, err := config.LoadConfig()
	if err == nil && cfg.CurrentFormatDetail != "" {
		return cfg.CurrentFormatDetail
	}
	
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
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	addCmd.Flags().Int("release-id", 0, "Add another copy of an existing release by release ID")
	addCmd.Flags().Int("page-size", 50, "Number of releases per page")
	addCmd.Flags().String("sort", "", "Sort by field(s): type, year, title (comma-separated)")
	addCmd.Flags().Bool("desc", false, "Reverse sort order")
	addCmd.Flags().Bool("pirate", false, "Mark releases as pirate copies")
	
	// Add common release group filter flags
	ReleaseGroupFilterFlags(addCmd)
	
	// Shell completions
	addCmd.RegisterFlagCompletionFunc("sort", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"type", "year", "title", "type,year", "type,year,title"}, cobra.ShellCompDirectiveNoFileComp
	})
	
	rootCmd.AddCommand(addCmd)
}
