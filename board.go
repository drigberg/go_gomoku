package main

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	axes                  = [4][2]int{[2]int{-1, -1}, [2]int{-1, 0}, [2]int{-1, 1}, [2]int{0, -1}}
)

const (
	topLeft               = "\u250F"
	bottomLeft            = "\u2517"
	topRight              = "\u2513"
	bottomRight           = "\u251B"
	horizontal            = "\u2500"
	horizontal2           = "\u2500\u2500"
	horizontal3           = "\u2500\u2500\u2500"
	space                 = " "
	space2                = "  "
	space3                = "   "
	vertical              = "\u2503"
	bottomIntersection    = "\u253B"
	rightIntersection     = "\u252B"
	topIntersection       = "\u2533"
	leftIntersection      = "\u2523"
	fullIntersection      = "\u254B"
	black                 = "\u25C9"
	white                 = "\u25EF"
	horizontalAfterPiece  = space
	horizontal2AfterPiece = space + horizontal
	horizontal3AfterPiece = space + horizontal2
)

// Board contains the state of the game board
type Board struct {
	Spaces map[string]map[string]bool
}

// BoardInterface defines methods a Board should implement
type BoardInterface interface {
	checkForWin(Coord, string) bool
	checkOwnership(int, string, Coord) (bool, Request)
	printBoard()
	checkAlongAxis(string, [2]int, Coord, int) int
	getAxisLabel(int, int) string
	getColorCode(string) string
	getColumnChar(int, int) string
	getCoord(int, int) Coord
	getRowChar(int, int, string, string) string
	intersectionOrSpace(string, string, string, bool) string
	isTakenBy(Coord) string
}

// assert that Board implements Interface
var _ BoardInterface = (*Board)(nil)

// New creates an empty board
func NewBoard() Board {
	spaces := make(map[string]map[string]bool)
	spaces["white"] = make(map[string]bool)
	spaces["black"] = make(map[string]bool)
	return Board{
		Spaces: spaces,
	}
}

func (board *Board) printBoard() {
	prevOccupied := FREE
	occupied := FREE

	for y := 0; y <= 29; y++ {
		row := ""
		for x := 1; x <= 27; x++ {
			label := board.getAxisLabel(x, y)
			row += label

			if y == 0 && label != "" {
				continue
			}

			coord := board.getCoord(x, y)

			occupied = board.isTakenBy(coord)

			if y%2 == 1 {
				row += board.getRowChar(x, y, occupied, prevOccupied)
			} else {
				row += board.getColumnChar(x, y)
			}
			prevOccupied = occupied
		}

		fmt.Println(row)
	}
}

func (board *Board) checkAlongAxis(color string, axis [2]int, move Coord, num int) int {
	next := Coord{
		X: move.X + axis[0],
		Y: move.Y + axis[1],
	}

	for c := range board.Spaces[color] {
		coordinates := strings.Split(c, " ")
		x, _ := strconv.Atoi(coordinates[0])
		y, _ := strconv.Atoi(coordinates[1])

		if x == next.X && y == next.Y {
			return board.checkAlongAxis(color, axis, next, num+1)
		}
	}

	return num
}

func (board *Board) checkForWin(move Coord, color string) bool {
	// check within color from move coordinates
	for _, axis := range axes {
		len := board.checkAlongAxis(color, axis, move, 1)
		complement := [2]int{axis[0] * -1, axis[1] * -1}
		len = board.checkAlongAxis(color, complement, move, len)

		if len == 5 {
			return true
		}
	}

	return false
}

func (board *Board) isTakenBy(move Coord) string {
	spotStr := move.String()

	for color := range board.Spaces {
		if board.Spaces[color][spotStr] {
			return color
		}
	}

	return FREE
}

func (board *Board) getColorCode(color string) string {
	if color == "white" {
		return white
	}
	return black
}

func (board *Board) intersectionOrSpace(horizontal string, intersection string, occupied string, spaceFirst bool) string {
	space := intersection
	if occupied != FREE {
		space = board.getColorCode(occupied)
	}

	if spaceFirst {
		return space + horizontal
	}

	return horizontal + space
}

