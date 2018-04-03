package server

import (
    "net"
    "fmt"
    "bufio"
    "gomoku/util"
    "gomoku/types"
    "gomoku/constants"
)

var (
    games map[int]types.GameRoom
    gameId = 0
)

func CreateGame(req types.Request, rw *bufio.ReadWriter) int {
    gameId += 1
    players := [2]int{req.UserId, -1}
    channels := [2]*bufio.ReadWriter{rw}

    spots := make(map[int][]types.Coord)
    spots[req.UserId] = []types.Coord{}

    games[gameId] = types.GameRoom{
        Id: gameId,
        Channels: channels,
        Spots: spots,
        Players: players,
        Messages: [6]string{},
    }

    return gameId
}

func HandleRequest(rw *bufio.ReadWriter, req types.Request) {
    fmt.Println("Received!")
    fmt.Println(req)

    switch action := req.Action; action {
    case constants.CREATE:
        gameId := CreateGame(req, rw)

        response := types.Request{
            GameId: gameId,
			Action: constants.SUCCESS,
		}

		data, err := util.GobToBytes(response)

		if err != nil {
			fmt.Println(err)
			return
		}

        _, err = rw.Write(data)
        util.CheckError(err)
    case constants.JOIN:
        game := games[req.GameId]

        game.Players[1] = req.UserId
        game.Spots[req.UserId] = []types.Coord{}
        game.Messages[0] = "Let the game begin!"
        game.Turn = 1

        response := types.Request{
            GameId: req.GameId,
			Action: constants.SUCCESS,
		}

		data, err := util.GobToBytes(response)

		if err != nil {
			fmt.Println(err)
			return
		}

        _, err = rw.Write(data)
        util.CheckError(err)

        // choose random player to go first
    case constants.MESSAGE:
        response := types.Request{
			Action: constants.MESSAGE,
			Data: req.Data,
        }

        game := games[req.GameId]

        otherPlayerRW := game.Channels[1]

		data, err := util.GobToBytes(response)

		if err != nil {
			fmt.Println(err)
			return
		}

        _, err = otherPlayerRW.Write(data)
        util.CheckError(err)

        err = otherPlayerRW.Flush()
        if err != nil {
            fmt.Println("Flush failed!")
        }

        _, err = rw.Write(data)
        util.CheckError(err)


    default:
        fmt.Println("Unrecognized action!")
    }

	err := rw.Flush()
	if err != nil {
		fmt.Println("Flush failed!")
	}

    fmt.Println(games)
    fmt.Println("\n")
}

func Run(port string) {
    games = make(map[int]types.GameRoom)
    tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	util.CheckError(err)

    listener, err := net.ListenTCP("tcp", tcpAddr)
	util.CheckError(err)

    fmt.Println("Listening on " + listener.Addr().String())

    for {
		conn, err := listener.Accept()

        if err != nil {
            continue
		}

        go util.AcceptConnection(conn, HandleRequest)
    }
}