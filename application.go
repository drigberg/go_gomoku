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

	if *clientMode == true {
		client := client.CreateClient()
		client.Run(host, port)
	} else {
		server := server.CreateServer()
		server.Listen(port)
	}
}
