package types

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

// Client contains the connection to a client
type Client struct {
	Socket net.Conn
	Data   chan []byte
	Closed bool
	M      sync.Mutex
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
	Author  string
}

func (message Message) Print() {
	output := message.Author + ": " + message.Content
	fmt.Println(output)
}

type GameRoom struct {
	ID            int
	Players       map[string]*Player
	Turn          int
	Board         map[string]map[string]bool
	FirstPlayerId string
	IsOver        bool
}

func (game GameRoom) PlayMove(move Coord, color string) {
	moveStr := move.String()

	game.Board[color][moveStr] = true
}

type OpenRoom struct {
	ID     int
	UserID string
}

type Request struct {
	GameID   int
	UserID   string
	Action   string
	Success  bool
	GameOver bool
	Colors   map[string]string
	Data     string
	YourTurn bool
	Turn     int
	Board    map[string]map[string]bool
	Home     []OpenRoom
}

type Player struct {
	UserID string
	Client *Client
	Color  string
}
