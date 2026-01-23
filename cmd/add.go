package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"musiccat/internal/config"
	"musiccat/internal/db"
)

type ArtistSearchResponse struct {
	Artists []Artist `json:"artists"`
}

type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
}

type ReleaseGroupSearchResponse struct {
	ReleaseGroups []ReleaseGroup `json:"release-groups"`
}

type ReleaseGroup struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	FirstReleaseDate string `json:"first-release-date"`
	PrimaryType      string `json:"primary-type"`
}

var addCmd = &cobra.Command{
	Use:   `add "<artist name>"`,
	Short: "Search and add releases by artist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manual, _ := cmd.Flags().GetBool("manual")
		if manual {
			if len(args) != 0 {
				return fmt.Errorf("--manual takes no arguments")
			}
			return addManual()
		}
		if len(args) != 1 {
			return fmt.Errorf("requires artist name argument")
		}
		artistName := args[0]
		return addFromMusicBrainz(artistName)
	},
}

func searchArtists(query string) ([]Artist, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/artist?query=artist:%s&fmt=json", query)
	resp, err := makeRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ArtistSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Artists, nil
}

func searchReleaseGroups(artistID string) ([]ReleaseGroup, error) {
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release-group?artist=%s&fmt=json", artistID)
	resp, err := makeRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ReleaseGroupSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.ReleaseGroups, nil
}

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "musiccat/0.1.0 (robertjamespeacock@gmail.com)")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	time.Sleep(1 * time.Second) // Rate limit
	return resp, nil
}

func selectItem[T any](prompt string, items []T) (T, error) {
	var zero T
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(items) {
			fmt.Printf("Invalid selection. Enter a number between 1 and %d: ", len(items))
			continue
		}
		return items[num-1], nil
	}
	return zero, scanner.Err()
}

func selectMultipleItems[T any](prompt string, items []T) ([]T, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		parts := strings.Split(input, ",")
		var selected []T
		valid := true
		for _, part := range parts {
			part = strings.TrimSpace(part)
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 || num > len(items) {
				valid = false
				break
			}
			selected = append(selected, items[num-1])
		}
		if !valid {
			fmt.Printf("Invalid selection. Enter numbers between 1 and %d, comma-separated: ", len(items))
			continue
		}
		return selected, nil
	}
	return nil, scanner.Err()
}

func parseYear(date string) *int {
	if len(date) < 4 {
		return nil
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return nil
	}
	return &year
}

func addFromMusicBrainz(artistName string) error {
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
	artists, err := searchArtists(artistName)
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

	// Search release groups
	releaseGroups, err := searchReleaseGroups(selectedArtist.ID)
	if err != nil {
		return err
	}
	if len(releaseGroups) == 0 {
		return fmt.Errorf("no releases found for artist '%s'", selectedArtist.Name)
	}

	// Display and select release groups (multiple)
	fmt.Println("Releases found:")
	for i, rg := range releaseGroups {
		year := parseYear(rg.FirstReleaseDate)
		yearStr := ""
		if year != nil {
			yearStr = fmt.Sprintf(" (%d)", *year)
		}
		typeStr := ""
		if rg.PrimaryType != "" {
			typeStr = " [" + rg.PrimaryType + "]"
		}
		fmt.Printf("%d. %s%s%s\n", i+1, rg.Title, yearStr, typeStr)
	}
	selectedRGs, err := selectMultipleItems("Select releases (numbers, comma-separated): ", releaseGroups)
	if err != nil {
		return err
	}

	// Insert each selected release and ownership
	for _, rg := range selectedRGs {
		var releaseYear *int = parseYear(rg.FirstReleaseDate)
		releaseID, err := db.InsertRelease(database, selectedArtist.Name, rg.Title, releaseYear, &rg.ID)
		if err != nil {
			return err
		}
		_, err = db.InsertOwnership(database, releaseID, cfg.CurrentFormat, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Added %d releases to collection.\n", len(selectedRGs))
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
	formatCategory := promptValidFormat("Format category (CD, Vinyl, Tape, Digital): ")
	formatDetailInput := promptString("Format detail (optional): ")
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
	_, err = db.InsertOwnership(database, id, formatCategory, formatDetail, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Added release to collection.")
	return nil
}

func promptString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptOptionalInt(prompt string) *int {
	input := promptString(prompt)
	if input == "" {
		return nil
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid number, ignoring")
		return nil
	}
	return &num
}

func promptValidFormat(prompt string) string {
	for {
		input := promptString(prompt)
		for _, f := range ValidFormats {
			if strings.EqualFold(input, f) {
				return f
			}
		}
		fmt.Printf("Invalid format. Valid: %s\n", strings.Join(ValidFormats, ", "))
	}
}

func init() {
	addCmd.Flags().Bool("manual", false, "Manually enter release details")
	rootCmd.AddCommand(addCmd)
}
