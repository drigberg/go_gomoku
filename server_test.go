package main

import (
	"testing"
)

type ParseCoordsTestCase struct {
	input         string
	expectedMove Coord
}

func getValidParseCoordsTestCases() []ParseCoordsTestCase {
	testcases := []ParseCoordsTestCase{
		ParseCoordsTestCase{
			"3 3",
			Coord{X: 3, Y: 3},
		},
		ParseCoordsTestCase{
			"4 4",
			Coord{X: 4, Y: 4},
		},
		ParseCoordsTestCase{
			"1 1",
			Coord{X: 1, Y: 1},
		},
		ParseCoordsTestCase{
			"15 15",
			Coord{X: 15, Y: 15},
		},
		ParseCoordsTestCase{
			"1 15",
			Coord{X: 1, Y: 15},
		},
		ParseCoordsTestCase{
			"15 1",
			Coord{X: 15, Y: 1},
		},
	}
	return testcases
}


func getInvalidParseCoordsTestCasesOffBoardOrNonNumber() []string {
	testcases := []string{
		"0 3",
		"16 4",
		"1 -1",
		"3 16",
		"0 0",
		"16 16",
		"100 3",
		"0 a",
		"b 4",
		"-- -1",
		"3 *",
		"% $",
	}
	return testcases
}

func getParseCoordsTestCasesInvalidSyntax() []string {
	testcases := []string{
		"16",
		"100 3 4",
		"' 3 3",
		"33",
	}
	return testcases
}

func TestServerParseCoordsValid(t *testing.T) {
	newServer := NewServer()
	gameID := 3
	player1 := Player{
		UserID:       "mock_user1",
		Color: "white",

	}
	player2 := Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	newServer.games[gameID] = &GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   NewBoard(),
	}

	testcases := getValidParseCoordsTestCases()
	for _, testcase := range testcases {
		req := Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, move, errorResponse := newServer.parseMove(req, testcase.input)
		if !is_valid {
			t.Errorf("Expected move to be valid")
		}
		if move.X != testcase.expectedMove.X && move.Y != testcase.expectedMove.Y {
			t.Errorf("Expected move to be %s, got %s", testcase.expectedMove, move)
		}
		if errorResponse.Data != "" {
			t.Errorf("Got error: %s", errorResponse.Data)
		}
	}
}

func TestServerParseCoordsInvalidTaken(t *testing.T) {
	newServer := NewServer()
	gameID := 3
	player1 := Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   NewBoard(),
	}
	newServer.games[gameID] = &game

	testcases := getValidParseCoordsTestCases()
	for _, testcase := range testcases {
		game.PlayMove(testcase.expectedMove, game.Players[player1.UserID].Color)
		req := Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.parseMove(req, testcase.input)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expectedMessage := "That spot is already taken!"
		if errorResponse.Data != expectedMessage {
			t.Errorf("Expected message to be %s, got %s", expectedMessage, errorResponse.Data)
		}
	}
}

func TestServerParseCoordsInvalidOffBoard(t *testing.T) {
	newServer := NewServer()
	gameID := 3
	player1 := Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   NewBoard(),
	}
	newServer.games[gameID] = &game

	testcases := getInvalidParseCoordsTestCasesOffBoardOrNonNumber()
	for _, testcase := range testcases {
		req := Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.parseMove(req, testcase)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expectedMessage := "Both x and y must be integers from 1 to 15"
		if errorResponse.Data != expectedMessage {
			t.Errorf("Expected message to be %s, got %s", expectedMessage, errorResponse.Data)
		}
	}
}

func TestServerParseCoordsInvalidSyntax(t *testing.T) {
	newServer := NewServer()
	gameID := 3
	player1 := Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   NewBoard(),
	}
	newServer.games[gameID] = &game

	testcases := getParseCoordsTestCasesInvalidSyntax()
	for _, testcase := range testcases {
		req := Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.parseMove(req, testcase)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expectedMessage := "The syntax for a move is mv <x> <y>"
		if errorResponse.Data != expectedMessage {
			t.Errorf("Expected message to be %s, got %s", expectedMessage, errorResponse.Data)
		}
	}
}

