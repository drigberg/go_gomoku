package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Client runs the CLI for players
type Client struct {
	disablePrint  	bool
	handledRequests chan Request
	messages      	[]Message
	serverName    	string
	GameID        	int
	userID        	string
	opponentID    	string
	yourColor     	string
	opponentColor 	string
	gameOver      	bool
	connection    	net.Conn
	yourTurn      	bool
	turn          	int
	board         	Board
}

// Interface defines methods a Client should implement
type ClientInterface interface {
	Run(string, string)
	handler([]byte)
	createGame()
	listenForInput(io.Reader)
	addMessage(string, string)
	backToHome()
	clearScreen()
	getTurnOneInstructions() string
	getTurnTwoInstructions() string
	handleCreateRequest(Request)
	handleHomeRequest(Request)
	handleJoinRequest(Request)
	handleMessageRequest(Request)
	handleMoveRequest(Request)
	handleOtherJoinedRequest(Request)
	joinGame(string)
	makeMove(string)
	printBoard()
	printBoardAndMessages()
	printError(err error)
	printHomeScreen(Request)
	printMessages()
	printString(message string)
	printTurn()
	sendMessage(string)
	sendToServer(Request)
}

// assert that Board implements Interface
var _ ClientInterface = (*Client)(nil)

// New creates a client instance
func NewClient(serverName string) Client {
	client := Client{
		serverName: serverName,
		handledRequests: make(chan Request),
	}
	client.reset()
	return client
}

func (client *Client) reset() {
	client.GameID = -1
	client.gameOver = false
	client.opponentID = ""
	client.yourColor = ""
	client.opponentColor = ""
	client.messages = []Message{}
	client.turn = 0
	client.board = NewBoard()
}

func (client *Client) clearScreen() {
	if !client.disablePrint {
		clearScreen()
	}
}

func (client *Client) printBoard() {
	if !client.disablePrint {
		client.board.printBoard()
	}
}

func (client *Client) printString(message string) {
	if !client.disablePrint {
		fmt.Println(message)
	}
}

func (client *Client) printError(err error) {
	if !client.disablePrint {
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
	toPrint := client.messages
	length := len(client.messages)
	if length > 5 {
		toPrint = client.messages[len(client.messages)-6:]
	}

	for _, message := range toPrint {
		client.printString(message.Author + ": " + message.Content)
	}
}

func (client *Client) printHomeScreen(request Request) {
	client.clearScreen()
	client.printString("WELCOME TO GOMOKU!")
	client.printString("Type 'mk' to make a new game")
	client.printString("Type 'hm' to refresh")
	client.printString("Type 'jn' followed by a game id to join a game")
	client.printString("_________")
	if len(request.Home) == 0 {
		client.printString("(no open games)")
	} else {
		for _, game := range request.Home {
			client.printString("Game ID: " + strconv.Itoa(game.ID) + " ----- User: " + game.UserID)
		}
	}
}

func (client *Client) printBoardAndMessages() {
	client.clearScreen()
	client.printTurn()
	client.printBoard()
	client.printMessages()
}

func (client *Client) sendToServer(request Request) {
	data, err := gobToBytes(request)

	if err != nil {
		client.printError(err)
		return
	}

	client.connection.Write(data) // TODO: get error, handle
}

func (client *Client) addMessage(content string, author string) {
	message := Message{
		Content: content,
		Author:  author,
	}

	client.messages = append(client.messages, message)
	client.printBoardAndMessages()
}

func (client *Client) createGame() {
	if client.GameID != -1 {
		client.printString("You're already in a game!")
		return
	}

	request := Request{
		UserID: client.userID,
		Action: CREATE,
	}

	client.sendToServer(request)
}

func (client *Client) joinGame(gameIDStr string) {
	if client.GameID != -1 {
		client.addMessage("You're already in a game!!", client.serverName)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		client.addMessage("Please enter a valid integer as the game id to join!", client.serverName)
		return
	}

	request := Request{
		GameID: gameID,
		UserID: client.userID,
		Action: JOIN,
	}

	client.sendToServer(request)

}

func (client *Client) makeMove(text string) {
	if client.GameID == -1 {
		client.addMessage("You're not in a game yet!", client.serverName)
	} else if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
	} else {
		request := Request{
			GameID: client.GameID,
			UserID: client.userID,
			Action: MOVE,
			Data:   text,
		}

		client.sendToServer(request)
	}
}

func (client *Client) sendMessage(text string) {
	if client.GameID == -1 {
		client.addMessage("You're not in a game yet!", client.serverName)
	} else if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
	} else {
		request := Request{
			GameID: client.GameID,
			UserID: client.userID,
			Action: MESSAGE,
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

func (client *Client) handleCreateRequest(request Request) {
	if request.Success {
		client.gameOver = false
		client.GameID = request.GameID
		client.yourTurn = true
		gameIDStr := strconv.Itoa(request.GameID)
		client.addMessage("Created game #"+gameIDStr, client.serverName)
	} else {
		client.addMessage("Error! Could not create game.", client.serverName)
	}
}

func (client *Client) handleJoinRequest(request Request) {
	if request.Success {
		client.gameOver = false
		client.GameID = request.GameID
		client.opponentID = request.UserID
		client.turn = request.Turn
		client.yourTurn = request.YourTurn
		gameIDStr := strconv.Itoa(request.GameID)
		client.addMessage("Joined game #"+gameIDStr, client.serverName)
		if client.yourTurn {
			client.addMessage(client.getTurnOneInstructions(), client.serverName)
		}
	} else {
		client.printString(request.Data)
	}
}

func (client *Client) handleOtherJoinedRequest(request Request) {
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

func (client *Client) handleMessageRequest(request Request) {
	if request.Success {
		client.addMessage(request.Data, "Opponent")
	} else {
		client.addMessage("Error! Could not parse message from opponent.", client.serverName)
	}
}

func (client *Client) handleHomeRequest(request Request) {
	client.reset()
	client.printHomeScreen(request)
}

func (client *Client) handleMoveRequest(request Request) {
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

// handler handles requests
func (client *Client) handler(message []byte) {
	request := decodeGob(message)

	switch action := request.Action; action {
	case CREATE:
		client.handleCreateRequest(request)
	case JOIN:
		client.handleJoinRequest(request)
	case OTHERJOINED:
		client.handleOtherJoinedRequest(request)
	case MESSAGE:
		client.handleMessageRequest(request)
	case HOME:
		client.handleHomeRequest(request)
	case MOVE:
		client.handleMoveRequest(request)
	}
	go func() {client.handledRequests <- request}()
}

func (client *Client) backToHome() {
	request := Request{
		Action: HOME,
	}

	client.sendToServer(request)
}

func (client *Client) listenForInput(readstream io.Reader) {
	goingHome := false
	scanner := bufio.NewScanner(readstream)
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
			if client.gameOver || client.GameID == -1 {
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

func (client *Client) Connect(host string, port string) *SocketClient {
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
	socketClient := &SocketClient{Socket: client.connection}
	return socketClient
}

// Run begins the CLI and connects to the server
func (client *Client) Run(host string, port string) {
	socketClient := client.Connect(host, port)
	connected := make(chan bool)

	go socketClient.Receive(client.handler, &connected)

	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		client.printString("Could not connect!")
		os.Exit(1)
	}

	client.listenForInput(os.Stdin)
}
