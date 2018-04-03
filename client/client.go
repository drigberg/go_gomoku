
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
	rw *bufio.ReadWriter
	yourTurn = false
)

func HandleRequest(rw *bufio.ReadWriter, req types.Request) {
    fmt.Print("received: ", req)

    rw.Write([]byte("message received!"))
}

func SendToServer(request types.Request) {
	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = rw.Write(data)
	util.CheckError(err)

	err = rw.Flush()

	util.CheckError(err)
}

func CreateGame() {
	if gameId != 0 {
		fmt.Println("You're already in a game!")
		return
	}

	request := types.Request{
		UserId: 1,
		Action: constants.CREATE,
	}

	SendToServer(request)

	result := util.HandleRwGob(rw)

	if result.Action == constants.SUCCESS {
		fmt.Println("Created game #", result.GameId)
		gameId = result.GameId
		yourTurn = true
	} else {
		fmt.Println("Error! Could not create game.")
	}
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

	result := util.HandleRwGob(rw)

	if result.Action == constants.SUCCESS {
		fmt.Println("Joined game #", result.GameId)
		gameId = result.GameId
	} else {
		fmt.Println("Error! Could not create game.")
	}
}

func SendMessage(tcpAddr *net.TCPAddr, text string) {
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

	result := util.HandleRwGob(rw)

	fmt.Print(result)
	fmt.Println(result.Data)
}

func ListenForInput(tcpAddr *net.TCPAddr) {
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
			SendMessage(tcpAddr, text[3:])
		default:
			fmt.Println("Unrecognized command! Type 'hp' for help!")
		}

		if scanner.Err() != nil {
			fmt.Fprintf(os.Stderr, "Error reading console input!")
		}
	}
}

func listenForWrite() {
	for {
		response, err := rw.Read([]byte(""))
		fmt.Println("RESPONSE:", response)
		fmt.Println(err)

		data := util.HandleRwGob(rw)

		fmt.Println("DATA:", data)
	}
}

func ListenForConnection(listener *net.TCPListener) {
    for {
		conn, err := listener.Accept()

        if err != nil {
            continue
		}

        go util.AcceptConnection(conn, HandleRequest)
    }
}

func Run(clientPort string, serverPort string) {
	// create addresses
    clientAddr, err := net.ResolveTCPAddr("tcp4", clientPort)
	util.CheckError(err)

	serverAddr, err := net.ResolveTCPAddr("tcp4", serverPort)
	util.CheckError(err)

	// create listener
    listener, err := net.ListenTCP("tcp", clientAddr)
	util.CheckError(err)

	fmt.Println("Running client!")

	// establish connection to server
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	util.CheckError(err)

	rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// listen?
	go listenForWrite()
	go ListenForConnection(listener)
    ListenForInput(serverAddr)
}