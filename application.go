package main

import (
    "flag"
    "go_gomoku/server"
    "go_gomoku/client"
    "go_gomoku/util"
    "os/exec"
    "os"
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
    clientMode := flag.Bool("play", false, "activate client mode")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	host := flag.String("host", "GoGomoku-env.ipxp2bqu4z.eu-west-1.elasticbeanstalk.com", "server tcp host")

    flag.Parse()

    if *clientMode == false {
        server.Run(port)
    } else {
        client.Run(host, port)
    }
}

