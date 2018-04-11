package server

import (
    "net"
    "fmt"
    "strings"
    "strconv"
    "math/rand"
    "go_gomoku/util"
    "go_gomoku/types"
    "go_gomoku/constants"
    "go_gomoku/helpers"
)


var (
    games map[int]*types.GameRoom
    gameId = 0
)

func CreateGame(req types.Request, client *types.Client) int {
    gameId += 1
    player := types.Player{
        UserId: req.UserId,
        Client: client,
    }

    players := make(map[string]*types.Player)

    players[req.UserId] = &player

    games[gameId] = &types.GameRoom{
        Id: gameId,
        Players: players,
        Turn: 0,
        Board: make(map[string]map[string]bool),
    }

    games[gameId].Board["white"] = make(map[string]bool)
    games[gameId].Board["black"] = make(map[string]bool)

    return gameId
}

func ParseMove(req types.Request, moveStr string) (bool, types.Coord, types.Request) {
    coordinates := strings.Split(moveStr, " ")

    if (len(coordinates) != 2) {
        errorResponse := types.Request{
            GameId: req.GameId,
            UserId: req.UserId,
            Action: constants.MOVE,
            Data: "The syntax for this move is mv <x> <y>, <x> <y>, <x> <y>",
            Success: false,
        }

        return false, types.Coord{}, errorResponse
    }

    x, xErr := strconv.Atoi(coordinates[0])
    y, yErr := strconv.Atoi(coordinates[1])

    isNil := xErr != nil || yErr != nil
    notInRange := x < 1 || x > 15 || y < 1 || y > 15

    if isNil || notInRange {
        errorResponse := types.Request{
            GameId: req.GameId,
            UserId: req.UserId,
            Action: constants.MOVE,
            Data: "Both x and y must be integers from 1 to 15",
            Success: false,
        }

        return false, types.Coord{}, errorResponse
    }

    move := types.Coord{
        X: x,
        Y: y,
    }

    ok, errorResponse := helpers.CheckOwnership(games[req.GameId], req.UserId, move)

    if !ok {
        return false, types.Coord{}, errorResponse
    }

    return true, move, types.Request{}
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

		helpers.SendToClient(response, client)
    case constants.JOIN:
        otherClient := helpers.OtherClient(games[req.GameId], req.UserId)

        if len(games[req.GameId].Players) == 2 {
            response := types.Request{
                GameId: req.GameId,
                Action: constants.JOIN,
                Success: false,
                Data: "Game is full already",
            }

            helpers.SendToClient(response, client)
            return;
        }

        if otherClient.Closed {
            response := types.Request{
                GameId: req.GameId,
                Action: constants.JOIN,
                Success: false,
                Data: "Other player already left that game",
            }

            helpers.SendToClient(response, client)
            return;
        }

        player := types.Player{
            UserId: req.UserId,
            Client: client,
        }

        games[req.GameId].Players[req.UserId] = &player
        games[req.GameId].Turn = 1

        games[req.GameId].FirstPlayerId = req.UserId

        opponentId := helpers.GetOpponentId(games[req.GameId], req.UserId)

        if rand.Intn(2) == 0 {
            games[req.GameId].FirstPlayerId = opponentId
        }

        response := types.Request{
            GameId: req.GameId,
            UserId: opponentId,
            Action: constants.JOIN,
            Success: true,
            YourTurn: games[req.GameId].FirstPlayerId == req.UserId,
            Turn: games[req.GameId].Turn,
		}

		helpers.SendToClient(response, client)

        notification := types.Request{
            GameId: req.GameId,
            Action: constants.OTHER_JOINED,
            Success: true,
            YourTurn: games[req.GameId].FirstPlayerId != req.UserId,
            Turn: games[req.GameId].Turn,
            UserId: req.UserId,
		}

		helpers.SendToClient(notification, otherClient)
    case constants.MESSAGE:
        response := types.Request{
            GameId: req.GameId,
            UserId: req.UserId,
			Action: constants.MESSAGE,
            Data: req.Data,
            Success: true,
        }

        otherClient := helpers.OtherClient(games[req.GameId], req.UserId)

        helpers.SendToClient(response, otherClient)
    case constants.HOME:
        sendHome(client)
    case constants.MOVE:
        response := types.Request{
            GameId: req.GameId,
            UserId: req.UserId,
            Action: constants.MOVE,
        }

        valid := helpers.IsTurn(games[req.GameId], req.UserId)
        gameOver := false
        var message string

        if !valid {
            errorResponse := types.Request{
                GameId: req.GameId,
                UserId: req.UserId,
                Action: constants.MOVE,
                Data: "It's not your turn!",
                Success: false,
            }

            helpers.SendToClient(errorResponse, client)
            return
        }

        switch turn := games[req.GameId].Turn; turn {
        case 1:
            if games[req.GameId].Turn == 1 {
                moveStrs := strings.Split(req.Data, ", ")
                if len(moveStrs) != 3 {
                    errorResponse := types.Request{
                        GameId: req.GameId,
                        UserId: req.UserId,
                        Action: constants.MOVE,
                        Data: "Please choose exectly three sets of two values",
                        Success: false,
                    }

                    helpers.SendToClient(errorResponse, client)
                    return
                }

                moves := [3]types.Coord{}
                for i, moveStr := range(moveStrs) {
                    ok, move, errorResponse := ParseMove(req, moveStr)

                    if !ok {
                        helpers.SendToClient(errorResponse, client)
                        return
                    }

                    moves[i] = move
                }

                games[req.GameId].PlayMove(moves[0], "black")
                games[req.GameId].PlayMove(moves[1], "black")
                games[req.GameId].PlayMove(moves[2], "white")
                message = "(played black on " + moves[0].String() + ", black on " +  moves[1].String() + ", and white on " +  moves[2].String() + " )"
            }
        case 2:
            response.Colors = make(map[string]string)
            opponentId := helpers.GetOpponentId(games[req.GameId], req.UserId)

            if req.Data == "pass" {
                message = "(passed and is now black -- back to player 1)"

                games[req.GameId].Players[req.UserId].Color = "black"
                games[req.GameId].Players[opponentId].Color = "white"
                response.Colors[req.UserId] = "black"
                response.Colors[opponentId] = "white"
            } else {
                ok, move, errorResponse := ParseMove(req, req.Data)

                if !ok {
                    helpers.SendToClient(errorResponse, client)
                    return
                }

                games[req.GameId].Players[req.UserId].Color = "white"
                games[req.GameId].Players[opponentId].Color = "black"
                response.Colors[req.UserId] = "white"
                response.Colors[opponentId] = "black"

                games[req.GameId].PlayMove(move, "white")
            }
        default:
            ok, move, errorResponse := ParseMove(req, req.Data)

            if !ok {
                helpers.SendToClient(errorResponse, client)
                return
            }

            games[req.GameId].PlayMove(move, games[req.GameId].Players[req.UserId].Color)

            gameOver = helpers.CheckForWin(games[req.GameId], move, games[req.GameId].Players[req.UserId].Color)
            if (gameOver) {
                games[req.GameId].IsOver = true
                response.GameOver = true
                message = "won!!!! (" + req.Data + " )"
            } else {
                message = "(played on " + req.Data + " )"
            }
        }

        response.Success = true
        response.Board = games[req.GameId].Board
        response.Data = message

        if !gameOver {
            games[req.GameId].Turn += 1

            response.YourTurn = false
            response.Turn = games[req.GameId].Turn
        }

        helpers.SendToClient(response, client)

        otherClient := helpers.OtherClient(games[req.GameId], req.UserId)
        response.YourTurn = true
        helpers.SendToClient(response, otherClient)

    default:
        fmt.Println("Unrecognized action!")
    }
}

