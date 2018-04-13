
package client

import (
	"bufio"
	// "io/ioutil"
	"net"
	"os"
	"fmt"
	"strconv"
	"time"
	"github.com/google/uuid"
	"go_gomoku/util"
	"go_gomoku/types"
	"go_gomoku/helpers"
	"go_gomoku/constants"
)

var (
	serverName = "Gomoku"
	scanner = bufio.NewScanner(os.Stdin)
	goingHome = false
	gameId = 0
	userId string
	opponentId string
	yourColor string
	opponentColor string
	gameOver = false
	connection net.Conn
	yourTurn = false
	messages = []types.Message{}
	connected chan bool
	turn = 0
	board map[string]map[string]bool
	turnOneInstructions = "You go first! Begin by placing two black pieces and then one white. Ex: 'mv 8 8, 8 7, 6 6'"
	turnTwoInstructions = "If you want to play white, play a move as normal. Otherwise, type 'mv pass'."
)

func printTurn() {
	var turnStr string

	if (turn > 0) {
		turnStr = "Turn #" + strconv.Itoa(turn)
		if (yourTurn) {
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
		toPrint = messages[length - 6:]
	}

	for _, message := range(toPrint) {
		message.Print()
	}
}

func RefreshScreen() {
	util.CallClear()

	printTurn()
	helpers.PrintBoard(board)
	printMessages()
}

// Send request struct to server as byte array
func SendToServer(request types.Request) {
	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	connection.Write(data) // TODO: get error, handle
}

func AddMessage(content string, author string) {
	message := types.Message{
		Content: content,
		Author: author,
	}

	messages = append(messages, message)
	RefreshScreen()
}

func CreateGame() {
	if gameId != 0 {
		fmt.Println("You're already in a game!")
		return
	}

	request := types.Request{
		UserId: userId,
		Action: constants.CREATE,
	}

	SendToServer(request)
}

func JoinGame(gameIdStr string) {
	if gameId != 0 {
		AddMessage("You're already in a game!!", serverName)
		return
	}

	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		AddMessage("Please enter a valid integer as the game id to join!", serverName)
		return
	}

	request := types.Request{
		GameId: gameId,
		UserId: userId,
		Action: constants.JOIN,
	}

	SendToServer(request)

}

func MakeMove(text string) {
	if gameId == 0 {
		AddMessage("You're not in a game yet!", serverName)
		return
	}

	if turn == 0 {
		AddMessage("The game hasn't started yet!", serverName)
		return
	}

	request := types.Request{
		GameId: gameId,
		UserId: userId,
		Action: constants.MOVE,
		Data: text,
	}

	SendToServer(request)
}

func SendMessage(text string) {
	if gameId == 0 {
		AddMessage("You're not in a game yet!", serverName)
		return
	}

	if turn == 0 {
		AddMessage("The game hasn't started yet!", serverName)
		return
	}

	request := types.Request{
		GameId: gameId,
		UserId: userId,
		Action: constants.MESSAGE,
		Data: text,
	}

	AddMessage(request.Data, "You")
	SendToServer(request)
}

