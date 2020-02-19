package client

import (
	"bufio"
	// "io/ioutil"
	"fmt"
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	serverName          = "Gomoku"
	scanner             = bufio.NewScanner(os.Stdin)
	goingHome           = false
	gameID              = 0
	userID              string
	opponentID          string
	yourColor           string
	opponentColor       string
	gameOver            = false
	connection          net.Conn
	yourTurn            = false
	messages            = []types.Message{}
	connected           chan bool
	turn                = 0
	board               map[string]map[string]bool
	turnOneInstructions = "You go first! Begin by placing two black pieces and then one white. Ex: 'mv 8 8, 8 7, 6 6'"
	turnTwoInstructions = "If you want to play white, play a move as normal. Otherwise, type 'mv pass'."
)

func printTurn() {
	var turnStr string

	if turn > 0 {
		turnStr = "Turn #" + strconv.Itoa(turn)
		if yourTurn {
			turnStr += ": You"
		} else {
			turnStr += ": Opponent"
		}
	} else {
		if gameOver {
			turnStr = "Game over!"
		} else {
			turnStr = "Waiting for player to join..."
		}
	}

	fmt.Println(turnStr)
}

func printMessages() {
	toPrint := messages
	length := len(messages)

	if length > 5 {
		toPrint = messages[length-6:]
	}

	for _, message := range toPrint {
		message.Print()
	}
}

func refreshScreen() {
	helpers.ClearScreen()

	printTurn()
	helpers.PrintBoard(board)
	printMessages()
}

func sendToServer(request types.Request) {
	data, err := helpers.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	connection.Write(data) // TODO: get error, handle
}

func addMessage(content string, author string) {
	message := types.Message{
		Content: content,
		Author:  author,
	}

	messages = append(messages, message)
	refreshScreen()
}

func createGame() {
	if gameID != 0 {
		fmt.Println("You're already in a game!")
		return
	}

	request := types.Request{
		UserID: userID,
		Action: constants.CREATE,
	}

	sendToServer(request)
}

func joinGame(gameIDStr string) {
	if gameID != 0 {
		addMessage("You're already in a game!!", serverName)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		addMessage("Please enter a valid integer as the game id to join!", serverName)
		return
	}

	request := types.Request{
		GameID: gameID,
		UserID: userID,
		Action: constants.JOIN,
	}

	sendToServer(request)

}

func makeMove(text string) {
	if gameID == 0 {
		addMessage("You're not in a game yet!", serverName)
		return
	}

	if turn == 0 {
		addMessage("The game hasn't started yet!", serverName)
		return
	}

	request := types.Request{
		GameID: gameID,
		UserID: userID,
		Action: constants.MOVE,
		Data:   text,
	}

	sendToServer(request)
}

func sendMessage(text string) {
	if gameID == 0 {
		addMessage("You're not in a game yet!", serverName)
		return
	}

	if turn == 0 {
		addMessage("The game hasn't started yet!", serverName)
		return
	}

	request := types.Request{
		GameID: gameID,
		UserID: userID,
		Action: constants.MESSAGE,
		Data:   text,
	}

	addMessage(request.Data, "You")
	sendToServer(request)
}

// Handler handles requests
func Handler(message []byte) {
	request := helpers.DecodeGob(message)

	switch action := request.Action; action {
	case constants.CREATE:
		if request.Success {
			gameOver = false
			gameIDStr := strconv.Itoa(request.GameID)

			addMessage("Created game #"+gameIDStr, serverName)

			gameID = request.GameID
			yourTurn = true
		} else {
			addMessage("Error! Could not create game.", serverName)
		}
	case constants.JOIN:
		if request.Success {
			gameOver = false
			gameID = request.GameID
			gameIDStr := strconv.Itoa(request.GameID)
			opponentID = request.UserID
			turn = request.Turn
			yourTurn = request.YourTurn

			addMessage("Joined game #"+gameIDStr, serverName)
			if yourTurn {
				addMessage(turnOneInstructions, serverName)
			}
		} else {
			fmt.Println(request.Data)
		}
	case constants.OTHERJOINED:
		if request.Success {
			turn = request.Turn
			opponentID = request.UserID

			yourTurn = request.YourTurn
			addMessage("Let the game begin!", serverName)
			if yourTurn {
				addMessage(turnOneInstructions, serverName)
			}
		}
	case constants.MESSAGE:
		if request.Success {
			addMessage(request.Data, "Opponent")
		} else {
			addMessage("Error! Could not parse message from opponent.", serverName)
		}
	case constants.HOME:
		go func() { connected <- true }()

		gameOver = false
		gameID = 0
		goingHome = false
		opponentID = ""
		yourColor = ""
		opponentColor = ""
		messages = []types.Message{}
		turn = 0
		board = make(map[string]map[string]bool)

		helpers.ClearScreen()
		fmt.Println("WELCOME TO GOMOKU!")
		if len(request.Home) == 0 {
			fmt.Println("No open games! Type 'mk' to make a new game!")
		} else {
			fmt.Println("Open Games")
			fmt.Println("(type hm to refresh)")
			fmt.Println("_________")

			for _, game := range request.Home {
				fmt.Println("Game ID: " + strconv.Itoa(game.ID) + " ----- User: " + game.UserID)
			}
		}

	case constants.MOVE:
		if request.Success {
			if turn == 2 {
				yourColor = request.Colors[userID]
				if yourColor == "white" {
					opponentColor = "black"
				} else {
					opponentColor = "white"
				}
			}

			turn = request.Turn
			board = request.Board

			player := "You"
			if request.UserID == opponentID {
				player = "Opponent"
			}

			gameOver = request.GameOver
			yourTurn = request.YourTurn

			addMessage(request.Data, player)
			if gameOver {
				addMessage("Type hm to go back to the main screen!", serverName)
			}

			if yourTurn && turn == 2 {
				addMessage(turnTwoInstructions, serverName)
			}
		} else {
			addMessage(request.Data, serverName)
		}
	}
}

func backToHome() {
	request := types.Request{
		Action: constants.HOME,
	}

	sendToServer(request)
}

func listenForInput() {
	for scanner.Scan() {
		text := scanner.Text()

		if text == "y" && goingHome {
			backToHome()
			continue
		}

		// reset confirmation if user gives a different command
		goingHome = false

		if len(text) < 2 {
			continue
		}

		switch action := text[:2]; action {
		case "hp":
			addMessage("Type mk to make a game; jn <game_id> to join a game; mv <x> <y> to make a move; mg <message> to send a message; hp for help", serverName)
		case "mk":
			createGame()
		case "jn":
			if len(text) < 4 {
				addMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			joinGame(text[3:])
		case "mg":
			if len(text) < 4 {
				addMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			sendMessage(text[3:])
		case "mv":
			if len(text) < 4 {
				addMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			makeMove(text[3:])
		case "hm":
			if gameOver || gameID == 0 {
				backToHome()
				continue
			}

			goingHome = true
			addMessage("Are you sure you want to leave the game? Type y if you DEFINITELY want to go back to the home screen.", serverName)
		default:
			addMessage("Unrecognized command! Type 'hp' for help!", serverName)
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

// Run runs the client
func Run(host string, port string) {
	connected = make(chan bool)
	helpers.InitMaps()

	// create addresses
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
	}
	userID = uuid.String()

	fmt.Println("Connecting to host on port " + port + "...")

	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Println(err)
	}

	connection = conn

	client := &types.Client{Socket: connection}

	go client.Receive(Handler)

	board = make(map[string]map[string]bool)

	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		fmt.Println("Could not connect!")
		os.Exit(1)
	}

	listenForInput()
}
