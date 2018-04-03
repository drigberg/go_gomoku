package types

import (
    "bufio"
    "net"
    "fmt"
)

type Client struct {
    Socket net.Conn
    Data chan []byte
}


func (client *Client) Receive() {
	for {
        message := make([]byte, 4096)
        length, err := client.Socket.Read(message)
        if err != nil {
                client.Socket.Close()
                break
        }
        if length > 0 {
            // switch
                fmt.Println("RECEIVED: " + string(message))
        }
	}
}

type Coord struct {
    x int
    y int
}

type GameRoom struct {
    Id int
    Channels [2]*bufio.ReadWriter
    Spots map[int][]Coord
    Players [2]int
    Messages [6]string
    Turn int
}

type Request struct {
    GameId int
    UserId int
    Action string
    Data string
}
