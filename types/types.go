package types

import (
    "net"
    "fmt"
)

type Client struct {
    Socket net.Conn
    Data chan []byte
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
    Players [2]Player
    Turn int
}

type Request struct {
    GameId int
    UserId string
    Action string
    Success bool
    Data string
    Turn int
    Board map[string]map[Coord]bool
}

type Player struct {
    UserId string
    Spots map[Coord]bool
    Client *Client
    Color string
}
