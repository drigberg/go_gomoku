package client

import (
	"bufio"
	"fmt"
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
	serverName    string
	gameID        int
	userID        string
	opponentID    string
	yourColor     string
	opponentColor string
	gameOver      bool
	connection    net.Conn
	yourTurn      bool
	messages      []types.Message
	turn          int
	board         map[string]map[string]bool
}

func (client *Client) printTurn() {
	var turnStr string

	if client.turn > 0 {
		turnStr = "Turn #" + strconv.Itoa(client.turn)
		if client.yourTurn {
			turnStr += ": You"
		} else {
			turnStr += ": Opponent"
		}
	} else {
		if client.gameOver {
			turnStr = "Game over!"
		} else {
			turnStr = "Waiting for player to join..."
		}
	}

	fmt.Println(turnStr)
}

func (client *Client) printMessages() {
	toPrint := client.messages
	length := len(client.messages)

	if length > 5 {
		toPrint = client.messages[length-6:]
	}

	for _, message := range toPrint {
		message.Print()
	}
}

func (client *Client) refreshScreen() {
	helpers.ClearScreen()

	client.printTurn()
	helpers.PrintBoard(client.board)
	client.printMessages()
}

func (client *Client) sendToServer(request types.Request) {
	data, err := helpers.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	client.connection.Write(data) // TODO: get error, handle
}

func (client *Client) addMessage(content string, author string) {
	message := types.Message{
		Content: content,
		Author:  author,
	}

	client.messages = append(client.messages, message)
	client.refreshScreen()
}

func (client *Client) createGame() {
	if client.gameID != 0 {
		fmt.Println("You're already in a game!")
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
		return
	}

	if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
		return
	}

	request := types.Request{
		GameID: client.gameID,
		UserID: client.userID,
		Action: constants.MOVE,
		Data:   text,
	}

	client.sendToServer(request)
}

func (client *Client) sendMessage(text string) {
	if client.gameID == 0 {
		client.addMessage("You're not in a game yet!", client.serverName)
		return
	}

	if client.turn == 0 {
		client.addMessage("The game hasn't started yet!", client.serverName)
		return
	}

	request := types.Request{
		GameID: client.gameID,
		UserID: client.userID,
		Action: constants.MESSAGE,
		Data:   text,
	}

	client.addMessage(request.Data, "You")
	client.sendToServer(request)
}

func (client *Client) getTurnOneInstructions() string {
	return "You go first! Begin by placing two black pieces and then one white. Ex: 'mv 8 8, 8 7, 6 6'"
}

func (client *Client) getTurnTwoInstructions() string {
	return "If you want to play white, play a move as normal. Otherwise, type 'mv pass'."
}

// Handler handles requests
func (client *Client) Handler(message []byte) {
	request := helpers.DecodeGob(message)

	switch action := request.Action; action {
	case constants.CREATE:
		if request.Success {
			client.gameOver = false
			gameIDStr := strconv.Itoa(request.GameID)

			client.addMessage("Created game #"+gameIDStr, client.serverName)

			client.gameID = request.GameID
			client.yourTurn = true
		} else {
			client.addMessage("Error! Could not create game.", client.serverName)
		}
	case constants.JOIN:
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
			fmt.Println(request.Data)
		}
	case constants.OTHERJOINED:
		if request.Success {
			client.turn = request.Turn
			client.opponentID = request.UserID

			client.yourTurn = request.YourTurn
			client.addMessage("Let the game begin!", client.serverName)
			if client.yourTurn {
				client.addMessage(client.getTurnOneInstructions(), client.serverName)
			}
		}
	case constants.MESSAGE:
		if request.Success {
			client.addMessage(request.Data, "Opponent")
		} else {
			client.addMessage("Error! Could not parse message from opponent.", client.serverName)
		}
	case constants.HOME:
		client.gameOver = false
		client.gameID = 0
		client.opponentID = ""
		client.yourColor = ""
		client.opponentColor = ""
		client.messages = []types.Message{}
		client.turn = 0
		client.board = make(map[string]map[string]bool)

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
			if client.turn == 2 {
				client.yourColor = request.Colors[client.userID]
				if client.yourColor == "white" {
					client.opponentColor = "black"
				} else {
					client.opponentColor = "white"
				}
			}

			client.turn = request.Turn
			client.board = request.Board

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
				client.addMessage(client.getTurnOneInstructions(), client.serverName)
			}
		} else {
			client.addMessage(request.Data, client.serverName)
		}
	}
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

func (client *Client) init() {
	client.board = make(map[string]map[string]bool)
}

// Run runs the client
func (client *Client) Run(host string, port string) {
	helpers.InitMaps()

	// create addresses
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	client.userID = uuid.String()

	fmt.Println("Connecting to host on port " + port + "...")
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
		fmt.Println("Could not connect!")
		os.Exit(1)
	}

	client.listenForInput()
}

// CreateClient creates a client instance
func CreateClient() Client {
	return Client{
		serverName: "GoGomoku",
	}
}
