package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Player struct {
	Name  string
	Goals int
}

type Game struct {
	ID       int
	TeamA    string
	TeamB    string
	ScoreA   int
	ScoreB   int
	ScorersA []Player
	ScorersB []Player
	Players  []Player
}

var (
	games  []Game
	gameID int
	mu     sync.Mutex
)

func addGame(teamA, teamB string) int {
	mu.Lock()
	defer mu.Unlock()
	gameID++
	newGame := Game{ID: gameID, TeamA: teamA, TeamB: teamB}
	games = append(games, newGame)
	return gameID
}

func validateTeamName(id int, teamName string) bool {
	mu.Lock()
	defer mu.Unlock()

	for _, game := range games {
		if game.ID == id && (game.TeamA == teamName || game.TeamB == teamName) {
			return true
		}
	}
	return false
}

func updateScore(id, scoreA, scoreB int) {
	mu.Lock()
	defer mu.Unlock()
	for i, game := range games {
		if game.ID == id {
			games[i].ScoreA = scoreA
			games[i].ScoreB = scoreB
			return
		}
	}
}

func addPlayerToGame(id int, player Player) {
	mu.Lock()
	defer mu.Unlock()
	for i, game := range games {
		if game.ID == id {
			games[i].Players = append(games[i].Players, player)
			return
		}
	}
}

func listGames() {
	mu.Lock()
	defer mu.Unlock()
	for _, game := range games {
		fmt.Printf("Game %d: %s vs %s | Score: %d-%d\n", game.ID, game.TeamA, game.TeamB, game.ScoreA, game.ScoreB)
		for _, player := range game.Players {
			fmt.Printf("  Player: %s | Goals: %d\n", player.Name, player.Goals)
		}
	}
}

func sumGoals(goals map[string]int) int {
	total := 0
	for _, g := range goals {
		total += g
	}
	return total
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n==============================")
		fmt.Println("Football Match Tracker")
		fmt.Println("==============================")
		fmt.Println("1. Add New Game")
		fmt.Println("2. Update Game Score")
		fmt.Println("3. Add Player Stats")
		fmt.Println("4. Show All Games")
		fmt.Println("5. Exit")
		fmt.Print("Choose an option (1-5): ")

		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())
		fmt.Println()

		switch choice {
		case "1":
			fmt.Println("Add New Game")
			fmt.Print("Enter Team A: ")
			scanner.Scan()
			teamA := strings.TrimSpace(scanner.Text())
			fmt.Print("Enter Team B: ")
			scanner.Scan()
			teamB := strings.TrimSpace(scanner.Text())
			gameID := addGame(teamA, teamB)

			goalsA := make(map[string]int)
			goalsB := make(map[string]int)

			fmt.Println("\nEnter all goals scored in the match (type 'x' to finish):")
			for {
				fmt.Print("Goal format (Team,Player): ")
				scanner.Scan()
				input := strings.TrimSpace(scanner.Text())

				if input == "x" {
					fmt.Println("All goals entered.")
					break
				}

				inputSplit := strings.Split(input, ",")
				if len(inputSplit) >= 2 {
					team := strings.TrimSpace(inputSplit[0])
					player := strings.TrimSpace(inputSplit[1])

					if !validateTeamName(gameID, team) {
						fmt.Println("Invalid team name. Try again.")
						continue
					}

					if team == teamA {
						goalsA[player]++
					} else if team == teamB {
						goalsB[player]++
					}

					fmt.Printf("Goal recorded - Team: %s | Player: %s | Total Goals: %d\n", team, player, goalsA[player]+goalsB[player])
				} else {
					fmt.Println("Invalid input format. Please use Team,Player.")
				}
			}

			for name, goals := range goalsA {
				addPlayerToGame(gameID, Player{Name: name, Goals: goals})
			}
			for name, goals := range goalsB {
				addPlayerToGame(gameID, Player{Name: name, Goals: goals})
			}

			updateScore(gameID, sumGoals(goalsA), sumGoals(goalsB))
			fmt.Printf("Game successfully added with ID: %d\n", gameID)

		case "2":
			fmt.Println("Update Game Score")
			fmt.Print("Enter Game ID: ")
			scanner.Scan()
			id, _ := strconv.Atoi(scanner.Text())
			fmt.Print("Enter Score for Team A: ")
			scanner.Scan()
			scoreA, _ := strconv.Atoi(scanner.Text())
			fmt.Print("Enter Score for Team B: ")
			scanner.Scan()
			scoreB, _ := strconv.Atoi(scanner.Text())
			updateScore(id, scoreA, scoreB)
			fmt.Println("Score updated.")

		case "3":
			fmt.Println("Add Player to Game")
			fmt.Print("Enter Game ID: ")
			scanner.Scan()
			id, _ := strconv.Atoi(scanner.Text())
			fmt.Print("Enter Player Name: ")
			scanner.Scan()
			name := scanner.Text()
			fmt.Print("Enter Goals Scored: ")
			scanner.Scan()
			goals, _ := strconv.Atoi(scanner.Text())
			addPlayerToGame(id, Player{Name: name, Goals: goals})
			fmt.Println("Player added.")

		case "4":
			fmt.Println("Game List:")
			listGames()

		case "5":
			fmt.Println("Exiting... Goodbye.")
			return

		default:
			fmt.Println("Invalid option, please choose between 1 and 5.")
		}
	}
}
