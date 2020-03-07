package server

import (
	"go_gomoku/board"
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GameRoom contains all info related to a room
type GameRoom struct {
	M 		      sync.Mutex
	ID            int
	Players       map[string]*types.Player
	Turn          int
	Board         board.Board
	FirstPlayerID string
	IsOver        bool
}

// PlayMove places a piece
func (game *GameRoom) PlayMove(move types.Coord, color string) {
	moveStr := move.String()
	game.Board.Spaces[color][moveStr] = true
}

// GetOpponentID returns the other player's id
func GetOpponentID(game *GameRoom, userID string) string {
	for id := range game.Players {
		if id != userID {
			return id
		}
	}
	return ""
}

// OtherClient returns the other player's client connection
func OtherClient(game *GameRoom, userID string) *types.SocketClient {
	opponentID := GetOpponentID(game, userID)

	player := game.Players[opponentID]
	
	if player == nil {
		return nil
	}
	return player.SocketClient
}

// IsTurn turns if it's a user's turn or not
func IsTurn(game *GameRoom, userID string) bool {
	if userID == game.FirstPlayerID {
		return game.Turn%2 == 1
	}
	return game.Turn%2 == 0
}

// Server handles all requests and game states
type Server struct {
	M      			sync.Mutex
	DisablePrint	bool
	games  			map[int]*GameRoom
	gameID 			int
	quit 			chan interface{}
	listener 		net.Listener
	wg 				sync.WaitGroup
}

// New creates a server instances
func New() Server {
	return Server{
		games:  make(map[int]*GameRoom),
		gameID: 0,
		quit: make(chan interface {}),
	}
}

func (server *Server) printString(message string) {
	if !server.DisablePrint {
		log.Println(message)
	}
}

func (server *Server) printError(err error) {
	if !server.DisablePrint {
		log.Println(err)
	}
}

// CreateGame creates a game and returns the id
func (server *Server) CreateGame(req types.Request, socketClient *types.SocketClient) int {
	server.M.Lock()
	defer server.M.Unlock()
	defer func() { server.gameID++ }()

	player := types.Player{
		UserID:       req.UserID,
		SocketClient: socketClient,
	}

	players := make(map[string]*types.Player)

	players[req.UserID] = &player

	server.games[server.gameID] = &GameRoom{
		ID:      server.gameID,
		Players: players,
		Turn:    0,
		Board:   board.New(),
	}

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

	ok, errorResponse := server.games[req.GameID].Board.CheckOwnership(req.GameID, req.UserID, move)
	if !ok {
		return false, types.Coord{}, errorResponse
	}

	return true, move, types.Request{}
}

type SocketClientResponse struct {
	socketClient *types.SocketClient
	response types.Request
}

// SendBackoff tries to send and keeps trying
func (socketClientResponse *SocketClientResponse) sendBackoff(data []byte, i int, disablePrint bool) {
	if i > 1 {
		if !disablePrint {
			log.Println("Retrying message send! Attempt:", i)
		}
	}
	if socketClientResponse.socketClient.Closed {
		return
	}

	select {
	case socketClientResponse.socketClient.Data <- data:
		return
	default:
		time.Sleep(500 * time.Millisecond)

		if i > 5 {
			return
		}
		socketClientResponse.sendBackoff(data, i+1, disablePrint)
	}
}

// SendToClient tries to send a request to socketClient, with backoff
func (socketClientResponse *SocketClientResponse) send(disablePrint bool) {
	data, err := helpers.GobToBytes(socketClientResponse.response)

	if err != nil {
		if !disablePrint {
			log.Println(err)
		}
		return
	}

	socketClientResponse.sendBackoff(data, 1, disablePrint)
}

func (server *Server) handleCreate(req types.Request, socketClient *types.SocketClient) []SocketClientResponse {
	gameID := server.CreateGame(req, socketClient)
	response := types.Request{
		GameID:  gameID,
		Action:  constants.CREATE,
		Success: true,
	}
	return []SocketClientResponse{
		SocketClientResponse{
			socketClient,
			response,
		},
	}
}

func (server *Server) handleJoin(req types.Request, socketClient *types.SocketClient, activeGame *GameRoom) []SocketClientResponse {
	otherClient := OtherClient(activeGame, req.UserID)

	if otherClient == nil {
		response := types.Request{
			GameID:  req.GameID,
			Action:  constants.JOIN,
			Success: false,
			Data:    "You are already in this room, and you can't go back! Sorry!",
		}

		return []SocketClientResponse{
			SocketClientResponse{
				socketClient,
				response,
			},
		}
	}

	if len(activeGame.Players) == 2 {
		response := types.Request{
			GameID:  req.GameID,
			Action:  constants.JOIN,
			Success: false,
			Data:    "Game is full already",
		}

		return []SocketClientResponse{
			SocketClientResponse{
				socketClient,
				response,
			},
		}
	}

	if otherClient.Closed {
		response := types.Request{
			GameID:  req.GameID,
			Action:  constants.JOIN,
			Success: false,
			Data:    "Other player already left that game",
		}

		return []SocketClientResponse{
			SocketClientResponse{
				socketClient,
				response,
			},
		}
	}

	player := types.Player{
		UserID:       req.UserID,
		SocketClient: socketClient,
	}

	activeGame.Players[req.UserID] = &player
	activeGame.Turn = 1
	activeGame.FirstPlayerID = req.UserID
	opponentID := GetOpponentID(activeGame, req.UserID)
	if rand.Intn(2) == 0 {
		activeGame.FirstPlayerID = opponentID
	}

	// OpponentID is used to alert player to opponent's ID -- should be in Data
	response1 := types.Request{
		GameID:   req.GameID,
		UserID:   opponentID,
		Action:   constants.JOIN,
		YourTurn: activeGame.FirstPlayerID == req.UserID,
		Success:  true,
		Turn:     activeGame.Turn,
	}

	// Req.UserID is used to alert player to new player ID -- should be in Data
	response2 := types.Request{
		GameID:   req.GameID,
		UserID:   req.UserID,
		Action:   constants.OTHERJOINED,
		YourTurn: activeGame.FirstPlayerID != req.UserID,
		Success:  true,
		Turn:     activeGame.Turn,
	}

	return []SocketClientResponse{
		SocketClientResponse{
			socketClient,
			response1,
		},
		SocketClientResponse{
			otherClient,
			response2,
		},
	}
}

func (server *Server) handleMessage(req types.Request, socketClient *types.SocketClient, activeGame *GameRoom) []SocketClientResponse {
	response := types.Request{
		GameID:  req.GameID,
		UserID:  req.UserID,
		Action:  constants.MESSAGE,
		Data:    req.Data,
		Success: true,
	}
	otherClient := OtherClient(activeGame, req.UserID)

	return []SocketClientResponse{
		SocketClientResponse{
			otherClient,
			response,
		},
	}
}

func (server *Server) handleMove(req types.Request, socketClient *types.SocketClient, activeGame *GameRoom) []SocketClientResponse {
	response := types.Request{
		GameID: req.GameID,
		UserID: req.UserID,
		Action: constants.MOVE,
	}

	valid := IsTurn(activeGame, req.UserID)
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

		return []SocketClientResponse{
			SocketClientResponse{
				socketClient,
				errorResponse,
			},
		}
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

				return []SocketClientResponse{
					SocketClientResponse{
						socketClient,
						errorResponse,
					},
				}
			}

			moves := [3]types.Coord{}
			for i, moveStr := range moveStrs {
				ok, move, errorResponse := server.ParseMove(req, moveStr)

				if !ok {
					return []SocketClientResponse{
						SocketClientResponse{
							socketClient,
							errorResponse,
						},
					}
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
		opponentID := GetOpponentID(activeGame, req.UserID)

		if req.Data == "pass" {
			message = "(passed and is now black -- back to player 1)"

			activeGame.Players[req.UserID].Color = "black"
			activeGame.Players[opponentID].Color = "white"
			response.Colors[req.UserID] = "black"
			response.Colors[opponentID] = "white"
		} else {
			ok, move, errorResponse := server.ParseMove(req, req.Data)

			if !ok {
				return []SocketClientResponse{
					SocketClientResponse{
						socketClient,
						errorResponse,
					},
				}
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
			return []SocketClientResponse{
				SocketClientResponse{
					socketClient,
					errorResponse,
				},
			}
		}

		activeGame.PlayMove(move, activeGame.Players[req.UserID].Color)

		gameOver = activeGame.Board.CheckForWin(move, activeGame.Players[req.UserID].Color)
		if gameOver {
			activeGame.IsOver = true
			response.GameOver = true
			message = "won!!!! (" + req.Data + " )"
		} else {
			message = "(played on " + req.Data + " )"
		}
	}

	response.Success = true
	response.Board = activeGame.Board.Spaces
	response.Data = message

	if !gameOver {
		activeGame.Turn++
		response.YourTurn = false
		response.Turn = activeGame.Turn
	}

	otherClient := OtherClient(activeGame, req.UserID)
	otherClientResponse := types.Request{
		YourTurn: true,
		GameID: response.GameID,
		UserID: response.UserID,
		Action: response.Action,
		Success: response.Success,
		GameOver: response.GameOver,
		Data: response.Data,
		Turn: response.Turn,
		Colors: response.Colors,
		Board: response.Board,
		Home: response.Home,
	}

	return []SocketClientResponse{
		SocketClientResponse{
			socketClient,
			response,
		},
		SocketClientResponse{
			otherClient,
			otherClientResponse,
		},
	}
}

// HandleRequest handles a request
func (server *Server) HandleRequest(req types.Request, socketClient *types.SocketClient) {
	server.printString("Request received [userID "+req.UserID+", action "+req.Action+"]")

	activeGame := server.games[req.GameID]
	if activeGame != nil {
		activeGame.M.Lock()
		defer activeGame.M.Unlock()
	}

	socketClientResponses := []SocketClientResponse{}
	switch action := req.Action; action {
	case constants.CREATE:
		socketClientResponses = server.handleCreate(req, socketClient)
	case constants.JOIN:
		socketClientResponses = server.handleJoin(req, socketClient, activeGame)
	case constants.MESSAGE:
		socketClientResponses = server.handleMessage(req, socketClient, activeGame)
	case constants.MOVE:
		socketClientResponses = server.handleMove(req, socketClient, activeGame)
	case constants.HOME:
		socketClientResponses = server.handleSendToHome(socketClient)
	default:
		server.printString("Unrecognized action: "+req.Action)
	}

	for _, socketClientResponse := range socketClientResponses {
		socketClientResponse.send(server.DisablePrint)
	}
}

func (server *Server) handleSendToHome(socketClient *types.SocketClient) []SocketClientResponse {
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

	response := types.Request{
		Action: constants.HOME,
		Home:   home,
	}

	return []SocketClientResponse{
		SocketClientResponse{
			socketClient,
			response,
		},
	}
}

func (server *Server) Stop() {
	close(server.quit)
	server.listener.Close()
	server.wg.Wait()
}

// Listen starts the server
func (server *Server) Listen(port string) {
	server.printString("Starting server...")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	server.listener = listener
	server.wg.Add(1)
	defer server.wg.Done()
	
	// create socket socketClient manager
	socketClientManager := SocketClientManager{
		server: 	server,
		clients:    make(map[*types.SocketClient]bool),
		register:   make(chan *types.SocketClient),
		unregister: make(chan *types.SocketClient),
	}

	go socketClientManager.start()

	server.printString("Server listening on port " + port + "!")

	for {
		connection, err := listener.Accept()
		if err != nil {
			select {
			case <-server.quit:
			  return
			default:
			  server.printError(err)
			}
		} else {
			server.wg.Add(1)
			socketClient := &types.SocketClient{Socket: connection, Data: make(chan []byte)}
			socketClientManager.register <- socketClient
			go func() {
				socketClientManager.receive(socketClient, server)
				server.wg.Done()
			}()
			go socketClientManager.send(socketClient)
	
			socketClientResponses := server.handleSendToHome(socketClient)
			for _, socketClientResponse := range socketClientResponses {
				go socketClientResponse.send(server.DisablePrint)
			}
		}
	}
}

// SocketClientManager handles all clients
type SocketClientManager struct {
	server	   *Server
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
	manager.server.printString("SocketClient manager listening for clients joining/leaving...")
	for {
		select {
		case socketClient := <-manager.register:
			manager.clients[socketClient] = true
			manager.server.printString("Added new socketClient!")
		case socketClient := <-manager.unregister:
			if _, ok := manager.clients[socketClient]; ok {
				manager.server.printString("A socketClient has left!")
				close(socketClient.Data)
				delete(manager.clients, socketClient)
			}
		}
	}
}
