package helpers

import (
	"fmt"
	"go_gomoku/types"
	"go_gomoku/util"
	"go_gomoku/constants"
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