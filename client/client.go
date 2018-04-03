
package client

import (
	"bufio"
	// "io/ioutil"
	"net"
	"os"
	"fmt"
	"strconv"
	"gomoku/util"
	"gomoku/types"
	"gomoku/constants"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	gameId = 0
)

func HandleRequest(conn net.Conn, req types.Request) {
    fmt.Print("received: ", req)

    conn.Write([]byte("message received!"))
}

func CreateGame(tcpAddr *net.TCPAddr) {
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	request := types.Request{
		UserId: 1,
		Action: constants.CREATE,
	}

	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write(data)
	util.CheckError(err)

	result := util.HandleGob(conn)

	if result.Action == constants.SUCCESS {
		fmt.Println("Created game #", result.GameId)
	} else {
		fmt.Println("Error! Could not create game.")
	}
}

func JoinGame(tcpAddr *net.TCPAddr, gameIdStr string) {
	gameId, err := strconv.Atoi(gameIdStr)
	fmt.Println(gameId, err)
	if err != nil {
		fmt.Println("Please enter a valid integer as the game id to join!")
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	request := types.Request{
		GameId: gameId,
		UserId: 2,
		Action: constants.JOIN,
	}

	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write(data)
	util.CheckError(err)

	result := util.HandleGob(conn)

	if result.Action == constants.SUCCESS {
		fmt.Println("Joined game #", result.GameId)
	} else {
		fmt.Println("Error! Could not create game.")
	}
}

func SendMessage(tcpAddr *net.TCPAddr, text string) {
	if gameId == 0 {
		fmt.Println("You're not in a game yet!")
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	request := types.Request{
		GameId: gameId,
		UserId: 1,
		Action: constants.MESSAGE,
		Data: text,
	}

	data, err := util.GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write(data)
	util.CheckError(err)

	// result, err := ioutil.ReadAll(conn)

	result := util.HandleGob(conn)

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
			CreateGame(tcpAddr)
		case "jn":
			JoinGame(tcpAddr, text[3:])
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
    clientAddr, err := net.ResolveTCPAddr("tcp4", clientPort)
	util.CheckError(err)

	serverAddr, err := net.ResolveTCPAddr("tcp4", serverPort)
	util.CheckError(err)

    listener, err := net.ListenTCP("tcp", clientAddr)
	util.CheckError(err)

	fmt.Println("Running client!")

	go ListenForConnection(listener)
    ListenForInput(serverAddr)
}