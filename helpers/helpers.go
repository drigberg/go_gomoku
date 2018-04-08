package helpers

import (
	"fmt"
	"strings"
	"strconv"
	"go_gomoku/types"
	"go_gomoku/util"
	"go_gomoku/constants"
)

var (
	axes = [4][2]int{[2]int{-1, -1}, [2]int{-1, 0}, [2]int{-1, 1}, [2]int{0, -1}}
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

func SendToClient(request types.Request, client *types.Client) {
	data, err := util.GobToBytes(request)

	if err != nil {
			fmt.Println(err)
			return
	}

	select {
	case client.Data <- data:
	default:
			close(client.Data)
	}
}