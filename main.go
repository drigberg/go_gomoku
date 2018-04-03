package main

import (
    "flag"
    "go_gomoku/server"
    "go_gomoku/client"
    "fmt"
    "os"
)

var (
    serverPort = ":9000"
)

func main() {
    connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")
    clientPort := flag.String("port", "", "Port to use for client")

    flag.Parse()

    if len(*connect) == 0 && len(*clientPort) == 0 {
        fmt.Println("Please provide a port number!")
        os.Exit(1)
    }

    if *connect == "server" {
        server.Run(serverPort)
    } else {
        client.Run(*clientPort, serverPort)
    }
}