func sendHome(client *types.Client) {
    home := []types.OpenRoom{}

    for _, game := range(games) {
        if !game.IsOver && len(game.Players) == 1 {
            var userId string
            for id := range(game.Players) {
                userId = id
            }


            openRoom := types.OpenRoom{
                Id: game.Id,
                UserId: userId,
            }

            home = append(home, openRoom)
        }
    }

    data := types.Request{
        Action: constants.HOME,
        Home: home,
    }

    helpers.SendToClient(data, client)
}

type ClientManager struct {
    clients map[*types.Client]bool
    register chan *types.Client
    unregister chan *types.Client
}

func CloseSocket(client *types.Client) {
    client.M.Lock()
    defer client.M.Unlock()

    if !client.Closed {
        client.Socket.Close()
        client.Closed = true
    }
}

func (manager *ClientManager) receive(client *types.Client) {
    for {
        message := make([]byte, 4096)
        length, err := client.Socket.Read(message)

        if length > 0 {
            request := util.DecodeGob(message)
            HandleRequest(request, client)
        }

        if err != nil {
            manager.unregister <- client
            CloseSocket(client)
            break
        }

    }
}

func (manager *ClientManager) send(client *types.Client) {
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
    fmt.Println("Listening for TCP connections...")
    for {
        select {
        case connection := <-manager.register:
            manager.clients[connection] = true
            fmt.Println("Added new connection!")
        case connection := <-manager.unregister:
            if _, ok := manager.clients[connection]; ok {
                fmt.Println("A connection has terminated!")
                close(connection.Data)
                delete(manager.clients, connection)
            }
        }
    }
}

func Run(port string) {
    fmt.Println("Starting server...")

    games = make(map[int]*types.GameRoom)

    listener, error := net.Listen("tcp", port)

    if error != nil {
        fmt.Println(error)
    }

    manager := ClientManager{
        clients: make(map[*types.Client]bool),
        register: make(chan *types.Client),
        unregister: make(chan *types.Client),
    }

    go manager.start()

    fmt.Println("Server running on port " + port + "!")

    for {
        connection, _ := listener.Accept()
        if error != nil {
            fmt.Println(error)
        }

        client := &types.Client{Socket: connection, Data: make(chan []byte)}

        manager.register <- client
        go manager.receive(client)
        go manager.send(client)

        sendHome(client)
    }
}
