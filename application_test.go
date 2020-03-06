package main

import (
	"strings"
	"testing"
	"time"

	"go_gomoku/client"
	"go_gomoku/constants"
	"go_gomoku/server"
)

func TestGoGomokuCreateGame(t *testing.T) {
	newServer := server.New()
	go newServer.Listen("3003")
	defer newServer.Stop()

	newClient := client.New("Test")
	newClient.DisablePrint = true
	socketClient := newClient.Connect("localhost", "3003")
	defer socketClient.Socket.Close()

	connected := make(chan bool)
	go socketClient.Receive(newClient.Handler, &connected)
	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		t.Error("Could not connect")
	}

	reader := strings.NewReader("mk")
	newClient.ListenForInput(reader)
	done := make(chan bool)
	go func() {
		for {
			request := <- newClient.HandledRequests
			if request.Action == constants.CREATE {
				done <- true
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("Unable to connect and create game")
	}
	
	if newClient.GameID == -1 {
		t.Errorf("Client still has gameID -1")
	}
}


func TestGoGomokuCreateGame2(t *testing.T) {
	newServer := server.New()
	go newServer.Listen("3003")
	defer newServer.Stop()

	newClient := client.New("Test")
	newClient.DisablePrint = true
	socketClient := newClient.Connect("localhost", "3003")
	defer socketClient.Socket.Close()

	connected := make(chan bool)
	go socketClient.Receive(newClient.Handler, &connected)
	select {
	case <-connected:
	case <-time.After(5 * time.Second):
		t.Error("Could not connect")
	}

	reader := strings.NewReader("mk")
	newClient.ListenForInput(reader)
	done := make(chan bool)
	go func() {
		for {
			request := <- newClient.HandledRequests
			if request.Action == constants.CREATE {
				done <- true
				return
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("Unable to connect and create game")
	}
	
	if newClient.GameID == -1 {
		t.Errorf("Client still has gameID -1")
	}
}