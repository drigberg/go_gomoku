package helpers

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"go_gomoku/types"
	"go_gomoku/util"
	"go_gomoku/constants"
)

var (
	axes = [4][2]int{[2]int{-1, -1}, [2]int{-1, 0}, [2]int{-1, 1}, [2]int{0, -1}}
	TOP_LEFT = "\u250F"
	BOTTOM_LEFT = "\u2517"
	TOP_RIGHT = "\u2513"
	BOTTOM_RIGHT = "\u251B"
	HORIZONTAL = "\u2500"
	HORIZONTAL2 = "\u2500\u2500"
	HORIZONTAL3 = "\u2500\u2500\u2500"
	SPACE = " "
	SPACE2 = "  "
	SPACE3 = "   "
	VERTICAL = "\u2503"
	BOTTOM_INTERSECTION = "\u253B"
	RIGHT_INTERSECTION = "\u252B"
	TOP_INTERSECTION = "\u2533"
	LEFT_INTERSECTION = "\u2523"
	FULL_INTERSECTION = "\u254B"
	BLACK = "\u25C9"
	WHITE = "\u25EF"
	HORIZONTAL_AFTER_PIECE = SPACE
	HORIZONTAL2_AFTER_PIECE = SPACE + HORIZONTAL
	HORIZONTAL3_AFTER_PIECE = SPACE + HORIZONTAL2
	COLORS map[string]string
)
func CheckOwnership(game *types.GameRoom, userId string, move types.Coord) (bool, types.Request) {
	if util.IsTakenBy(game.Board, move) != constants.FREE {
			errorResponse := types.Request{
					GameId: game.Id,
					UserId: userId,
					Action: constants.MOVE,
					Data: "That spot is already taken!",
					Success: false,
			}

			return false, errorResponse
	}

	return true, types.Request{}
}

func CheckAlongAxis(spots map[string]bool, axis [2]int, move types.Coord, num int) int {
	next := types.Coord{
		X: move.X + axis[0],
		Y: move.Y + axis[1],
	}

	for c := range(spots) {
		coordinates := strings.Split(c, " ")
		x, _ := strconv.Atoi(coordinates[0])
		y, _ := strconv.Atoi(coordinates[1])

		if x == next.X && y == next.Y {
			return CheckAlongAxis(spots, axis, next, num + 1)
		}
	}

	return num
}

func CheckForWin(game *types.GameRoom, move types.Coord, color string) (bool) {
	// check within color from move coordinates
	spots := game.Board[color]

	for _, axis := range(axes) {
		fmt.Println("\n")

		len := CheckAlongAxis(spots, axis, move, 1)
		fmt.Println(len)
		complement := [2]int{axis[0] * -1, axis[1] * -1}
		len = CheckAlongAxis(spots, complement, move, len)
		fmt.Println(len)


		if len == 5 {
			return true
		}
	}

	return false
}

func GetOpponentId(game *types.GameRoom, userId string) string {
	for id := range(game.Players) {
			if id != userId {
					return id
			}
	}
	return ""
}

func OtherClient(game *types.GameRoom, userId string) *types.Client {
	opponentId := GetOpponentId(game, userId)

	return game.Players[opponentId].Client
}

func IsTurn(game *types.GameRoom, userId string) bool {
	if (userId == game.FirstPlayerId) {
			return game.Turn % 2 == 1
	}
	return game.Turn % 2 == 0
}

func SendBackoff(data []byte, client *types.Client, i int) {
	if client.Closed {
		return
	}

	select {
	case client.Data <- data:
		return
	default:
		time.Sleep(500 * time.Millisecond)

		fmt.Println("Trying again!", i)
		if (i > 5) {
			return
		}
		SendBackoff(data, client, i + 1)
	}
}

func intersectionOrSpace(horizontal string, intersection string, occupied string, spaceFirst bool) string {
	SPACE := intersection
	if occupied != constants.FREE {
		SPACE = COLORS[occupied]
	}

	if spaceFirst {
		return SPACE + horizontal
	}

	return horizontal + SPACE
}