func TestServerHandleCreate(t *testing.T) {
	playerID := "mock_player_1"
	newServer := NewServer()
	req := Request{
		UserID: playerID,
	}
	socketClient := SocketClient{}
	socketClientResponses := newServer.handleCreate(req, &socketClient)
	if len(socketClientResponses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(socketClientResponses))
	}
	response := socketClientResponses[0].response

	if response.Action != CREATE {
		t.Errorf("Expected response to be type CREATE, got %s", response.Action)
	}

	gameID := response.GameID
	game := newServer.games[gameID]
	if game == nil {
		t.Errorf("Expected to find game by id from response")
	}

	if game.ID != gameID {
		t.Errorf("Expected game to have id %d, got %d", gameID, game.ID)
	}

	if len(game.Players) != 1 {
		t.Errorf("Expected game to have 1 player, got %d", len(game.Players))
	}

	player := game.Players[playerID]

	if player == nil {
		t.Errorf("Expected to find player by ID %s", playerID)
	}

	if game.Turn != 0 {
		t.Errorf("Expected game to have turn 0, got %d", game.Turn)
	}
}

func TestServerHandleJoinSuccess(t *testing.T) {
	newServer := NewServer()
	createRequest := Request{
		UserID: "mock_player_1",
	}

	socketClient := SocketClient{}
	socketClientResponsesCreate := newServer.handleCreate(createRequest, &socketClient)
	gameID := socketClientResponsesCreate[0].response.GameID
	game := newServer.games[gameID]

	joinRequest := Request{
		UserID: "mock_player_2",
	}

	otherSocketClient := SocketClient{}
	socketClientResponsesJoin := newServer.handleJoin(joinRequest, &otherSocketClient, game)
	if len(socketClientResponsesJoin) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(socketClientResponsesJoin))
	}

	responses := []Request{socketClientResponsesJoin[0].response, socketClientResponsesJoin[1].response}

	for i, response := range responses {
		if response.GameID != gameID {
			t.Errorf("Expected response %d to have gameID %d, got %d", i, gameID, response.GameID)
		}

		expectedYourTurn := game.FirstPlayerID != response.UserID
		if response.YourTurn != expectedYourTurn {
			t.Errorf("Expected response %d for player %s to have YourTurn %t, got %t", i, response.UserID, expectedYourTurn, response.YourTurn)
		}
		if response.UserID == joinRequest.UserID && response.Action != OTHERJOINED {
			t.Errorf("Expected response %d to have action OTHERJOINED, got %s", i, OTHERJOINED)
		}
		if response.UserID == createRequest.UserID && response.Action != JOIN {
			t.Errorf("Expected response %d to have action JOIN, got %s", i, JOIN)
		}
		if response.Success != true {
			t.Errorf("Expected response %d to have success=true", i)
		}
		if response.Turn != 1 {
			t.Errorf("Expected response %d to have turn 1, got %d", i, response.Turn)
		}
	}
}

func TestServerHandleJoinAlreadyInRoom(t *testing.T) {
	newServer := NewServer()
	createRequest := Request{
		UserID: "mock_player_1",
	}

	socketClient := SocketClient{}
	socketClientResponsesCreate := newServer.handleCreate(createRequest, &socketClient)
	gameID := socketClientResponsesCreate[0].response.GameID
	game := newServer.games[gameID]

	joinRequest := Request{
		UserID: "mock_player_1",
	}

	otherSocketClient := SocketClient{}
	socketClientResponsesJoin := newServer.handleJoin(joinRequest, &otherSocketClient, game)
	if len(socketClientResponsesJoin) != 1 {
		t.Errorf("Expected 1 response, got %d", len(socketClientResponsesJoin))
	}

	response := socketClientResponsesJoin[0].response

	if response.Action != JOIN {
		t.Errorf("Expected response to be type JOIN, got %s", response.Action)
	}

	expectedMessage := "You are already in this room, and you can't go back! Sorry!"
	if response.Data != expectedMessage {
		t.Errorf("Expected message to be %s, got %s", expectedMessage, response.Data)
	}
}
