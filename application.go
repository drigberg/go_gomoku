package main

import (
	"flag"
	"go_gomoku/client"
	"go_gomoku/server"
	"go_gomoku/types"
	"os"
)

func main() {
	clientMode := flag.Bool("play", false, "activate client mode")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	flag.Parse()

	if *clientMode == false {
		server := server.Server{
			Games:  make(map[int]*types.GameRoom),
			GameID: 0,
		}
		server.Listen(port)
	} else {
		client.Run(host, port)
	}
}