func getRowChar(x int, y int, occupied string, prevOccupied string) string {
	HORIZONTALS := [3]string{HORIZONTAL, HORIZONTAL2, HORIZONTAL3}

	if prevOccupied != constants.FREE {
		HORIZONTALS[0] = HORIZONTAL_AFTER_PIECE
		HORIZONTALS[1] = HORIZONTAL2_AFTER_PIECE
		HORIZONTALS[2] = HORIZONTAL3_AFTER_PIECE
	}

  if y == 1 {
    switch x {
    case 1:
			return intersectionOrSpace(HORIZONTALS[1], TOP_LEFT, occupied, true)
    case 27:
			return intersectionOrSpace(HORIZONTALS[2], TOP_RIGHT, occupied, false)
		default:
			if x % 2 == 0 {
				return intersectionOrSpace(HORIZONTALS[0], TOP_INTERSECTION, occupied, false)
			}

			return HORIZONTALS[1]
    }
  }

  if y == 29 {
    switch x {
    case 1:
			return intersectionOrSpace(HORIZONTALS[1], BOTTOM_LEFT, occupied, true)
    case 27:
			return intersectionOrSpace(HORIZONTALS[2], BOTTOM_RIGHT, occupied, false)
		default:
			if x % 2 == 0 {
				return intersectionOrSpace(HORIZONTALS[0], BOTTOM_INTERSECTION, occupied, false)
			}

			return HORIZONTALS[1]
    }
  }

  switch x {
		case 1:
			return intersectionOrSpace(HORIZONTALS[1], LEFT_INTERSECTION, occupied, true)
		case 27:
			return intersectionOrSpace(HORIZONTALS[2], RIGHT_INTERSECTION, occupied, false)
		default:
			if x % 2 == 0 {
				return intersectionOrSpace(HORIZONTALS[0], FULL_INTERSECTION, occupied, false)
			}

			return HORIZONTALS[1]
    }
}


func getColumnChar(x int, y int) string {
  if (x == 1) {
    return VERTICAL + SPACE2
  }

  if (x == 27) {
    return SPACE3 + VERTICAL
  }

	if x % 2 == 0 {
		return SPACE + VERTICAL
	}

  return SPACE2
}


func GetCoord(x int, y int) types.Coord {
	coord := types.Coord {
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

	if (y + 1) % 2 == 0 && coord.Y != 0 {
		coord.X = (y + 1) / 2
		return coord
	}

	if coord.X != 0 && x % 2 == 0 {
		coord.Y = (x / 2) + 1
		return coord
	}

	// convert to visual coords
	if (y + 1) % 2 == 0 && x % 2 == 0 {
		coord.X = (y + 1) / 2
		coord.Y = (x / 2) + 1

		return coord
	}

	return coord
}

func GetAxisLabel(x int, y int) string {
	// x axis
	ret := ""
	if y == 0 {
		if x == 1 {
			ret += SPACE3
		}

		if x % 2 == 1 {
			char := strconv.Itoa(((x - 1) / 2) + 1)

			ret += char + SPACE

			if x == 27 {
				ret += " 15"
				return ret
			}
		} else {
			if x >= 20 {
				ret += SPACE
				return ret
			}

			ret += SPACE2
		}
		return ret
	}

	// y axis
	if x == 1 {
		if y % 2 == 1 {
			char := strconv.Itoa(((y - 1) / 2) + 1)

			if y >= 19 {
				ret += char + SPACE
			} else {
				ret += char + SPACE2
			}
		} else {
			ret += SPACE3
		}
		return ret
	}

	return ret
}

func PrintBoard(board map[string]map[string]bool) {
	prevOccupied := constants.FREE
	occupied := constants.FREE

	for y := 0; y <= 29; y++ {
		row := ""
		for x := 1; x <= 27; x++ {
			label := GetAxisLabel(x, y)
			row += label

			if y == 0 && label != "" {
				continue
			}

			coord := GetCoord(x, y)

			occupied = util.IsTakenBy(board, coord)

			if (y % 2 == 1) {
				row += getRowChar(x, y, occupied, prevOccupied)
			} else {
				row += getColumnChar(x, y)
			}
			prevOccupied = occupied
		}

		fmt.Println(row)
	}
}

func InitMaps() {
	COLORS = make(map[string]string)
	COLORS["white"] = WHITE
	COLORS["black"] = BLACK
}

func SendToClient(request types.Request, client *types.Client) {
	data, err := util.GobToBytes(request)

	if err != nil {
			fmt.Println(err)
			return
	}

	SendBackoff(data, client, 1)
}