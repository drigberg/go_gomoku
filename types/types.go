package types

import (
	"fmt"
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

func (message Message) Print() {
	output := message.Author + ": " + message.Content
	fmt.Println(output)
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
	UserID       string
	SocketClient *SocketClient
	Color        string
}
