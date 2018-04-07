
package client

import (
	"bufio"
	// "io/ioutil"
	"net"
	"os"
	"fmt"
	"strconv"
	"github.com/google/uuid"
	"go_gomoku/util"
	"go_gomoku/types"
	"go_gomoku/constants"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	gameId = 0
	userId string
	opponentId string
	yourColor string
	opponentColor string
	connection net.Conn
	yourTurn = false
	messages = []types.Message{}
	turn = 0
	board map[string]map[string]bool
	colors map[string]string
	turnOneInstructions = "You go first! Begin by placing two black pieces and then one white. Ex: 'mv 8 8, 8 7, 6 6'"
	turnTwoInstructions = "If you want to play white, play a move as normal. Otherwise, type 'mv pass'."
)

func PrintBoard() {
	for y := 0; y < 15; y++ {
		row := ""
		for x := 0; x < 16; x++ {
			coord := types.Coord {
				X: x + 1,
				Y: y,
			}

			if y == 0 {
				row += " "
			} else {
				row += "|"
			}

			if (x != 15) {
				color := util.IsTakenBy(board, coord)
				if color == constants.FREE {
					row += "_"
				} else {
					row += colors[color]
				}
			}
		}
		fmt.Println(row)
	}
}

func RefreshScreen() {
	util.CallClear()

	var turnStr string 

	if (turn > 0) {
		turnStr = "Turn #" + strconv.Itoa(turn)
		if (yourTurn) {
			turnStr += ": You"
		} else {
			turnStr += ": Opponent"
		}
	} else {
		turnStr = "Waiting for player to join..."
	}


	fmt.Println(turnStr)
	PrintBoard()
	for _, message := range(messages) {
		message.Print()
	}
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
		AddMessage("You're already in a game!!", "Gomoku")
		return
	}

	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		AddMessage("Please enter a valid integer as the game id to join!", "Gomoku")
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
		AddMessage("You're not in a game yet!", "Gomoku")
		return
	}

	if turn == 0 {
		AddMessage("The game hasn't started yet!", "Gomoku")
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
		AddMessage("You're not in a game yet!", "Gomoku")
		return
	}

	if turn == 0 {
		AddMessage("The game hasn't started yet!", "Gomoku")
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
			gameIdStr := strconv.Itoa(request.GameId)

			AddMessage("Created game #" + gameIdStr, "Gomoku")

			gameId = request.GameId
			yourTurn = true
		} else {
			AddMessage("Error! Could not create game.", "Gomoku")
		}
	case constants.JOIN:
			if request.Success {
				gameId = request.GameId
				gameIdStr := strconv.Itoa(request.GameId)
				opponentId = request.UserId
				turn = request.Turn
				yourTurn = request.YourTurn

				AddMessage("Joined game #" + gameIdStr, "Gomoku")
				AddMessage(turnOneInstructions, "Gomoku")
			} else {
				AddMessage(request.Data, "Gomoku")
			}
	case constants.OTHER_JOINED:
		if request.Success {
			turn = request.Turn
			opponentId = request.UserId

			yourTurn = request.YourTurn
			AddMessage("Let the game begin!", "Gomoku")
			if yourTurn {
				AddMessage(turnOneInstructions, "Gomoku")
			}
		}
	case constants.MESSAGE:
		if request.Success {
			AddMessage(request.Data, "Opponent")
		} else {
			AddMessage("Error! Could not parse message from opponent.", "Gomoku")
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

			yourTurn = request.YourTurn
			AddMessage(request.Data, player)

			if yourTurn && turn == 2 {
				AddMessage(turnTwoInstructions, "Gomoku")
			}
		} else {
			AddMessage(request.Data, "Gomoku")
		}
	}
}

func ListenForInput() {
	for scanner.Scan() {
		text := scanner.Text()

		switch action := text[:2]; action {
		case "hp":
			AddMessage("Type mk to make a game; jn <game_id> to join a game; mv <x> <y> to make a move; mg <message> to send a message; hp for help", "Gomoku")
		case "mk":
			CreateGame()
		case "jn":
			JoinGame(text[3:])
		case "mg":
			SendMessage(text[3:])
		case "mv":
			MakeMove(text[3:])
		default:
			AddMessage("Unrecognized command! Type 'hp' for help!", "Gomoku")
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

func Run(serverPort string) {
	colors = make(map[string]string)
	colors["white"] = "\u25CF"
	colors["black"] = "\u25CB"

	// create addresses
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println(err)
	}
	userId = uuid.String()

	conn, err := net.Dial("tcp", "localhost" + serverPort)
	if err != nil {
			fmt.Println(err)
	}

	connection = conn

	client := &types.Client{Socket: connection}

	go client.Receive(Handler)

	board = make(map[string]map[string]bool)
	PrintBoard()

	ListenForInput()
}