func (board *Board) checkOwnership(gameID int, userID string, move Coord) (bool, Request) {
	if board.isTakenBy(move) != FREE {
		errorResponse := Request{
			GameID:  gameID,
			UserID:  userID,
			Action:  MOVE,
			Data:    "That spot is already taken!",
			Success: false,
		}

		return false, errorResponse
	}

	return true, Request{}
}

func (board *Board) getRowChar(x int, y int, occupied string, prevOccupied string) string {
	HORIZONTALS := [3]string{horizontal, horizontal2, horizontal3}

	if prevOccupied != FREE {
		HORIZONTALS[0] = horizontalAfterPiece
		HORIZONTALS[1] = horizontal2AfterPiece
		HORIZONTALS[2] = horizontal3AfterPiece
	}

	if y == 1 {
		switch x {
		case 1:
			return board.intersectionOrSpace(HORIZONTALS[1], topLeft, occupied, true)
		case 27:
			return board.intersectionOrSpace(HORIZONTALS[2], topRight, occupied, false)
		default:
			if x%2 == 0 {
				return board.intersectionOrSpace(HORIZONTALS[0], topIntersection, occupied, false)
			}

			return HORIZONTALS[1]
		}
	}

	if y == 29 {
		switch x {
		case 1:
			return board.intersectionOrSpace(HORIZONTALS[1], bottomLeft, occupied, true)
		case 27:
			return board.intersectionOrSpace(HORIZONTALS[2], bottomRight, occupied, false)
		default:
			if x%2 == 0 {
				return board.intersectionOrSpace(HORIZONTALS[0], bottomIntersection, occupied, false)
			}

			return HORIZONTALS[1]
		}
	}

	switch x {
	case 1:
		return board.intersectionOrSpace(HORIZONTALS[1], leftIntersection, occupied, true)
	case 27:
		return board.intersectionOrSpace(HORIZONTALS[2], rightIntersection, occupied, false)
	default:
		if x%2 == 0 {
			return board.intersectionOrSpace(HORIZONTALS[0], fullIntersection, occupied, false)
		}

		return HORIZONTALS[1]
	}
}

func (board *Board) getColumnChar(x int, y int) string {
	if x == 1 {
		return vertical + space2
	}

	if x == 27 {
		return space3 + vertical
	}

	if x%2 == 0 {
		return space + vertical
	}

	return space2
}

func (board *Board) getCoord(x int, y int) Coord {
	coord := Coord{
		X: 0,
		Y: 0,
	}

	if x == 27 {
		coord.Y = 15
	}

	if x == 1 {
		coord.Y = 1
	}

	if y == 1 {
		coord.X = 1
	}

	if y == 29 {
		coord.X = 15
	}

	if coord.X != 0 && coord.Y != 0 {
		return coord
	}

	if (y+1)%2 == 0 && coord.Y != 0 {
		coord.X = (y + 1) / 2
		return coord
	}

	if coord.X != 0 && x%2 == 0 {
		coord.Y = (x / 2) + 1
		return coord
	}

	// convert to visual coords
	if (y+1)%2 == 0 && x%2 == 0 {
		coord.X = (y + 1) / 2
		coord.Y = (x / 2) + 1

		return coord
	}

	return coord
}

func (board *Board) getAxisLabel(x int, y int) string {
	// x axis
	ret := ""
	if y == 0 {
		if x == 1 {
			ret += space3
		}

		if x%2 == 1 {
			char := strconv.Itoa(((x - 1) / 2) + 1)

			ret += char + space

			if x == 27 {
				ret += " 15"
				return ret
			}
		} else {
			if x >= 20 {
				ret += space
				return ret
			}

			ret += space2
		}
		return ret
	}

	// y axis
	if x == 1 {
		if y%2 == 1 {
			char := strconv.Itoa(((y - 1) / 2) + 1)

			if y >= 19 {
				ret += char + space
			} else {
				ret += char + space2
			}
		} else {
			ret += space3
		}
		return ret
	}

	return ret
}
