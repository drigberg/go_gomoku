package server

import (
	"fmt"
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"go_gomoku/util"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

var (
	games  map[int]*types.GameRoom
	gameID = 0
)

// CreateGame creates a game and returns the id
func CreateGame(req types.Request, client *types.Client) int {
	gameID++
	player := types.Player{
		UserID: req.UserID,
		Client: client,
	}

	players := make(map[string]*types.Player)

	players[req.UserID] = &player

	games[gameID] = &types.GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   make(map[string]map[string]bool),
	}

	games[gameID].Board["white"] = make(map[string]bool)
	games[gameID].Board["black"] = make(map[string]bool)

	return gameID
}

// ParseMove validates the string from a mv command
func ParseMove(req types.Request, moveStr string) (bool, types.Coord, types.Request) {
	coordinates := strings.Split(moveStr, " ")

	if len(coordinates) != 2 {
		errorResponse := types.Request{
			GameID:  req.GameID,
			UserID:  req.UserID,
			Action:  constants.MOVE,
			Data:    "The syntax for a move is mv <x> <y>",
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
			GameID:  req.GameID,
			UserID:  req.UserID,
			Action:  constants.MOVE,
			Data:    "Both x and y must be integers from 1 to 15",
			Success: false,
		}

		return false, types.Coord{}, errorResponse
	}

	move := types.Coord{
		X: x,
		Y: y,
	}

	ok, errorResponse := helpers.CheckOwnership(games[req.GameID], req.UserID, move)

	if !ok {
		return false, types.Coord{}, errorResponse
	}

	return true, move, types.Request{}
}

