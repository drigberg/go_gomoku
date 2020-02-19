package server

import (
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

// Server handles all requests and game states
type Server struct {
	Games  map[int]*types.GameRoom
	GameID int
}

// Listen starts the server
func (server *Server) Listen(port string) {
	log.Println("Starting server...")
	listener, error := net.Listen("tcp", ":"+port)
	if error != nil {
		log.Println(error)
	}

	// create client manager
	manager := ClientManager{
		clients:    make(map[*types.Client]bool),
		register:   make(chan *types.Client),
		unregister: make(chan *types.Client),
	}

	go manager.start()

	log.Println("Server listening on port " + port + "!")

	for {
		connection, _ := listener.Accept()
		if error != nil {
			log.Println(error)
		}

		client := &types.Client{Socket: connection, Data: make(chan []byte)}

		manager.register <- client
		go manager.receive(client, server)
		go manager.send(client)

		server.sendHome(client)
	}
}

// CreateGame creates a game and returns the id
func (server *Server) CreateGame(req types.Request, client *types.Client) int {
	server.GameID++
	player := types.Player{
		UserID: req.UserID,
		Client: client,
	}

	players := make(map[string]*types.Player)

	players[req.UserID] = &player

	server.Games[server.GameID] = &types.GameRoom{
		ID:      server.GameID,
		Players: players,
		Turn:    0,
		Board:   make(map[string]map[string]bool),
	}

	server.Games[server.GameID].Board["white"] = make(map[string]bool)
	server.Games[server.GameID].Board["black"] = make(map[string]bool)

	return server.GameID
}

// ParseMove validates the string from a mv command
func (server *Server) ParseMove(req types.Request, moveStr string) (bool, types.Coord, types.Request) {
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

	ok, errorResponse := helpers.CheckOwnership(server.Games[req.GameID], req.UserID, move)

	if !ok {
		return false, types.Coord{}, errorResponse
	}

	return true, move, types.Request{}
}

// HandleRequest handles a request
func (server *Server) HandleRequest(req types.Request, client *types.Client) {
	log.Print("Received!", client, req)
	activeGame := server.Games[req.GameID]

	switch action := req.Action; action {
	case constants.CREATE:
		gameID := server.CreateGame(req, client)

		response := types.Request{
			GameID:  gameID,
			Action:  constants.CREATE,
			Success: true,
		}

		helpers.SendToClient(response, client)
	case constants.JOIN:
		otherClient := helpers.OtherClient(activeGame, req.UserID)

		if len(activeGame.Players) == 2 {
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

		activeGame.Players[req.UserID] = &player
		activeGame.Turn = 1

		activeGame.FirstPlayerId = req.UserID

		opponentID := helpers.GetOpponentID(activeGame, req.UserID)

		if rand.Intn(2) == 0 {
			activeGame.FirstPlayerId = opponentID
		}

		response := types.Request{
			GameID:   req.GameID,
			UserID:   opponentID,
			Action:   constants.JOIN,
			Success:  true,
			YourTurn: activeGame.FirstPlayerId == req.UserID,
			Turn:     activeGame.Turn,
		}

		helpers.SendToClient(response, client)

		notification := types.Request{
			GameID:   req.GameID,
			Action:   constants.OTHERJOINED,
			Success:  true,
			YourTurn: activeGame.FirstPlayerId != req.UserID,
			Turn:     activeGame.Turn,
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

		otherClient := helpers.OtherClient(activeGame, req.UserID)

		helpers.SendToClient(response, otherClient)
	case constants.HOME:
		server.sendHome(client)
	case constants.MOVE:
		response := types.Request{
			GameID: req.GameID,
			UserID: req.UserID,
			Action: constants.MOVE,
		}

		valid := helpers.IsTurn(activeGame, req.UserID)
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

		switch turn := activeGame.Turn; turn {
		case 1:
			if activeGame.Turn == 1 {
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
					ok, move, errorResponse := server.ParseMove(req, moveStr)

					if !ok {
						helpers.SendToClient(errorResponse, client)
						return
					}

					moves[i] = move
				}

				activeGame.PlayMove(moves[0], "black")
				activeGame.PlayMove(moves[1], "black")
				activeGame.PlayMove(moves[2], "white")
				message = "(played black on " + moves[0].String() + ", black on " + moves[1].String() + ", and white on " + moves[2].String() + " )"
			}
		case 2:
			response.Colors = make(map[string]string)
			opponentID := helpers.GetOpponentID(activeGame, req.UserID)

			if req.Data == "pass" {
				message = "(passed and is now black -- back to player 1)"

				activeGame.Players[req.UserID].Color = "black"
				activeGame.Players[opponentID].Color = "white"
				response.Colors[req.UserID] = "black"
				response.Colors[opponentID] = "white"
			} else {
				ok, move, errorResponse := server.ParseMove(req, req.Data)

				if !ok {
					helpers.SendToClient(errorResponse, client)
					return
				}

				activeGame.Players[req.UserID].Color = "white"
				activeGame.Players[opponentID].Color = "black"
				response.Colors[req.UserID] = "white"
				response.Colors[opponentID] = "black"

				activeGame.PlayMove(move, "white")
				message = "(played on " + req.Data + " )"
			}
		default:
			ok, move, errorResponse := server.ParseMove(req, req.Data)

			if !ok {
				helpers.SendToClient(errorResponse, client)
				return
			}

			activeGame.PlayMove(move, activeGame.Players[req.UserID].Color)

			gameOver = helpers.CheckForWin(activeGame, move, activeGame.Players[req.UserID].Color)
			if gameOver {
				activeGame.IsOver = true
				response.GameOver = true
				message = "won!!!! (" + req.Data + " )"
			} else {
				message = "(played on " + req.Data + " )"
			}
		}

		response.Success = true
		response.Board = activeGame.Board
		response.Data = message

		if !gameOver {
			activeGame.Turn++

			response.YourTurn = false
			response.Turn = activeGame.Turn
		}

		helpers.SendToClient(response, client)

		otherClient := helpers.OtherClient(activeGame, req.UserID)
		response.YourTurn = true
		helpers.SendToClient(response, otherClient)

	default:
		log.Println("Unrecognized action:", req.Action)
	}
}

func (server *Server) sendHome(client *types.Client) {
	home := []types.OpenRoom{}

	for _, game := range server.Games {
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

func (manager *ClientManager) receive(client *types.Client, server *Server) {
	for {
		message := make([]byte, 4096)
		length, err := client.Socket.Read(message)

		if length > 0 {
			request := helpers.DecodeGob(message)
			server.HandleRequest(request, client)
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
	log.Println("Client manager listening for clients joining/leaving...")
	for {
		select {
		case client := <-manager.register:
			manager.clients[client] = true
			log.Println("Added new client!")
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				log.Println("A client has left!")
				close(client.Data)
				delete(manager.clients, client)
			}
		}
	}
}
