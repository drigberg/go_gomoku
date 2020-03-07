package main

import (
	"flag"
	"os"
)

func parseEnv() (string, string, bool) {
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
	return port, host, *clientMode
}

func main() {
	port, host, clientMode := parseEnv()

	if clientMode == true {
		client := NewClient("GoGomoku")
		client.Run(host, port)
	} else {
		server := NewServer()
		server.Listen(port)
	}
}