// HandleRequest handles a request
func HandleRequest(req types.Request, client *types.Client) {
	fmt.Println("Received!", client, req)

	switch action := req.Action; action {
	case constants.CREATE:
		gameID := CreateGame(req, client)

		response := types.Request{
			GameID:  gameID,
			Action:  constants.CREATE,
			Success: true,
		}

		helpers.SendToClient(response, client)
	case constants.JOIN:
		otherClient := helpers.OtherClient(games[req.GameID], req.UserID)

		if len(games[req.GameID].Players) == 2 {
			response := types.Request{
				GameID:  req.GameID,
				Action:  constants.JOIN,
				Success: false,
				Data:    "Game is full already",
			}

			helpers.SendToClient(response, client)
			return
		}

		if otherClient.Closed {
			response := types.Request{
				GameID:  req.GameID,
				Action:  constants.JOIN,
				Success: false,
				Data:    "Other player already left that game",
			}

			helpers.SendToClient(response, client)
			return
		}

		player := types.Player{
			UserID: req.UserID,
			Client: client,
		}

		games[req.GameID].Players[req.UserID] = &player
		games[req.GameID].Turn = 1

		games[req.GameID].FirstPlayerId = req.UserID

		opponentID := helpers.GetOpponentID(games[req.GameID], req.UserID)

		if rand.Intn(2) == 0 {
			games[req.GameID].FirstPlayerId = opponentID
		}

		response := types.Request{
			GameID:   req.GameID,
			UserID:   opponentID,
			Action:   constants.JOIN,
			Success:  true,
			YourTurn: games[req.GameID].FirstPlayerId == req.UserID,
			Turn:     games[req.GameID].Turn,
		}

		helpers.SendToClient(response, client)

		notification := types.Request{
			GameID:   req.GameID,
			Action:   constants.OTHERJOINED,
			Success:  true,
			YourTurn: games[req.GameID].FirstPlayerId != req.UserID,
			Turn:     games[req.GameID].Turn,
			UserID:   req.UserID,
		}

		helpers.SendToClient(notification, otherClient)
	case constants.MESSAGE:
		response := types.Request{
			GameID:  req.GameID,
			UserID:  req.UserID,
			Action:  constants.MESSAGE,
			Data:    req.Data,
			Success: true,
		}

		otherClient := helpers.OtherClient(games[req.GameID], req.UserID)

		helpers.SendToClient(response, otherClient)
	case constants.HOME:
		sendHome(client)
	case constants.MOVE:
		response := types.Request{
			GameID: req.GameID,
			UserID: req.UserID,
			Action: constants.MOVE,
		}

		valid := helpers.IsTurn(games[req.GameID], req.UserID)
		gameOver := false
		var message string

		if !valid {
			errorResponse := types.Request{
				GameID:  req.GameID,
				UserID:  req.UserID,
				Action:  constants.MOVE,
				Data:    "It's not your turn!",
				Success: false,
			}

			helpers.SendToClient(errorResponse, client)
			return
		}

		switch turn := games[req.GameID].Turn; turn {
		case 1:
			if games[req.GameID].Turn == 1 {
				moveStrs := strings.Split(req.Data, ", ")
				if len(moveStrs) != 3 {
					errorResponse := types.Request{
						GameID:  req.GameID,
						UserID:  req.UserID,
						Action:  constants.MOVE,
						Data:    "Please choose exectly three sets of two values",
						Success: false,
					}

					helpers.SendToClient(errorResponse, client)
					return
				}

				moves := [3]types.Coord{}
				for i, moveStr := range moveStrs {
					ok, move, errorResponse := ParseMove(req, moveStr)

					if !ok {
						helpers.SendToClient(errorResponse, client)
						return
					}

					moves[i] = move
				}

				games[req.GameID].PlayMove(moves[0], "black")
				games[req.GameID].PlayMove(moves[1], "black")
				games[req.GameID].PlayMove(moves[2], "white")
				message = "(played black on " + moves[0].String() + ", black on " + moves[1].String() + ", and white on " + moves[2].String() + " )"
			}
		case 2:
			response.Colors = make(map[string]string)
			opponentID := helpers.GetOpponentID(games[req.GameID], req.UserID)

			if req.Data == "pass" {
				message = "(passed and is now black -- back to player 1)"

				games[req.GameID].Players[req.UserID].Color = "black"
				games[req.GameID].Players[opponentID].Color = "white"
				response.Colors[req.UserID] = "black"
				response.Colors[opponentID] = "white"
			} else {
				ok, move, errorResponse := ParseMove(req, req.Data)

				if !ok {
					helpers.SendToClient(errorResponse, client)
					return
				}

				games[req.GameID].Players[req.UserID].Color = "white"
				games[req.GameID].Players[opponentID].Color = "black"
				response.Colors[req.UserID] = "white"
				response.Colors[opponentID] = "black"

				games[req.GameID].PlayMove(move, "white")
				message = "(played on " + req.Data + " )"
			}
		default:
			ok, move, errorResponse := ParseMove(req, req.Data)

			if !ok {
				helpers.SendToClient(errorResponse, client)
				return
			}

			games[req.GameID].PlayMove(move, games[req.GameID].Players[req.UserID].Color)

			gameOver = helpers.CheckForWin(games[req.GameID], move, games[req.GameID].Players[req.UserID].Color)
			if gameOver {
				games[req.GameID].IsOver = true
				response.GameOver = true
				message = "won!!!! (" + req.Data + " )"
			} else {
				message = "(played on " + req.Data + " )"
			}
		}

		response.Success = true
		response.Board = games[req.GameID].Board
		response.Data = message

		if !gameOver {
			games[req.GameID].Turn++

			response.YourTurn = false
			response.Turn = games[req.GameID].Turn
		}

		helpers.SendToClient(response, client)

		otherClient := helpers.OtherClient(games[req.GameID], req.UserID)
		response.YourTurn = true
		helpers.SendToClient(response, otherClient)

	default:
		fmt.Println("Unrecognized action:", req.Action)
	}
}

func sendHome(client *types.Client) {
	home := []types.OpenRoom{}

	for _, game := range games {
		if !game.IsOver && len(game.Players) == 1 {
			var userID string
			for id := range game.Players {
				userID = id
			}

			openRoom := types.OpenRoom{
				ID:     game.ID,
				UserID: userID,
			}

			home = append(home, openRoom)
		}
	}

	data := types.Request{
		Action: constants.HOME,
		Home:   home,
	}

	helpers.SendToClient(data, client)
}

// ClientManager handles all clients
type ClientManager struct {
	clients    map[*types.Client]bool
	register   chan *types.Client
	unregister chan *types.Client
}

// CloseSocket safely closes socket connection
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
	fmt.Println("Client manager listening for clients joining/leaving...")
	for {
		select {
		case client := <-manager.register:
			manager.clients[client] = true
			fmt.Println("Added new client!")
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				fmt.Println("A client has left!")
				close(client.Data)
				delete(manager.clients, client)
			}
		}
	}
}

// Run is the task for the client
func Run(port string) {
	fmt.Println("Starting server...")

	games = make(map[int]*types.GameRoom)

	listener, error := net.Listen("tcp", ":"+port)

	if error != nil {
		fmt.Println(error)
	}

	manager := ClientManager{
		clients:    make(map[*types.Client]bool),
		register:   make(chan *types.Client),
		unregister: make(chan *types.Client),
	}

	go manager.start()

	fmt.Println("Server listening on port " + port + "!")

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
