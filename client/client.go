
package client

import (
	"bufio"
	// "io/ioutil"
	"net"
	"os"
	"fmt"
	"strconv"
	"go_gomoku/util"
	"go_gomoku/types"
	"go_gomoku/constants"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	gameId = 0
	connection net.Conn
	yourTurn = false
)

// Send request struct to server as byte array
func SendToServer(request types.Request) {
	fmt.Println("Sending to server...")

	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Writing...")

	connection.Write(data) // TODO: get error, handle
}

func CreateGame() {
	fmt.Println("creating game...")
	if gameId != 0 {
		fmt.Println("You're already in a game!")
		return
	}

	request := types.Request{
		UserId: 1,
		Action: constants.CREATE,
	}

	SendToServer(request)
}

func JoinGame(gameIdStr string) {
	gameId, err := strconv.Atoi(gameIdStr)
	fmt.Println(gameId, err)
	if err != nil {
		fmt.Println("Please enter a valid integer as the game id to join!")
		return
	}

	request := types.Request{
		GameId: gameId,
		UserId: 2,
		Action: constants.JOIN,
	}

	SendToServer(request)

	// result := util.HandleRwGob(rw)

	// if result.Action == constants.SUCCESS {
	// 	fmt.Println("Joined game #", result.GameId)
	// 	gameId = result.GameId
	// } else {
	// 	fmt.Println("Error! Could not create game.")
	// }
}

func SendMessage(text string) {
	if gameId == 0 {
		fmt.Println("You're not in a game yet!")
		return
	}

	request := types.Request{
		GameId: gameId,
		UserId: 1,
		Action: constants.MESSAGE,
		Data: text,
	}

	SendToServer(request)

	// result := util.HandleRwGob(rw)

	// fmt.Print(result)
	// fmt.Println(result.Data)
}

func ListenForInput() {
    for scanner.Scan() {
		text := scanner.Text()

		switch action := text[:2]; action {
		case "hp":
			fmt.Println("Type mk to make a game; mg <message> to send a message")
		case "mk":
			CreateGame()
		case "jn":
			JoinGame(text[3:])
		case "mg":
			SendMessage(text[3:])
		default:
			fmt.Println("Unrecognized command! Type 'hp' for help!")
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

func Handler(message []byte) {
	request := util.DecodeGob(message)
	fmt.Println("Decoded...")
	fmt.Println(request)

	switch action := request.Action; action {
	case constants.CREATE:
		if request.Success {
			fmt.Println("Created game #", request.GameId)
			gameId = request.GameId
			yourTurn = true
		} else {
			fmt.Println("Error! Could not create game.")
		}
	}
}


func Run(serverPort string) {
	// create addresses

	conn, error := net.Dial("tcp", "localhost" + serverPort)
	if error != nil {
			fmt.Println(error)
	}

	connection = conn

	client := &types.Client{Socket: connection}

	go client.Receive(Handler)
	ListenForInput()
}
