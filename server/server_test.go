package server

import (
	"testing"
	"go_gomoku/board"
	"go_gomoku/types"
)

type ParseCoordsTestCase struct {
	input         string
	expectedMove types.Coord
}

func getValidParseCoordsTestCases() []ParseCoordsTestCase {
	testcases := []ParseCoordsTestCase{
		ParseCoordsTestCase{
			"3 3",
			types.Coord{X: 3, Y: 3},
		},
		ParseCoordsTestCase{
			"4 4",
			types.Coord{X: 4, Y: 4},
		},
		ParseCoordsTestCase{
			"1 1",
			types.Coord{X: 1, Y: 1},
		},
		ParseCoordsTestCase{
			"15 15",
			types.Coord{X: 15, Y: 15},
		},
		ParseCoordsTestCase{
			"1 15",
			types.Coord{X: 1, Y: 15},
		},
		ParseCoordsTestCase{
			"15 1",
			types.Coord{X: 15, Y: 1},
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
	newServer := New()
	gameID := 3
	player1 := types.Player{
		UserID:       "mock_user1",
		Color: "white",

	}
	player2 := types.Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*types.Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	newServer.games[gameID] = &GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   board.New(),
	}

	testcases := getValidParseCoordsTestCases()
	for _, testcase := range testcases {
		req := types.Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, move, errorResponse := newServer.ParseMove(req, testcase.input)
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
	newServer := New()
	gameID := 3
	player1 := types.Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := types.Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*types.Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   board.New(),
	}
	newServer.games[gameID] = &game

	testcases := getValidParseCoordsTestCases()
	for _, testcase := range testcases {
		game.PlayMove(testcase.expectedMove, game.Players[player1.UserID].Color)
		req := types.Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.ParseMove(req, testcase.input)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expected_message := "That spot is already taken!"
		if errorResponse.Data != expected_message {
			t.Errorf("Expected message to be %s, got %s", expected_message, errorResponse.Data)
		}
	}
}

func TestServerParseCoordsInvalidOffBoard(t *testing.T) {
	newServer := New()
	gameID := 3
	player1 := types.Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := types.Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*types.Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   board.New(),
	}
	newServer.games[gameID] = &game

	testcases := getInvalidParseCoordsTestCasesOffBoardOrNonNumber()
	for _, testcase := range testcases {
		req := types.Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.ParseMove(req, testcase)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expected_message := "Both x and y must be integers from 1 to 15"
		if errorResponse.Data != expected_message {
			t.Errorf("Expected message to be %s, got %s", expected_message, errorResponse.Data)
		}
	}
}

func TestServerParseCoordsInvalidSyntax(t *testing.T) {
	newServer := New()
	gameID := 3
	player1 := types.Player{
		UserID:       "mock_user1",
		Color: "white",
	}
	player2 := types.Player{
		UserID:       "mock_user2",
		Color: "black",
	}
	players := make(map[string]*types.Player)
	players[player1.UserID] = &player1
	players[player2.UserID] = &player2

	game := GameRoom{
		ID:      gameID,
		Players: players,
		Turn:    0,
		Board:   board.New(),
	}
	newServer.games[gameID] = &game

	testcases := getParseCoordsTestCasesInvalidSyntax()
	for _, testcase := range testcases {
		req := types.Request{
			GameID:  gameID,
			UserID:  player1.UserID,
		}
		is_valid, _, errorResponse := newServer.ParseMove(req, testcase)
		if is_valid {
			t.Errorf("Expected move to not be valid")
		}
		expected_message := "The syntax for a move is mv <x> <y>"
		if errorResponse.Data != expected_message {
			t.Errorf("Expected message to be %s, got %s", expected_message, errorResponse.Data)
		}
	}
}