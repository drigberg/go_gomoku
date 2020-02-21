package client

import (
	"bufio"
	"fmt"
	"go_gomoku/board"
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Client runs the CLI for players
type Client struct {
	DisablePrint  bool
	Messages      []types.Message
	serverName    string
	gameID        int
	userID        string
	opponentID    string
	yourColor     string
	opponentColor string
	gameOver      bool
	connection    net.Conn
	yourTurn      bool
	turn          int
	board         board.Board
}

// New creates a client instance
func New(serverName string) Client {
	return Client{
		serverName: serverName,
	}
}

func (client *Client) printString(message string) {
	if !client.DisablePrint {
		fmt.Println(message)
	}
}

func (client *Client) printError(err error) {
	if !client.DisablePrint {
		fmt.Println(err)
	}
}

func (client *Client) printTurn() {
	var turnStr string

	if client.turn == 0 {
		if client.gameOver {
			turnStr = "Game over!"
		} else {
			turnStr = "Waiting for player to join..."
		}
	} else {
		turnStr = "Turn #" + strconv.Itoa(client.turn)
		if client.yourTurn {
			turnStr += ": You"
		} else {
			turnStr += ": Opponent"
		}
	}

	client.printString(turnStr)
}

func (client *Client) printMessages() {
	toPrint := client.Messages
	length := len(client.Messages)
	if length > 5 {
		toPrint = client.Messages[len(client.Messages)-6:]
	}

	for _, message := range toPrint {
		message.Print()
	}
}

func (client *Client) refreshScreen() {
	if !client.DisablePrint {
		helpers.ClearScreen()
		client.printTurn()
		client.board.PrintBoard()
		client.printMessages()
	}
}

func (client *Client) sendToServer(request types.Request) {
	data, err := helpers.GobToBytes(request)

	if err != nil {
		client.printError(err)
		return
	}

	client.connection.Write(data) // TODO: get error, handle
}

func (client *Client) addMessage(content string, author string) {
	message := types.Message{
		Content: content,
		Author:  author,
	}

	client.Messages = append(client.Messages, message)
	client.refreshScreen()
}

func (client *Client) createGame() {
	if client.gameID != 0 {
		client.printString("You're already in a game!")
		return
	}

	request := types.Request{
		UserID: client.userID,
		Action: constants.CREATE,
	}

	client.sendToServer(request)
}

func (client *Client) joinGame(gameIDStr string) {
	if client.gameID != 0 {
		client.addMessage("You're already in a game!!", client.serverName)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		client.addMessage("Please enter a valid integer as the game id to join!", client.serverName)
		return
	}

	request := types.Request{
		GameID: gameID,
		UserID: client.userID,
		Action: constants.JOIN,
	}

	client.sendToServer(request)

}

func (client *Client) makeMove(text string) {
	if client.gameID == 0 {
		client.addMessage("You're not in a game yet!", client.serverName)
	} else if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
	} else {
		request := types.Request{
			GameID: client.gameID,
			UserID: client.userID,
			Action: constants.MOVE,
			Data:   text,
		}

		client.sendToServer(request)
	}
}

func (client *Client) sendMessage(text string) {
	if client.gameID == 0 {
		client.addMessage("You're not in a game yet!", client.serverName)
	} else if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
	} else {
		request := types.Request{
			GameID: client.gameID,
			UserID: client.userID,
			Action: constants.MESSAGE,
			Data:   text,
		}

		client.addMessage(request.Data, "You")
		client.sendToServer(request)
	}
}

func (client *Client) getTurnOneInstructions() string {
	return "You go first! Begin by placing two black pieces and then one white. Ex: 'mv 8 8, 8 7, 6 6'"
}

func (client *Client) getTurnTwoInstructions() string {
	return "If you want to play white, play a move as normal. Otherwise, type 'mv pass'."
}

func (client *Client) handleCreateRequest(request types.Request) {
	if request.Success {
		client.gameOver = false
		gameIDStr := strconv.Itoa(request.GameID)

		client.addMessage("Created game #"+gameIDStr, client.serverName)

		client.gameID = request.GameID
		client.yourTurn = true
	} else {
		client.addMessage("Error! Could not create game.", client.serverName)
	}
}

func (client *Client) handleJoinRequest(request types.Request) {
	if request.Success {
		client.gameOver = false
		client.gameID = request.GameID
		gameIDStr := strconv.Itoa(request.GameID)
		client.opponentID = request.UserID
		client.turn = request.Turn
		client.yourTurn = request.YourTurn

		client.addMessage("Joined game #"+gameIDStr, client.serverName)
		if client.yourTurn {
			client.addMessage(client.getTurnOneInstructions(), client.serverName)
		}
	} else {
		client.printString(request.Data)
	}
}

