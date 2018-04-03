package types

import (
    "bufio"
)

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
