package main

import (
	_ "embed"
	"fmt"
	_ "image/png"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.png
var iconData []byte

type Player struct {
	Name  string
	Goals int
}

type Game struct {
	ID      int
	TeamA   string
	TeamB   string
	ScoreA  int
	ScoreB  int
	Players []Player
}

var (
	games  []Game
	gameID int
	mu     sync.Mutex
)

func addGame(teamA, teamB string, players []Player, scoreA, scoreB int) int {
	mu.Lock()
	defer mu.Unlock()
	gameID++
	newGame := Game{
		ID:      gameID,
		TeamA:   teamA,
		TeamB:   teamB,
		ScoreA:  scoreA,
		ScoreB:  scoreB,
		Players: players,
	}
	games = append(games, newGame)
	return gameID
}

func listGamesUI() *fyne.Container {
	mu.Lock()
	defer mu.Unlock()

	list := container.NewVBox()
	for _, g := range games {
		header := widget.NewLabel(fmt.Sprintf("Game %d: %s vs %s | Score: %d-%d", g.ID, g.TeamA, g.TeamB, g.ScoreA, g.ScoreB))
		list.Add(header)
		for _, p := range g.Players {
			list.Add(widget.NewLabel(fmt.Sprintf("  Player: %s | Goals: %d", p.Name, p.Goals)))
		}
		list.Add(widget.NewSeparator())
	}
	return container.NewStack(container.NewVScroll(list))
}

func showAddGamePopup(win fyne.Window) {
	teamAEntry := widget.NewEntry()
	teamAEntry.SetPlaceHolder("Team A")

	teamBEntry := widget.NewEntry()
	teamBEntry.SetPlaceHolder("Team B")

	scoreAEntry := widget.NewEntry()
	scoreAEntry.SetPlaceHolder("Score A")

	scoreBEntry := widget.NewEntry()
	scoreBEntry.SetPlaceHolder("Score B")

	playerName := widget.NewEntry()
	playerGoals := widget.NewEntry()
	players := []Player{}
	playerListLabel := widget.NewLabel("No players added yet.")

	addPlayerBtn := widget.NewButton("Add Player", func() {
		name := playerName.Text
		goals, err := strconv.Atoi(playerGoals.Text)
		if name == "" || err != nil {
			dialog.ShowError(fmt.Errorf("invalid player input"), win)
			return
		}
		players = append(players, Player{Name: name, Goals: goals})
		playerName.SetText("")
		playerGoals.SetText("")
		updatePlayerListLabel(players, playerListLabel)
	})

	form := container.NewVBox(
		teamAEntry,
		teamBEntry,
		scoreAEntry,
		scoreBEntry,
		widget.NewSeparator(),
		widget.NewLabel("Add Player:"),
		playerName,
		playerGoals,
		addPlayerBtn,
		playerListLabel,
	)

	dialog.ShowCustomConfirm("Add New Game", "Add", "Cancel", form, func(confirm bool) {
		if confirm {
			teamA := teamAEntry.Text
			teamB := teamBEntry.Text
			scoreA, errA := strconv.Atoi(scoreAEntry.Text)
			scoreB, errB := strconv.Atoi(scoreBEntry.Text)

			if teamA == "" || teamB == "" || errA != nil || errB != nil {
				dialog.ShowError(fmt.Errorf("invalid game data"), win)
				return
			}

			addGame(teamA, teamB, players, scoreA, scoreB)
			dialog.ShowInformation("Success", "Game added successfully!", win)
		}
	}, win)
}

func updatePlayerListLabel(players []Player, label *widget.Label) {
	if len(players) == 0 {
		label.SetText("No players added yet.")
		return
	}
	text := "Players:\n"
	for _, p := range players {
		text += fmt.Sprintf("- %s (%d goals)\n", p.Name, p.Goals)
	}
	label.SetText(text)
}

func showGameListPopup(win fyne.Window) {
	content := listGamesUI()
	dialog.ShowCustom("All Games", "Close", content, win)
}

func main() {
	a := app.New()
	w := a.NewWindow("Football Match Tracker")

	icon := fyne.NewStaticResource("icon.png", iconData)
	w.SetIcon(icon)
	w.SetTitle("Football Match Tracker")
	w.Resize(fyne.NewSize(600, 400))

	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		),
		fyne.NewMenu("Actions",
			fyne.NewMenuItem("Add Game", func() {
				showAddGamePopup(w)
			}),
			fyne.NewMenuItem("View Games", func() {
				showGameListPopup(w)
			}),
		),
	)

	w.SetMainMenu(menu)

	home := container.NewCenter(widget.NewLabel("Welcome to Football Match Tracker!\nUse the menu to perform actions."))
	w.SetContent(home)
	w.ShowAndRun()
}
