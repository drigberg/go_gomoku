package server

import (
    "net"
    "fmt"
    "go_gomoku/util"
    "go_gomoku/types"
    "go_gomoku/constants"
)

var (
    games map[int]types.GameRoom
    gameId = 0
)

func CreateGame(req types.Request, client *types.Client) int {
    gameId += 1
    player := types.Player{
        UserId: req.UserId,
        Spots: make(map[int][]types.Coord),
        Client: client,
    }

    players := [2]types.Player{player}

    games[gameId] = types.GameRoom{
        Id: gameId,
        Players: players,
        Messages: [6]string{},
        Turn: 0,
    }

    return gameId
}

func HandleRequest(req types.Request, client *types.Client) {
    fmt.Println("Received!", client)
    fmt.Println(req)

    switch action := req.Action; action {
    case constants.CREATE:
        gameId := CreateGame(req, client)

        response := types.Request{
            GameId: gameId,
            Action: constants.CREATE,
            Success: true,
		}

		data, err := util.GobToBytes(response)

		if err != nil {
			fmt.Println(err)
			return
		}

        select {
        case client.Data <- data:
        default:
            close(client.Data)
        }
    case constants.JOIN:
        fmt.Println("JOINING")
        game := games[req.GameId]

        player := types.Player{
            UserId: req.UserId,
            Spots: make(map[int][]types.Coord),
            Client: client,
        }

        fmt.Println(player)

        game.Players[1] = player

        fmt.Println(game.Players)
        fmt.Println(game.Players[1])
        fmt.Println(games[req.GameId])


        game.Messages[0] = "Let the game begin!"
        game.Turn = 1

        response := types.Request{
            GameId: req.GameId,
            Action: constants.JOIN,
            Success: true,
		}

		data, err := util.GobToBytes(response)

		if err != nil {
			fmt.Println(err)
			return
		}

        select {
        case client.Data <- data:
        default:
            close(client.Data)
        }
    // case constants.MESSAGE:
    //     response := types.Request{
	// 		Action: constants.MESSAGE,
    //         Data: req.Data,
    //     }

    //     game := games[req.GameId]

    //     otherPlayerRW := game.Channels[1]

	// 	data, err := util.GobToBytes(response)

	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}

    //     _, err = otherPlayerRW.Write(data)
    //     util.CheckError(err)

    //     err = otherPlayerRW.Flush()
    //     if err != nil {
    //         fmt.Println("Flush failed!")
    //     }

    //     _, err = rw.Write(data)
    //     util.CheckError(err)
    default:
        fmt.Println("Unrecognized action!")
    }

    fmt.Println(games)
    fmt.Println("\n")
}

type ClientManager struct {
    clients map[*types.Client]bool
    broadcast chan []byte
    register chan *types.Client
    unregister chan *types.Client
}

func (manager *ClientManager) receive(client *types.Client) {
    for {
        message := make([]byte, 4096)
        length, err := client.Socket.Read(message)
        if err != nil {
            manager.unregister <- client
            client.Socket.Close()
            break
        }
        if length > 0 {
            // convert to gob
            //handle
            fmt.Println("Received...")

            request := util.DecodeGob(message)
            fmt.Println("Decoded...")

            HandleRequest(request, client)
            // manager.broadcast <- message
        }
    }
}

func (manager *ClientManager) send(client *types.Client) {
    defer client.Socket.Close()
    for {
        select {
        case message, ok := <-client.Data:
            if !ok {
                return
            }
            client.Socket.Write(message)
        }
    }
}

func (manager *ClientManager) start() {
    for {
        select {
        case connection := <-manager.register:
            manager.clients[connection] = true
            fmt.Println("Added new connection!")
        case connection := <-manager.unregister:
            if _, ok := manager.clients[connection]; ok {
                close(connection.Data)
                delete(manager.clients, connection)
                fmt.Println("A connection has terminated!")
            }
        case message := <-manager.broadcast:
            for connection := range manager.clients {
                select {
                case connection.Data <- message:
                default:
                    close(connection.Data)
                    delete(manager.clients, connection)
                }
            }
        }
    }
}

func Run(port string) {
    fmt.Println("Starting server...")
    games = make(map[int]types.GameRoom)

    listener, error := net.Listen("tcp", port)

    if error != nil {
        fmt.Println(error)
    }

    manager := ClientManager{
        clients: make(map[*types.Client]bool),
        broadcast: make(chan []byte),
        register: make(chan *types.Client),
        unregister: make(chan *types.Client),
    }

    go manager.start()

    for {
        connection, _ := listener.Accept()
        if error != nil {
            fmt.Println(error)
        }
        client := &types.Client{Socket: connection, Data: make(chan []byte)}
        manager.register <- client
        go manager.receive(client)
        go manager.send(client)
    }
}


// func Run(port string) {
//     games = make(map[int]types.GameRoom)
//     tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
// 	util.CheckError(err)

//     listener, err := net.ListenTCP("tcp", tcpAddr)
// 	util.CheckError(err)

//     fmt.Println("Listening on " + listener.Addr().String())

//     for {
// 		conn, err := listener.Accept()

//         if err != nil {
//             continue
// 		}

//         go util.AcceptConnection(conn, HandleRequest)
//     }
// }