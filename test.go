package main

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.png
var iconData []byte

const csvFileName = "soccer_games.csv"

// Struct to manage cards and their containers
type GameCard struct {
	Card      *widget.Card
	Container *fyne.Container
}

// Add a card to the collection
func addCardToCollection(collection *[]GameCard, card GameCard) {
	*collection = append(*collection, card)
}

// Remove a card from the collection
func removeCardFromCollection(collection *[]GameCard, cardContainer *fyne.Container) {
	for i, gameCard := range *collection {
		if gameCard.Container == cardContainer {
			*collection = append((*collection)[:i], (*collection)[i+1:]...)
			break
		}
	}
}

// Save cards to a CSV file
func saveCardsToCSV(cards []GameCard) error {
	file, err := os.Create(csvFileName) // Creates the file if it does not exist
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"TeamA", "TeamB", "ScoreA", "ScoreB", "Comments", "GameDate"})

	// Write card data
	for _, gameCard := range cards {
		card := gameCard.Card
		title := card.Title
		subtitle := card.Subtitle
		content := card.Content.(*widget.Label).Text

		// Parse title and subtitle
		var teamA, teamB string
		var scoreA, scoreB, gameDate string

		// Split the title by " vs " to get team names
		teams := strings.Split(title, " vs ")
		if len(teams) == 2 {
			teamA = teams[0]
			teamB = teams[1]
		}

		// Split the subtitle by " - " and "| Date: " to get scores and date
		subtitleParts := strings.Split(subtitle, "| Date: ")
		if len(subtitleParts) == 2 {
			scores := strings.Split(strings.TrimPrefix(subtitleParts[0], "Score: "), " - ")
			if len(scores) == 2 {
				scoreA = scores[0]
				scoreB = scores[1]
			}
			gameDate = strings.TrimSpace(subtitleParts[1])
		}

		writer.Write([]string{teamA, teamB, scoreA, scoreB, content, gameDate})
	}

	return nil
}

// Load cards from a CSV file
func loadCardsFromCSV(win fyne.Window, collection *[]GameCard, cardContainer *fyne.Container) error {
	file, err := os.Open(csvFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file exists yet, so nothing to load
		}
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// Skip header row
	for _, record := range records[1:] {
		// Ensure the record has exactly 6 fields
		if len(record) != 6 {
			fmt.Printf("Skipping invalid record: %v\n", record)
			continue
		}

		teamA := record[0]
		teamB := record[1]
		scoreA := record[2]
		scoreB := record[3]
		comments := record[4]
		gameDate := record[5]

		// Create a card for each record
		gameCard := createGameCardFromData(win, collection, teamA, teamB, scoreA, scoreB, comments, gameDate)

		// Add the card's container to the cardContainer
		cardContainer.Add(gameCard.Container)
	}

	// Sort cards by date
	sortCardsByDate(collection)

	// Refresh the card container
	if cardContainer != nil {
		cardContainer.Refresh()
	}

	return nil
}

// Create a card from CSV data
func createGameCardFromData(win fyne.Window, collection *[]GameCard, teamA, teamB, scoreA, scoreB, comments, gameDate string) GameCard {
	card := widget.NewCard(
		fmt.Sprintf("%s vs %s", teamA, teamB),
		fmt.Sprintf("Score: %s - %s | Date: %s", scoreA, scoreB, gameDate),
		widget.NewLabel(comments),
	)
	card.Resize(fyne.NewSize(300, 80)) // Adjust the size to make it smaller

	// Wrap the card in a tappable container
	tappableCardContainer := createTappableCard(win, collection, card, container.NewMax(card))

	// Create the GameCard object
	gameCard := GameCard{Card: card, Container: tappableCardContainer}

	// Add the GameCard to the collection
	addCardToCollection(collection, gameCard)

	return gameCard
}

