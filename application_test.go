package main

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"
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

func waitForHandledRequest(newClient *Client, action string) (Request, error) {
	var err error
	var request Request
	done := make(chan Request)

	go func() {
		for {
			request := <- newClient.handledRequests
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
		err = errors.New("Timeout error: request not received")
	}
	return request, err
}

func TestGoGomokuConnectSuccess(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	newClient := NewClient("Test")
	newClient.disablePrint = true
	socketClient := newClient.Connect("localhost", "3003")
	defer socketClient.Socket.Close()

	connected := make(chan bool)
	go socketClient.Receive(newClient.handler, &connected)
	select {
	case <-connected:
	case <-time.After(1 * time.Second):
		t.Error("Could not connect")
	}

	// verify sent home
	request, err := waitForHandledRequest(&newClient, HOME)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Home) != 0 {
		t.Errorf("Expected there to be no active games, found %d", len(request.Home))
	}
}


type PlayerBundle struct {
	client 			*Client
	socketClient 	*SocketClient
	reader 			*incrementalReader
}

func setupClient(t *testing.T) (PlayerBundle, error) {
	var err error
	newClient := NewClient("Test")
	newClient.disablePrint = true
	newSocketClient := newClient.Connect("localhost", "3003")
	connected := make(chan bool)
	reader := incrementalReader{make(chan string)}

	go newSocketClient.Receive(newClient.handler, &connected)
	select {
	case <-connected:
	case <-time.After(1 * time.Second):
		err = errors.New("Could not connect")
	}

	go func() {
		newClient.listenForInput(reader)
	}()

	player := PlayerBundle{
		&newClient,
		newSocketClient,
		&reader,
	}

	if err == nil {
		// verify sent home
		_, homeError := waitForHandledRequest(&newClient, HOME)
		if homeError != nil {
			err = homeError
		}
	}

	return player, err
}

func TestGoGomokuCreateGameSuccess(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1Bundle, err := setupClient(t)
	defer player1Bundle.socketClient.Socket.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create game
	player1Bundle.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1Bundle.client, CREATE)
	if err != nil {
		t.Fatal(err)
	}
	
	if player1Bundle.client.GameID == -1 {
		t.Errorf("Client still has gameID -1")
	}
}

func TestGoGomokuHomeFromGameSuccess(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1Bundle, err := setupClient(t)
	defer player1Bundle.socketClient.Socket.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create game
	player1Bundle.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1Bundle.client, CREATE)
	if err != nil {
		t.Fatal(err)
	}

	player1Bundle.reader.input <- "hm\n"
	player1Bundle.reader.input <- "y\n"
	request, err := waitForHandledRequest(player1Bundle.client, HOME)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Home) != 1 {
		t.Errorf("Expected there to be 1 active game, found %d", len(request.Home))
	}	
}


func TestGoGomokuHomeFromGameUnconfirmed(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1Bundle, err := setupClient(t)
	defer player1Bundle.socketClient.Socket.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create game
	player1Bundle.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1Bundle.client, CREATE)
	if err != nil {
		t.Fatal(err)
	}

	// go home but do not confirm
	player1Bundle.reader.input <- "hm\n"
	_, err = waitForHandledRequest(player1Bundle.client, HOME)
	if err == nil {
		t.Error("Expected to not be sent home")
	}
}

func TestGoGomokuHomeFromGameRefused(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1Bundle, err := setupClient(t)
	defer player1Bundle.socketClient.Socket.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create game
	player1Bundle.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1Bundle.client, CREATE)
	if err != nil {
		t.Fatal(err)
	}

	// go home but type n for confirmation
	player1Bundle.reader.input <- "hm\n"
	player1Bundle.reader.input <- "n\n"
	_, err = waitForHandledRequest(player1Bundle.client, HOME)
	if err == nil {
		t.Error("Expected to not be sent home")
	}
}

func TestGoGomokuHomeRefresh(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1Bundle, err := setupClient(t)
	defer player1Bundle.socketClient.Socket.Close()
	if err != nil {
		t.Fatal(err)
	}

	// create game
	player1Bundle.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1Bundle.client, CREATE)
	if err != nil {
		t.Fatal(err)
	}

	// return home
	player1Bundle.reader.input <- "hm\n"
	player1Bundle.reader.input <- "y\n"
	request, err := waitForHandledRequest(player1Bundle.client, HOME)
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Home) != 1 {
		t.Errorf("Expected there to be 1 active game, found %d", len(request.Home))
	}

	// refresh home
	player1Bundle.reader.input <- "hm\n"
	request, err = waitForHandledRequest(player1Bundle.client, HOME)
	if err != nil {
		t.Error(err)
	}
	if request.Action != HOME {
		t.Errorf("Expected response action to be HOME, got %s", request.Action)
	}
	if len(request.Home) != 1 {
		t.Errorf("Expected there to be 1 active game, found %d", len(request.Home))
	}

	// refresh home
	player1Bundle.reader.input <- "hm\n"
	request, err = waitForHandledRequest(player1Bundle.client, HOME)
	if err != nil {
		t.Error(err)
	}
	if request.Action != HOME {
		t.Errorf("Expected response action to be HOME, got %s", request.Action)
	}
	if len(request.Home) != 1 {
		t.Errorf("Expected there to be 1 active game, found %d", len(request.Home))
	}
}

func setupGame(t *testing.T) (PlayerBundle, PlayerBundle, error) {
	// create player 1
	player1, err := setupClient(t)
	if err != nil {
		return player1, PlayerBundle{}, err
	}

	// create player 2
	player2, err := setupClient(t)
	if err != nil {
		return player1, player2, err
	}
	
	// create game
	player1.reader.input <- "mk\n"
	_, err = waitForHandledRequest(player1.client, CREATE)
	if err != nil {
		return player1, player2, err
	}
	
	if player1.client.GameID == -1 {
		return player1, player2, errors.New("Client still has gameID -1")
	}


	// join game
	player2.reader.input <- "jn " + strconv.Itoa(player1.client.GameID) + "\n"
	_, err = waitForHandledRequest(player2.client, JOIN)
	if err != nil {
		return player1, player2, err
	}

	_, err = waitForHandledRequest(player1.client, OTHERJOINED)
	if err != nil {
		return player1, player2, err
	}

	if player2.client.GameID != player1.client.GameID {
		return player1, player2, errors.New("PlayerBundles are not in same game")
	}

	return player1, player2, nil
}

func TestGoGomokuJoinGameSuccess(t *testing.T) {
	newServer := NewServer()
	go newServer.Listen("3003")
	defer newServer.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}
}