func Handler(message []byte) {
	request := util.DecodeGob(message)

	switch action := request.Action; action {
	case constants.CREATE:
		if request.Success {
			gameOver = false
			gameIdStr := strconv.Itoa(request.GameId)

			AddMessage("Created game #" + gameIdStr, serverName)

			gameId = request.GameId
			yourTurn = true
		} else {
			AddMessage("Error! Could not create game.", serverName)
		}
	case constants.JOIN:
			if request.Success {
				gameOver = false
				gameId = request.GameId
				gameIdStr := strconv.Itoa(request.GameId)
				opponentId = request.UserId
				turn = request.Turn
				yourTurn = request.YourTurn

				AddMessage("Joined game #" + gameIdStr, serverName)
				if yourTurn {
					AddMessage(turnOneInstructions, serverName)
				}
			} else {
				fmt.Println(request.Data)
			}
	case constants.OTHER_JOINED:
		if request.Success {
			turn = request.Turn
			opponentId = request.UserId

			yourTurn = request.YourTurn
			AddMessage("Let the game begin!", serverName)
			if yourTurn {
				AddMessage(turnOneInstructions, serverName)
			}
		}
	case constants.MESSAGE:
		if request.Success {
			AddMessage(request.Data, "Opponent")
		} else {
			AddMessage("Error! Could not parse message from opponent.", serverName)
		}
	case constants.HOME:
		go func() { connected <- true }()

		gameOver = false
		gameId = 0
		goingHome = false
		opponentId = ""
		yourColor = ""
		opponentColor = ""
		messages = []types.Message{}
		turn = 0
		board = make(map[string]map[string]bool)

		util.CallClear()
		fmt.Println("WELCOME TO GOMOKU!")
		if len(request.Home) == 0 {
			fmt.Println("No open games! Type 'mk' to make a new game!")
		} else {
			fmt.Println("Open Games")
			fmt.Println("(type hm to refresh)")
			fmt.Println("_________")

			for _, game := range(request.Home) {
				fmt.Println("Game Id: " + strconv.Itoa(game.Id) + " ----- User: " + game.UserId)
			}
		}


	case constants.MOVE:
		if request.Success {
			if turn == 2 {
				yourColor = request.Colors[userId]
				if yourColor == "white" {
					opponentColor = "black"
				} else {
					opponentColor = "white"
				}
			}

			turn = request.Turn
			board = request.Board

			player := "You"
			if request.UserId == opponentId {
				player = "Opponent"
			}

			gameOver = request.GameOver
			yourTurn = request.YourTurn

			AddMessage(request.Data, player)
			if gameOver {
				AddMessage("Type hm to go back to the main screen!", serverName)
			}

			if yourTurn && turn == 2 {
				AddMessage(turnTwoInstructions, serverName)
			}
		} else {
			AddMessage(request.Data, serverName)
		}
	}
}

func BackToHome() {
	request := types.Request{
		Action: constants.HOME,
	}

	SendToServer(request)
}

func ListenForInput() {
	for scanner.Scan() {
		text := scanner.Text()

		if text == "y" && goingHome {
			BackToHome()
			continue
		}

		// reset confirmation if user gives a different command
		goingHome = false

		if len(text) < 2 {
			continue
		}

		switch action := text[:2]; action {
		case "hp":
			AddMessage("Type mk to make a game; jn <game_id> to join a game; mv <x> <y> to make a move; mg <message> to send a message; hp for help", serverName)
		case "mk":
			CreateGame()
		case "jn":
			if len(text) < 4 {
				AddMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			JoinGame(text[3:])
		case "mg":
			if len(text) < 4 {
				AddMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			SendMessage(text[3:])
		case "mv":
			if len(text) < 4 {
				AddMessage("Invalid value! Type 'hp' for help!", serverName)
				continue
			}

			MakeMove(text[3:])
		case "hm":
			if gameOver || gameId == 0 {
				BackToHome()
				continue
			}

			goingHome = true
			AddMessage("Are you sure you want to leave the game? Type y if you DEFINITELY want to go back to the home screen.", serverName)
		default:
			AddMessage("Unrecognized command! Type 'hp' for help!", serverName)
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

func Run(host *string, port *int) {
	connected = make(chan bool)
	helpers.InitMaps()

	// create addresses
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
	}
	userId = uuid.String()

	fmt.Println("Connecting to host on port " + strconv.Itoa(*port) + "...")

	conn, err := net.Dial("tcp", *host + ":" + strconv.Itoa(*port))
	if err != nil {
			fmt.Println(err)
	}

	connection = conn

	client := &types.Client{Socket: connection}

	go client.Receive(Handler)

	board = make(map[string]map[string]bool)

	select {
	case <- connected:
	case <- time.After(5 * time.Second):
		fmt.Println("Could not connect!")
		os.Exit(1)
	}

	ListenForInput()
}