func (client *Client) handleOtherJoinedRequest(request types.Request) {
	if request.Success {
		client.turn = request.Turn
		client.opponentID = request.UserID

		client.yourTurn = request.YourTurn
		client.addMessage("Let the game begin!", client.serverName)
		if client.yourTurn {
			client.addMessage(client.getTurnOneInstructions(), client.serverName)
		}
	}
}

func (client *Client) handleMessageRequest(request types.Request) {
	if request.Success {
		client.addMessage(request.Data, "Opponent")
	} else {
		client.addMessage("Error! Could not parse message from opponent.", client.serverName)
	}
}

func (client *Client) handleHomeRequest(request types.Request) {
	client.gameOver = false
	client.gameID = 0
	client.opponentID = ""
	client.yourColor = ""
	client.opponentColor = ""
	client.Messages = []types.Message{}
	client.turn = 0
	client.board = board.New()

	helpers.ClearScreen()
	client.printString("WELCOME TO GOMOKU!")
	if len(request.Home) == 0 {
		client.printString("No open games! Type 'mk' to make a new game!")
	} else {
		client.printString("Open Games")
		client.printString("(type hm to refresh)")
		client.printString("_________")

		for _, game := range request.Home {
			client.printString("Game ID: " + strconv.Itoa(game.ID) + " ----- User: " + game.UserID)
		}
	}
}

func (client *Client) handleMoveRequest(request types.Request) {
	if request.Success {
		if client.turn == 2 {
			client.yourColor = request.Colors[client.userID]
			if client.yourColor == "white" {
				client.opponentColor = "black"
			} else {
				client.opponentColor = "white"
			}
		}

		client.turn = request.Turn
		client.board.Spaces = request.Board

		player := "You"
		if request.UserID == client.opponentID {
			player = "Opponent"
		}

		client.gameOver = request.GameOver
		client.yourTurn = request.YourTurn

		client.addMessage(request.Data, player)
		if client.gameOver {
			client.addMessage("Type hm to go back to the main screen!", client.serverName)
		}

		if client.yourTurn && client.turn == 2 {
			client.addMessage(client.getTurnTwoInstructions(), client.serverName)
		}
	} else {
		client.addMessage(request.Data, client.serverName)
	}
}

// Handler handles requests
func (client *Client) Handler(message []byte) {
	request := helpers.DecodeGob(message)

	switch action := request.Action; action {
	case constants.CREATE:
		client.handleCreateRequest(request)
	case constants.JOIN:
		client.handleJoinRequest(request)
	case constants.OTHERJOINED:
		client.handleOtherJoinedRequest(request)
	case constants.MESSAGE:
		client.handleMessageRequest(request)
	case constants.HOME:
		client.handleHomeRequest(request)
	case constants.MOVE:
		client.handleMoveRequest(request)
	}
	client.refreshScreen()
}

func (client *Client) backToHome() {
	request := types.Request{
		Action: constants.HOME,
	}

	client.sendToServer(request)
}

func (client *Client) listenForInput() {
	goingHome := false
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		if text == "y" && goingHome {
			goingHome = false
			client.backToHome()
			continue
		}

		// reset confirmation if user gives a different command
		goingHome = false

		if len(text) < 2 {
			continue
		}

		switch action := text[:2]; action {
		case "hp":
			client.addMessage("Type mk to make a game; jn <game_id> to join a game; mv <x> <y> to make a move; mg <message> to send a message; hp for help", client.serverName)
		case "mk":
			client.createGame()
		case "jn":
			if len(text) < 4 {
				client.addMessage("Invalid value! Type 'hp' for help!", client.serverName)
				continue
			}
			client.joinGame(text[3:])
		case "mg":
			if len(text) < 4 {
				client.addMessage("Invalid value! Type 'hp' for help!", client.serverName)
				continue
			}
			client.sendMessage(text[3:])
		case "mv":
			if len(text) < 4 {
				client.addMessage("Invalid value! Type 'hp' for help!", client.serverName)
				continue
			}
			client.makeMove(text[3:])
		case "hm":
			if client.gameOver || client.gameID == 0 {
				client.backToHome()
				continue
			}
			goingHome = true
			client.addMessage("Are you sure you want to leave the game? Type y if you DEFINITELY want to go back to the home screen.", client.serverName)
		default:
			client.addMessage("Unrecognized command! Type 'hp' for help!", client.serverName)
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

// Run begins the CLI and connects to the server
func (client *Client) Run(host string, port string) {
	// create addresses
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	client.userID = uuid.String()

	client.printString("Connecting to host on port " + port + "...")
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Fatal(err)
	}

	client.connection = conn
	socketClient := &types.SocketClient{Socket: client.connection}
	connected := make(chan bool)

	go socketClient.Receive(client.Handler, &connected)

	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		client.printString("Could not connect!")
		os.Exit(1)
	}

	client.listenForInput()
}
