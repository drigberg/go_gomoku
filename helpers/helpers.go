package helpers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"go_gomoku/constants"
	"go_gomoku/types"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	axes                  = [4][2]int{[2]int{-1, -1}, [2]int{-1, 0}, [2]int{-1, 1}, [2]int{0, -1}}
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
	colors                map[string]string
)

func clearWindows() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func clearLinuxOrDarwin() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// ClearScreen clears the screen, based on operating system
func ClearScreen() {
	switch runtime.GOOS {
	case "linux":
	case "darwin":
		clearLinuxOrDarwin()
	case "windows":
		clearWindows()
	default:
		panic("Your platform is unsupported! Can't clear the terminal screen.")
	}
}

// DecodeGob builds a request from a gob
func DecodeGob(message []byte) types.Request {
	var network bytes.Buffer
	network.Write(message)
	var request types.Request

	dec := gob.NewDecoder(&network)

	err := dec.Decode(&request)

	if err != nil {
		fmt.Println("Error decoding GOB data:", err)
	}

	return request
}

// GobToBytes converts a gob to bytes
func GobToBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isTakenBy(board map[string]map[string]bool, move types.Coord) string {
	spotStr := move.String()

	for color := range board {
		if board[color][spotStr] {
			return color
		}
	}

	return constants.FREE
}

// CheckOwnership checks whether a spot has been taken already
func CheckOwnership(game *types.GameRoom, userID string, move types.Coord) (bool, types.Request) {
	if isTakenBy(game.Board, move) != constants.FREE {
		errorResponse := types.Request{
			GameID:  game.ID,
			UserID:  userID,
			Action:  constants.MOVE,
			Data:    "That spot is already taken!",
			Success: false,
		}

		return false, errorResponse
	}

	return true, types.Request{}
}

func checkAlongAxis(spots map[string]bool, axis [2]int, move types.Coord, num int) int {
	next := types.Coord{
		X: move.X + axis[0],
		Y: move.Y + axis[1],
	}

	for c := range spots {
		coordinates := strings.Split(c, " ")
		x, _ := strconv.Atoi(coordinates[0])
		y, _ := strconv.Atoi(coordinates[1])

		if x == next.X && y == next.Y {
			return checkAlongAxis(spots, axis, next, num+1)
		}
	}

	return num
}

// CheckForWin looks for five in a row
func CheckForWin(game *types.GameRoom, move types.Coord, color string) bool {
	// check within color from move coordinates
	spots := game.Board[color]

	for _, axis := range axes {
		fmt.Print("\n\n")

		len := checkAlongAxis(spots, axis, move, 1)
		fmt.Println(len)
		complement := [2]int{axis[0] * -1, axis[1] * -1}
		len = checkAlongAxis(spots, complement, move, len)
		fmt.Println(len)

		if len == 5 {
			return true
		}
	}

	return false
}

// GetOpponentID returns the other player's id
func GetOpponentID(game *types.GameRoom, userID string) string {
	for id := range game.Players {
		if id != userID {
			return id
		}
	}
	return ""
}

// OtherClient returns the other player's client
func OtherClient(game *types.GameRoom, userID string) *types.Client {
	opponentID := GetOpponentID(game, userID)

	return game.Players[opponentID].Client
}

// IsTurn turns if it's a user's turn or not
func IsTurn(game *types.GameRoom, userID string) bool {
	if userID == game.FirstPlayerId {
		return game.Turn%2 == 1
	}
	return game.Turn%2 == 0
}

// SendBackoff tries to send and keeps trying
func SendBackoff(data []byte, client *types.Client, i int) {
	if i > 1 {
		fmt.Println("Retrying message send! Attempt:", i)
	}
	if client.Closed {
		return
	}

	select {
	case client.Data <- data:
		return
	default:
		time.Sleep(500 * time.Millisecond)

		if i > 5 {
			return
		}
		SendBackoff(data, client, i+1)
	}
}

func intersectionOrSpace(horizontal string, intersection string, occupied string, spaceFirst bool) string {
	space := intersection
	if occupied != constants.FREE {
		space = colors[occupied]
	}

	if spaceFirst {
		return space + horizontal
	}

	return horizontal + space
}

func getRowChar(x int, y int, occupied string, prevOccupied string) string {
	HORIZONTALS := [3]string{horizontal, horizontal2, horizontal3}

	if prevOccupied != constants.FREE {
		HORIZONTALS[0] = horizontalAfterPiece
		HORIZONTALS[1] = horizontal2AfterPiece
		HORIZONTALS[2] = horizontal3AfterPiece
	}

	if y == 1 {
		switch x {
		case 1:
			return intersectionOrSpace(HORIZONTALS[1], topLeft, occupied, true)
		case 27:
			return intersectionOrSpace(HORIZONTALS[2], topRight, occupied, false)
		default:
			if x%2 == 0 {
				return intersectionOrSpace(HORIZONTALS[0], topIntersection, occupied, false)
			}

			return HORIZONTALS[1]
		}
	}

	if y == 29 {
		switch x {
		case 1:
			return intersectionOrSpace(HORIZONTALS[1], bottomLeft, occupied, true)
		case 27:
			return intersectionOrSpace(HORIZONTALS[2], bottomRight, occupied, false)
		default:
			if x%2 == 0 {
				return intersectionOrSpace(HORIZONTALS[0], bottomIntersection, occupied, false)
			}

			return HORIZONTALS[1]
		}
	}

	switch x {
	case 1:
		return intersectionOrSpace(HORIZONTALS[1], leftIntersection, occupied, true)
	case 27:
		return intersectionOrSpace(HORIZONTALS[2], rightIntersection, occupied, false)
	default:
		if x%2 == 0 {
			return intersectionOrSpace(HORIZONTALS[0], fullIntersection, occupied, false)
		}

		return HORIZONTALS[1]
	}
}

func getColumnChar(x int, y int) string {
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

// GetCoord converts from the character grid locations to board locations
func GetCoord(x int, y int) types.Coord {
	coord := types.Coord{
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

func getAxisLabel(x int, y int) string {
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

// PrintBoard prints the current state of the board
func PrintBoard(board map[string]map[string]bool) {
	prevOccupied := constants.FREE
	occupied := constants.FREE

	for y := 0; y <= 29; y++ {
		row := ""
		for x := 1; x <= 27; x++ {
			label := getAxisLabel(x, y)
			row += label

			if y == 0 && label != "" {
				continue
			}

			coord := GetCoord(x, y)

			occupied = isTakenBy(board, coord)

			if y%2 == 1 {
				row += getRowChar(x, y, occupied, prevOccupied)
			} else {
				row += getColumnChar(x, y)
			}
			prevOccupied = occupied
		}

		fmt.Println(row)
	}
}

// InitMaps initializes maps
func InitMaps() {
	colors = make(map[string]string)
	colors["white"] = white
	colors["black"] = black
}

// SendToClient tries to send a request to client, with backoff
func SendToClient(request types.Request, client *types.Client) {
	data, err := GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	SendBackoff(data, client, 1)
}
