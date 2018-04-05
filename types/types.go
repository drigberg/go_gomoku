package types

import (
    "net"
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
    x int
    y int
}

type GameRoom struct {
    Id int
    Players [2]Player
    Messages [6]string
    Turn int
}

type Request struct {
    GameId int
    UserId string
    Action string
    Success bool
    Data string
}

type Player struct {
    UserId string
    Spots map[int][]Coord
    Client *Client
    Color string
}
