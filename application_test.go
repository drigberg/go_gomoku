package main

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"go_gomoku/client"
	"go_gomoku/constants"
	"go_gomoku/server"
	"go_gomoku/types"
)

type incrementalReader struct {
	input 	chan string
}

func (reader incrementalReader) Read(p []byte) (n int, err error) {
	str := <- reader.input
	buf := new(bytes.Buffer)
	buf.Write([]byte(str))
	n, err = buf.Read(p)
	return
}

func waitForHandledRequest(newClient *client.Client, action string) (types.Request, error) {
	var err error
	var request types.Request
	done := make(chan types.Request)

	go func() {
		for {
			request := <- newClient.HandledRequests
			if request.Action == action {
				done <- request
				return
			}
		}
	}()

	select {
	case r := <-done:
		request = r
	case <-time.After(1 * time.Second):
		err = errors.New("Unable to connect and create game")
	}
	return request, err
}

func TestGoGomokuConnectSuccess(t *testing.T) {
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
	case <-time.After(1 * time.Second):
		t.Error("Could not connect")
	}

	request, err := waitForHandledRequest(&newClient, constants.HOME)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Home) != 0 {
		t.Errorf("Expected there to be no active games, found %d", len(request.Home))
	}
}

func TestGoGomokuCreateGameSuccess(t *testing.T) {
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
	case <-time.After(1 * time.Second):
		t.Error("Could not connect")
	}
	waitForHandledRequest(&newClient, constants.HOME)

	reader := incrementalReader{make(chan string)}
	go func() {newClient.ListenForInput(reader)}()

	reader.input <- "mk\n"
	waitForHandledRequest(&newClient, constants.CREATE)
	
	if newClient.GameID == -1 {
		t.Errorf("Client still has gameID -1")
	}
}

func TestGoGomokuHomeFromGameSuccess(t *testing.T) {
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
	case <-time.After(1 * time.Second):
		t.Error("Could not connect")
	}

	waitForHandledRequest(&newClient, constants.HOME)

	reader := incrementalReader{make(chan string)}
	go func() {newClient.ListenForInput(reader)}()

	reader.input <- "mk\n"
	waitForHandledRequest(&newClient, constants.CREATE)

	reader.input <- "hm\n"
	reader.input <- "y\n"
	request, err := waitForHandledRequest(&newClient, constants.HOME)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Home) != 1 {
		t.Errorf("Expected there to be 1 active game, found %d", len(request.Home))
	}	
}