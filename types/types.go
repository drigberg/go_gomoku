package types

import (
    "net"
    "fmt"
    "strconv"
)

type Client struct {
    Socket net.Conn
    Data chan []byte
    Closed bool
}


func (client *Client) Receive(handler func([]byte)) {
	for {
        message := make([]byte, 4096)
        length, err := client.Socket.Read(message)
        if err != nil {
                client.Socket.Close()
                break
        }
        if length > 0 {
            // switch
            handler(message)
        }
	}
}

type Coord struct {
    X int
    Y int
}

func (coord Coord) String() string {
    return strconv.Itoa(coord.X) + " " + strconv.Itoa(coord.Y)
}

type Message struct {
    Content string
    Author string
}

func (message Message) Print() {
    output := message.Author + ": " + message.Content
    fmt.Println(output)
}

type GameRoom struct {
    Id int
    Players map[string]*Player
    Turn int
    Board map[string]map[string]bool
    FirstPlayerId string
}

func (game GameRoom) PlayMove(move Coord, color string) {
    moveStr := move.String()

    game.Board[color][moveStr] = true
}

type Request struct {
    GameId int
    UserId string
    Action string
    Success bool
    Colors map[string]string
    Data string
    YourTurn bool
    Turn int
    Board map[string]map[string]bool
}

type Player struct {
    UserId string
    Client *Client
    Color string
}