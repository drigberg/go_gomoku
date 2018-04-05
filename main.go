package main

import (
    "flag"
    "go_gomoku/server"
    "go_gomoku/client"
    "go_gomoku/util"
    "os/exec"
    "os"
)

var (
    serverPort = ":9000"
)

func init() {
	// define 'clear' command for each operating system
	util.Clear = make(map[string]func())
	util.Clear["linux"] = func() {
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
	}
	util.Clear["windows"] = func() {
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()
	}

	util.Clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
    connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")

    flag.Parse()

    if *connect == "server" {
        server.Run(serverPort)
    } else {
        client.Run(serverPort)
    }
}

