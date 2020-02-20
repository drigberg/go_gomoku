package main

import (
	"go_gomoku/board"
	"go_gomoku/types"
	"testing"
)

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
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			coord := types.Coord{X: x, Y: y}
			if gameBoard.CheckForWin(coord, "white") {
				t.Errorf("Expected CheckForWin to return false for white at coord (%d, %d)", x, y)
			}
			if gameBoard.CheckForWin(coord, "black") {
				t.Errorf("Expected CheckForWin to return false for black at coord (%d, %d)", x, y)
			}
		}
	}

}
