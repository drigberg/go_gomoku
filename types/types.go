package types

type Coord struct {
    x int
    y int
}

type GameRoom struct {
    Id int
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
