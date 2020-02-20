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
	games  map[int]*types.GameRoom
	gameID int
}

// Listen starts the server
func (server *Server) Listen(port string) {
	log.Println("Starting server...")
	listener, error := net.Listen("tcp", ":"+port)
	if error != nil {
		log.Println(error)
	}

	// create socket socketClient manager
	socketClientManager := SocketClientManager{
		clients:    make(map[*types.SocketClient]bool),
		register:   make(chan *types.SocketClient),
		unregister: make(chan *types.SocketClient),
	}

	go socketClientManager.start()

	log.Println("Server listening on port " + port + "!")

	for {
		connection, _ := listener.Accept()
		if error != nil {
			log.Println(error)
		}

		socketClient := &types.SocketClient{Socket: connection, Data: make(chan []byte)}

		socketClientManager.register <- socketClient
		go socketClientManager.receive(socketClient, server)
		go socketClientManager.send(socketClient)

		server.sendHome(socketClient)
	}
}

// New creates a server instances
func New() Server {
	return Server{
		games:  make(map[int]*types.GameRoom),
		gameID: 0,
	}
}

// CreateGame creates a game and returns the id
func (server *Server) CreateGame(req types.Request, socketClient *types.SocketClient) int {
	server.gameID++
	player := types.Player{
		UserID:       req.UserID,
		SocketClient: socketClient,
	}

	players := make(map[string]*types.Player)

	players[req.UserID] = &player

	server.games[server.gameID] = &types.GameRoom{
		ID:      server.gameID,
		Players: players,
		Turn:    0,
		Board:   make(map[string]map[string]bool),
	}

	server.games[server.gameID].Board["white"] = make(map[string]bool)
	server.games[server.gameID].Board["black"] = make(map[string]bool)

	return server.gameID
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

	ok, errorResponse := helpers.CheckOwnership(server.games[req.GameID], req.UserID, move)
	if !ok {
		return false, types.Coord{}, errorResponse
	}

	return true, move, types.Request{}
}

func (server *Server) handleCreate(req types.Request, socketClient *types.SocketClient) {
	gameID := server.CreateGame(req, socketClient)
	response := types.Request{
		GameID:  gameID,
		Action:  constants.CREATE,
		Success: true,
	}
	helpers.SendToClient(response, socketClient)
}

func (server *Server) handleJoin(req types.Request, socketClient *types.SocketClient, activeGame *types.GameRoom) {
	otherClient := helpers.OtherClient(activeGame, req.UserID)

	if len(activeGame.Players) == 2 {
		response := types.Request{
			GameID:  req.GameID,
			Action:  constants.JOIN,
			Success: false,
			Data:    "Game is full already",
		}

		helpers.SendToClient(response, socketClient)
		return
	}

	if otherClient.Closed {
		response := types.Request{
			GameID:  req.GameID,
			Action:  constants.JOIN,
			Success: false,
			Data:    "Other player already left that game",
		}

		helpers.SendToClient(response, socketClient)
		return
	}

	player := types.Player{
		UserID:       req.UserID,
		SocketClient: socketClient,
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
	notification := types.Request{
		GameID:   req.GameID,
		Action:   constants.OTHERJOINED,
		Success:  true,
		YourTurn: activeGame.FirstPlayerId != req.UserID,
		Turn:     activeGame.Turn,
		UserID:   req.UserID,
	}

	helpers.SendToClient(response, socketClient)
	helpers.SendToClient(notification, otherClient)
}

func (server *Server) handleMessage(req types.Request, socketClient *types.SocketClient, activeGame *types.GameRoom) {
	response := types.Request{
		GameID:  req.GameID,
		UserID:  req.UserID,
		Action:  constants.MESSAGE,
		Data:    req.Data,
		Success: true,
	}

	otherClient := helpers.OtherClient(activeGame, req.UserID)

	helpers.SendToClient(response, otherClient)
}

func (server *Server) handleMove(req types.Request, socketClient *types.SocketClient, activeGame *types.GameRoom) {
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

		helpers.SendToClient(errorResponse, socketClient)
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

				helpers.SendToClient(errorResponse, socketClient)
				return
			}

			moves := [3]types.Coord{}
			for i, moveStr := range moveStrs {
				ok, move, errorResponse := server.ParseMove(req, moveStr)

				if !ok {
					helpers.SendToClient(errorResponse, socketClient)
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
				helpers.SendToClient(errorResponse, socketClient)
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
			helpers.SendToClient(errorResponse, socketClient)
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

	helpers.SendToClient(response, socketClient)

	otherClient := helpers.OtherClient(activeGame, req.UserID)
	response.YourTurn = true
	helpers.SendToClient(response, otherClient)
}

// HandleRequest handles a request
func (server *Server) HandleRequest(req types.Request, socketClient *types.SocketClient) {
	log.Print("Received!", socketClient, req)
	activeGame := server.games[req.GameID]

	switch action := req.Action; action {
	case constants.CREATE:
		server.handleCreate(req, socketClient)
	case constants.JOIN:
		server.handleJoin(req, socketClient, activeGame)
	case constants.MESSAGE:
		server.handleMessage(req, socketClient, activeGame)
	case constants.MOVE:
		server.handleMove(req, socketClient, activeGame)
	case constants.HOME:
		server.sendHome(socketClient)
	default:
		log.Println("Unrecognized action:", req.Action)
	}
}

func (server *Server) sendHome(socketClient *types.SocketClient) {
	home := []types.OpenRoom{}

	for _, game := range server.games {
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

	helpers.SendToClient(data, socketClient)
}

// SocketClientManager handles all clients
type SocketClientManager struct {
	clients    map[*types.SocketClient]bool
	register   chan *types.SocketClient
	unregister chan *types.SocketClient
}

// CloseSocket safely closes socket connection
func CloseSocket(socketClient *types.SocketClient) {
	socketClient.M.Lock()
	defer socketClient.M.Unlock()

	if !socketClient.Closed {
		socketClient.Socket.Close()
		socketClient.Closed = true
	}
}

func (manager *SocketClientManager) receive(socketClient *types.SocketClient, server *Server) {
	for {
		message := make([]byte, 4096)
		length, err := socketClient.Socket.Read(message)

		if length > 0 {
			request := helpers.DecodeGob(message)
			server.HandleRequest(request, socketClient)
		}

		if err != nil {
			manager.unregister <- socketClient
			CloseSocket(socketClient)
			break
		}

	}
}

func (manager *SocketClientManager) send(socketClient *types.SocketClient) {
	for {
		select {
		case message, ok := <-socketClient.Data:
			if !ok {
				return
			}
			socketClient.Socket.Write(message)
		}
	}
}

func (manager *SocketClientManager) start() {
	log.Println("SocketClient manager listening for clients joining/leaving...")
	for {
		select {
		case socketClient := <-manager.register:
			manager.clients[socketClient] = true
			log.Println("Added new socketClient!")
		case socketClient := <-manager.unregister:
			if _, ok := manager.clients[socketClient]; ok {
				log.Println("A socketClient has left!")
				close(socketClient.Data)
				delete(manager.clients, socketClient)
			}
		}
	}
}