// Function to create the edit card dialog
func createEditCardDialog(card *widget.Card) (*widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry, fyne.CanvasObject) {
	teamAEntry := widget.NewEntry()
	teamBEntry := widget.NewEntry()
	scoreAEntry := widget.NewEntry()
	scoreBEntry := widget.NewEntry()
	commentsEntry := widget.NewMultiLineEntry()
	dateEntry := widget.NewEntry()

	// Parse existing card data
	titleParts := strings.Split(card.Title, " vs ")
	if len(titleParts) == 2 {
		teamAEntry.SetText(titleParts[0])
		teamBEntry.SetText(titleParts[1])
	}

	subtitleParts := strings.Split(card.Subtitle, "| Date: ")
	if len(subtitleParts) == 2 {
		scoreParts := strings.Split(strings.TrimPrefix(subtitleParts[0], "Score: "), " - ")
		if len(scoreParts) == 2 {
			scoreAEntry.SetText(scoreParts[0])
			scoreBEntry.SetText(scoreParts[1])
		}
		dateEntry.SetText(strings.TrimSpace(subtitleParts[1]))
	}

	commentsEntry.SetText(card.Content.(*widget.Label).Text)

	form := container.NewVBox(
		widget.NewLabel("Team A:"),
		teamAEntry,
		widget.NewLabel("Team B:"),
		teamBEntry,
		widget.NewLabel("Score A:"),
		scoreAEntry,
		widget.NewLabel("Score B:"),
		scoreBEntry,
		widget.NewLabel("Game Date (YYYY-MM-DD):"),
		dateEntry,
		widget.NewLabel("Comments:"),
		commentsEntry,
	)

	return teamAEntry, teamBEntry, scoreAEntry, scoreBEntry, commentsEntry, dateEntry, form
}

// Helper function to sort cards by date
func sortCardsByDate(collection *[]GameCard) {
	sort.SliceStable(*collection, func(i, j int) bool {
		cardI := (*collection)[i].Card
		cardJ := (*collection)[j].Card

		// Extract dates from subtitles
		dateI := extractDateFromSubtitle(cardI.Subtitle)
		dateJ := extractDateFromSubtitle(cardJ.Subtitle)

		// Parse dates
		parsedDateI, errI := time.Parse("2006-01-02", dateI)
		parsedDateJ, errJ := time.Parse("2006-01-02", dateJ)

		// Handle parsing errors
		if errI != nil {
			fmt.Printf("Error parsing date for card %s: %v\n", cardI.Title, errI)
			return false
		}
		if errJ != nil {
			fmt.Printf("Error parsing date for card %s: %v\n", cardJ.Title, errJ)
			return true
		}

		// Sort in ascending order
		return parsedDateI.Before(parsedDateJ)
	})
}

