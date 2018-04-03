package main

import (
    "flag"
    "go_gomoku/server"
    "go_gomoku/client"
)

var (
    serverPort = ":9000"
)

func main() {
    connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")

    flag.Parse()

    if *connect == "server" {
        server.Run(serverPort)
    } else {
        client.Run(serverPort)
    }
}

