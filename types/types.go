package types

import (
	"net"
	"strconv"
	"sync"
)

// SocketClient contains the connection to a client
type SocketClient struct {
	Socket net.Conn
	Data   chan []byte
	Closed bool
	M      sync.Mutex
}

// Receive listens for data and handles it
func (socketClient *SocketClient) Receive(handler func([]byte), connected *chan bool) {
	firstMessage := true
	for {
		message := make([]byte, 4096)
		length, err := socketClient.Socket.Read(message)
		if err != nil {
			socketClient.Socket.Close()
			break
		}
		if length > 0 {
			// switch
			if firstMessage {
				go func() { *connected <- true }()
				firstMessage = false
			}
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
	Data     string
	YourTurn bool
	Turn     int
	Colors   map[string]string
	Board    map[string]map[string]bool
	Home     []OpenRoom
}

type Player struct {
	UserID       string
	SocketClient *SocketClient
	Color        string
}
