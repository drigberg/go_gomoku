package main

import (
	"testing"
)

func coordsAreEqual(coord1 *Coord, coord2 *Coord) bool {
	return coord1.X == coord2.X && coord1.Y == coord2.Y
}

func coordSlicesAreEqual(slice1 *[]Coord, slice2 *[]Coord) bool {
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

func getWinningCoordsSlices(gameBoard *Board) (*[]Coord, *[]Coord) {
	winningCoordsWhite := []Coord{}
	winningCoordsBlack := []Coord{}
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			coord := Coord{X: x, Y: y}
			if (*gameBoard).checkForWin(coord, "white") {
				winningCoordsWhite = append(winningCoordsWhite, coord)
			}
			if (*gameBoard).checkForWin(coord, "black") {
				winningCoordsBlack = append(winningCoordsBlack, coord)
			}
		}
	}
	return &winningCoordsWhite, &winningCoordsBlack
}

func TestBoardNew(t *testing.T) {
	gameBoard := NewBoard()
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

func TestBoardcheckForWinEmpty(t *testing.T) {
	gameBoard := NewBoard()
	winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)
	if len(*winningCoordsWhite) > 0 {
		t.Errorf("Expected no winning spaces for white, found %d", len(*winningCoordsWhite))
	}
	if len(*winningCoordsBlack) > 0 {
		t.Errorf("Expected no winning spaces for black, found %d", len(*winningCoordsBlack))
	}
}

type BoardcheckForWinSuccessTestCase struct {
	label                 string
	coordStrings          []string
	expectedWinningCoords []Coord
}

func getcheckForWinTestcases() []BoardcheckForWinSuccessTestCase {
	testcases := []BoardcheckForWinSuccessTestCase{
		BoardcheckForWinSuccessTestCase{
			"horizontal",
			[]string{"3 3", "4 3", "5 3", "6 3", "7 3"},
			[]Coord{
				Coord{X: 3, Y: 3},
				Coord{X: 4, Y: 3},
				Coord{X: 5, Y: 3},
				Coord{X: 6, Y: 3},
				Coord{X: 7, Y: 3},
			},
		},
		BoardcheckForWinSuccessTestCase{
			"diagonal_right",
			[]string{"3 3", "4 4", "5 5", "6 6", "7 7"},
			[]Coord{
				Coord{X: 3, Y: 3},
				Coord{X: 4, Y: 4},
				Coord{X: 5, Y: 5},
				Coord{X: 6, Y: 6},
				Coord{X: 7, Y: 7},
			},
		},
		BoardcheckForWinSuccessTestCase{
			"vertical",
			[]string{"3 3", "3 4", "3 5", "3 6", "3 7"},
			[]Coord{
				Coord{X: 3, Y: 3},
				Coord{X: 3, Y: 4},
				Coord{X: 3, Y: 5},
				Coord{X: 3, Y: 6},
				Coord{X: 3, Y: 7},
			},
		},
		BoardcheckForWinSuccessTestCase{
			"diagonal_left",
			[]string{"7 3", "6 4", "5 5", "4 6", "3 7"},
			[]Coord{
				Coord{X: 3, Y: 7},
				Coord{X: 4, Y: 6},
				Coord{X: 5, Y: 5},
				Coord{X: 6, Y: 4},
				Coord{X: 7, Y: 3},
			},
		},
	}
	return testcases
}

func TestBoardcheckForWinSuccessWhite(t *testing.T) {
	testcases := getcheckForWinTestcases()
	for _, testcase := range testcases {
		t.Run(testcase.label, func(t *testing.T) {
			gameBoard := NewBoard()
			for _, coordString := range testcase.coordStrings {
				gameBoard.Spaces["white"][coordString] = true
			}

			winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

			if len(*winningCoordsBlack) > 0 {
				t.Errorf("Expected no winning spaces for black, found %d: %s", len(*winningCoordsBlack), winningCoordsBlack)
			}

			if !coordSlicesAreEqual(winningCoordsWhite, &testcase.expectedWinningCoords) {
				t.Errorf("Expected winning spaces for white to be %s, got %s", testcase.expectedWinningCoords, winningCoordsWhite)
			}
		})
	}
}

func TestBoardcheckForWinSuccessBlack(t *testing.T) {
	testcases := getcheckForWinTestcases()
	for _, testcase := range testcases {
		t.Run(testcase.label, func(t *testing.T) {
			gameBoard := NewBoard()
			for _, coordString := range testcase.coordStrings {
				gameBoard.Spaces["black"][coordString] = true
			}

			winningCoordsWhite, winningCoordsBlack := getWinningCoordsSlices(&gameBoard)

			if len(*winningCoordsWhite) > 0 {
				t.Errorf("Expected no winning spaces for black, found %d: %s", len(*winningCoordsWhite), winningCoordsBlack)
			}

			if !coordSlicesAreEqual(winningCoordsBlack, &testcase.expectedWinningCoords) {
				t.Errorf("Expected winning spaces for white to be %s, got %s", testcase.expectedWinningCoords, winningCoordsBlack)
			}
		})
	}
}