// Helper function to extract the date from a card's subtitle
func extractDateFromSubtitle(subtitle string) string {
	parts := strings.Split(subtitle, "| Date: ")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// Function to create a new game card
func createGameCard(win fyne.Window, collection *[]GameCard, cardContainer *fyne.Container) {
	teamAEntry := widget.NewEntry()
	teamBEntry := widget.NewEntry()
	scoreAEntry := widget.NewEntry()
	scoreBEntry := widget.NewEntry()
	commentsEntry := widget.NewMultiLineEntry()
	dateEntry := widget.NewEntry()

	// Set placeholders for better user experience
	teamAEntry.SetPlaceHolder("Enter Team A")
	teamBEntry.SetPlaceHolder("Enter Team B")
	scoreAEntry.SetPlaceHolder("Enter Score for Team A")
	scoreBEntry.SetPlaceHolder("Enter Score for Team B")
	commentsEntry.SetPlaceHolder("Enter Comments")
	dateEntry.SetPlaceHolder("YYYY-MM-DD")

	// Create a vertical layout for the input fields
	form := container.NewVBox(
		widget.NewLabel("Team A:"),
		teamAEntry,
		widget.NewLabel("Team B:"),
		teamBEntry,
		widget.NewLabel("Score A:"),
		scoreAEntry,
		widget.NewLabel("Score B:"),
		scoreBEntry,
		widget.NewLabel("Game Date (YYYY-MM-DD):"),
		dateEntry,
		widget.NewLabel("Comments:"),
		commentsEntry,
	)

	dialog.ShowCustomConfirm("New Game Card", "Add", "Cancel", form, func(confirmed bool) {
		if confirmed {
			// Create the card
			card := widget.NewCard(
				fmt.Sprintf("%s vs %s", teamAEntry.Text, teamBEntry.Text),
				fmt.Sprintf("Score: %s - %s | Date: %s", scoreAEntry.Text, scoreBEntry.Text, dateEntry.Text),
				widget.NewLabel(commentsEntry.Text),
			)
			card.Resize(fyne.NewSize(300, 80))

			// Wrap the card in a tappable container
			tappableCardContainer := createTappableCard(win, collection, card, container.NewMax(card))

			// Add the card to the collection and container
			newCard := GameCard{Card: card, Container: tappableCardContainer}
			addCardToCollection(collection, newCard)
			cardContainer.Add(newCard.Container)
			cardContainer.Refresh()

			// Save cards to CSV
			err := saveCardsToCSV(*collection)
			if err != nil {
				dialog.ShowError(err, win)
			}
		}
	}, win)
}

// Function to create a tappable card with edit and delete buttons
func createTappableCard(win fyne.Window, collection *[]GameCard, card *widget.Card, cardContainer *fyne.Container) *fyne.Container {
	// Edit button
	editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		if cardContainer != nil {
			teamAEntry, teamBEntry, scoreAEntry, scoreBEntry, commentsEntry, dateEntry, form := createEditCardDialog(card)

			dialog.ShowCustomConfirm("Edit Card", "Save", "Cancel", form, func(confirmed bool) {
				if confirmed {
					card.SetTitle(fmt.Sprintf("%s vs %s", teamAEntry.Text, teamBEntry.Text))
					card.SetSubTitle(fmt.Sprintf("Score: %s - %s | Date: %s", scoreAEntry.Text, scoreBEntry.Text, dateEntry.Text))
					card.SetContent(widget.NewLabel(commentsEntry.Text))
					card.Refresh()
					cardContainer.Refresh()

					err := saveCardsToCSV(*collection)
					if err != nil {
						dialog.ShowError(err, win)
					}
				}
			}, win)
		}
	})

	// Delete button
	deleteButton := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		if cardContainer != nil {
			removeCardFromCollection(collection, cardContainer)
			cardContainer.Objects = nil // Clear the objects inside the cardContainer
			cardContainer.Refresh()
			err := saveCardsToCSV(*collection)
			if err != nil {
				dialog.ShowError(err, win)
			}
		}
	})

	// Create a horizontal layout for the buttons
	buttons := container.NewHBox(editButton, deleteButton)

	// Return a container with the card and buttons
	return container.NewBorder(nil, buttons, nil, nil, card)
}

func main() {
	fmt.Println("Starting the application...")
	a := app.New()
	a.SetIcon(fyne.NewStaticResource("icon.png", iconData)) // Set the application icon
	w := a.NewWindow("Soccer Game Tracker")
	w.Resize(fyne.NewSize(400, 600))

	var collection []GameCard
	cardContainer := container.NewVBox()

	// Load cards from CSV on launch
	err := loadCardsFromCSV(w, &collection, cardContainer)
	if err != nil {
		dialog.ShowError(err, w)
	}

	// Make the collection scrollable
	scrollableContainer := container.NewScroll(cardContainer)

	plusBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		createGameCard(w, &collection, cardContainer)
	})

	content := container.NewBorder(nil, plusBtn, nil, nil, scrollableContainer)
	w.SetContent(content)

	w.ShowAndRun()
}
