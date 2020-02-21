package main

import (
	"go_gomoku/board"
	"go_gomoku/types"
	"testing"
)

func coordsAreEqual(coord1 *types.Coord, coord2 *types.Coord) bool {
	return coord1.X == coord2.X && coord1.Y == coord2.Y
}

func coordSlicesAreEqual(slice1 *[]types.Coord, slice2 *[]types.Coord) bool {
	if len(*slice1) != len(*slice2) {
		return false
	}

	for i := 0; i < len(*slice1); i++ {
		coord1 := (*slice1)[i]
		coord2 := (*slice2)[i]
		if !coordsAreEqual(&coord1, &coord2) {
			return false
		}
	}
	return true
}

func getWinningCoordsSlices(gameBoard *board.Board) (*[]types.Coord, *[]types.Coord) {
	winningCoordsWhite := []types.Coord{}
	winningCoordsBlack := []types.Coord{}
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			coord := types.Coord{X: x, Y: y}
			if (*gameBoard).CheckForWin(coord, "white") {
				winningCoordsWhite = append(winningCoordsWhite, coord)
			}
			if (*gameBoard).CheckForWin(coord, "black") {
				winningCoordsBlack = append(winningCoordsBlack, coord)
			}
		}
	}
	return &winningCoordsWhite, &winningCoordsBlack
}

func TestBoardNew(t *testing.T) {
	gameBoard := board.New()
	if gameBoard.Spaces == nil {
		t.Error("Expected board.Spaces to not be nil")
	}
	keys := make([]string, 0, len(gameBoard.Spaces))
	for k := range gameBoard.Spaces {
		keys = append(keys, k)
	}
	if len(keys) != 2 {
		t.Errorf("Expected board.Spaces to have 2 keys, found %d", len(keys))
	}

	if gameBoard.Spaces["white"] == nil {
		t.Error("Expected board.Spaces['white'] to not be nil")
	}

	if gameBoard.Spaces["black"] == nil {
		t.Error("Expected board.Spaces['black'] to not be nil")
	}
}

func TestBoardCheckForWinEmpty(t *testing.T) {
	gameBoard := board.New()
	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)
	if len(*winningCoordsWhite) > 0 {
		t.Errorf("Expected no winning spaces for white, found %d", len(*winningCoordsWhite))
	}
	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}
}

func TestBoardCheckForWinSuccessWhiteVertical(t *testing.T) {
	gameBoard := board.New()
	gameBoard.Spaces["white"]["3 3"] = true
	gameBoard.Spaces["white"]["3 4"] = true
	gameBoard.Spaces["white"]["3 5"] = true
	gameBoard.Spaces["white"]["3 6"] = true
	gameBoard.Spaces["white"]["3 7"] = true
	expectedWinningCoordsWhite := []types.Coord{
		types.Coord{X: 3, Y: 3},
		types.Coord{X: 3, Y: 4},
		types.Coord{X: 3, Y: 5},
		types.Coord{X: 3, Y: 6},
		types.Coord{X: 3, Y: 7},
	}

	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}

	if !coordSlicesAreEqual(winningCoordsWhite, &expectedWinningCoordsWhite) {
		t.Errorf("Expected winning spaces for white to be %s, got %s", expectedWinningCoordsWhite, winningCoordsWhite)
	}
}

func TestBoardCheckForWinSuccessWhiteHorizontal(t *testing.T) {
	gameBoard := board.New()
	gameBoard.Spaces["white"]["3 3"] = true
	gameBoard.Spaces["white"]["4 3"] = true
	gameBoard.Spaces["white"]["5 3"] = true
	gameBoard.Spaces["white"]["6 3"] = true
	gameBoard.Spaces["white"]["7 3"] = true

	expectedWinningCoordsWhite := []types.Coord{
		types.Coord{X: 3, Y: 3},
		types.Coord{X: 4, Y: 3},
		types.Coord{X: 5, Y: 3},
		types.Coord{X: 6, Y: 3},
		types.Coord{X: 7, Y: 3},
	}

	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}

	if !coordSlicesAreEqual(winningCoordsWhite, &expectedWinningCoordsWhite) {
		t.Errorf("Expected winning spaces for white to be %s, got %s", expectedWinningCoordsWhite, winningCoordsWhite)
	}
}

func TestBoardCheckForWinSuccessWhiteDiagonalDownRight(t *testing.T) {
	gameBoard := board.New()
	gameBoard.Spaces["white"]["3 3"] = true
	gameBoard.Spaces["white"]["4 4"] = true
	gameBoard.Spaces["white"]["5 5"] = true
	gameBoard.Spaces["white"]["6 6"] = true
	gameBoard.Spaces["white"]["7 7"] = true

	expectedWinningCoordsWhite := []types.Coord{
		types.Coord{X: 3, Y: 3},
		types.Coord{X: 4, Y: 4},
		types.Coord{X: 5, Y: 5},
		types.Coord{X: 6, Y: 6},
		types.Coord{X: 7, Y: 7},
	}
	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}

	if !coordSlicesAreEqual(winningCoordsWhite, &expectedWinningCoordsWhite) {
		t.Errorf("Expected winning spaces for white to be %s, got %s", expectedWinningCoordsWhite, winningCoordsWhite)
	}
}

func TestBoardCheckForWinSuccessWhiteDiagonalDownLeft(t *testing.T) {
	gameBoard := board.New()
	gameBoard.Spaces["white"]["7 3"] = true
	gameBoard.Spaces["white"]["6 4"] = true
	gameBoard.Spaces["white"]["5 5"] = true
	gameBoard.Spaces["white"]["4 6"] = true
	gameBoard.Spaces["white"]["3 7"] = true

	expectedWinningCoordsWhite := []types.Coord{
		types.Coord{X: 3, Y: 7},
		types.Coord{X: 4, Y: 6},
		types.Coord{X: 5, Y: 5},
		types.Coord{X: 6, Y: 4},
		types.Coord{X: 7, Y: 3},
	}
	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}

	if !coordSlicesAreEqual(winningCoordsWhite, &expectedWinningCoordsWhite) {
		t.Errorf("Expected winning spaces for white to be %s, got %s", expectedWinningCoordsWhite, winningCoordsWhite)
	}
}
