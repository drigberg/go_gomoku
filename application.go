package main

import (
	"flag"
	"go_gomoku/client"
	"go_gomoku/server"
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
		server.Run(port)
	} else {
		client.Run(host, port)
	}
}
