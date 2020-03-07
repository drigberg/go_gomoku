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

func waitForHandledRequest(client *Client, action string) (Request, error) {
	var err error
	var request Request
	done := make(chan Request)

	go func() {
		for {
			request := <- client.handledRequests
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
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	client := NewClient("Test")
	client.disablePrint = true
	socketClient := client.Connect("localhost", "3003")
	defer socketClient.Socket.Close()

	connected := make(chan bool)
	go socketClient.Receive(client.handler, &connected)
	select {
	case <-connected:
	case <-time.After(1 * time.Second):
		t.Error("Could not connect")
	}

	// verify sent home
	request, err := waitForHandledRequest(&client, HOME)
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
	client := NewClient("Test")
	client.disablePrint = true
	newSocketClient := client.Connect("localhost", "3003")
	connected := make(chan bool)
	reader := incrementalReader{make(chan string)}

	go newSocketClient.Receive(client.handler, &connected)
	select {
	case <-connected:
	case <-time.After(1 * time.Second):
		err = errors.New("Could not connect")
	}

	go func() {
		client.listenForInput(reader)
	}()

	player := PlayerBundle{
		&client,
		newSocketClient,
		&reader,
	}

	if err == nil {
		// verify sent home
		_, homeError := waitForHandledRequest(&client, HOME)
		if homeError != nil {
			err = homeError
		}
	}

	return player, err
}

func TestGoGomokuCreateGameSuccess(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

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
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

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
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

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
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

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
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

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

	if player1.client.yourTurn {
		return player1, player2, nil
	}
	return player2, player1, nil
}

func TestGoGomokuJoinGameSuccess(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}

	if player1.client.yourColor != "" {
		t.Errorf("Expected player1 color to be empty, got %s", player1.client.yourColor)
	}

	if player1.client.opponentColor != "" {
		t.Errorf("Expected player1 opponent color to be empty, got %s", player1.client.yourColor)
	}

	if player2.client.opponentColor != "" {
		t.Errorf("Expected player2 opponent color to be empty, got %s", player2.client.yourColor)
	}
}

func playMove(move string, movingPlayer PlayerBundle, waitingPlayer PlayerBundle) error {
	movingPlayer.reader.input <- move
	_, err := waitForHandledRequest(movingPlayer.client, MOVE)
	if err != nil {
		return err
	}
	_, err = waitForHandledRequest(waitingPlayer.client, MOVE)
	if err != nil {
		return err
	}

	if movingPlayer.client.yourTurn {
		return errors.New("Should not be moving player's turn anymore")
	}

	if !waitingPlayer.client.yourTurn {
		return errors.New("Should be waiting player's turn now")
	}
	return nil
}

func TestGoGomokuFirstMoveSuccess(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 1, 1 2, 1 3\n", player1, player2)
	if err != nil {
		t.Error(err)
	}
}

func TestGoGomokuSecondMovePass(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 1, 1 2, 1 3\n", player1, player2)
	if err != nil {
		t.Error(err)
	}

	err = playMove("mv pass\n", player2, player1)
	if err != nil {
		t.Error(err)
	}

	if player1.client.yourColor != "white" {
		t.Errorf("Expected player 1's color to be white, got %s", player1.client.yourColor)
	}
	if player1.client.opponentColor != "black" {
		t.Errorf("Expected player 1's opponentColor to be black, got %s", player1.client.opponentColor)
	}

	if player2.client.yourColor != "black" {
		t.Errorf("Expected player 2's color to be black, got %s", player2.client.yourColor)
	}
	if player2.client.opponentColor != "white" {
		t.Errorf("Expected player 2's opponentColor to be white, got %s", player2.client.opponentColor)
	}
}

func TestGoGomokuSecondMoveSuccess(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 1, 1 2, 1 3\n", player1, player2)
	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 4\n", player2, player1)
	if err != nil {
		t.Error(err)
	}

	if player1.client.yourColor != "black" {
		t.Errorf("Expected player 1's color to be black, got %s", player1.client.yourColor)
	}
	if player1.client.opponentColor != "white" {
		t.Errorf("Expected player 1's opponentColor to be white, got %s", player1.client.opponentColor)
	}

	if player2.client.yourColor != "white" {
		t.Errorf("Expected player 2's color to be white, got %s", player2.client.yourColor)
	}
	if player2.client.opponentColor != "black" {
		t.Errorf("Expected player 2's opponentColor to be black, got %s", player2.client.opponentColor)
	}
}

func TestGoGomokuFurtherMoveSuccessAfterPass(t *testing.T) {
	server := NewServer()
	go server.Listen("3003")
	defer server.Stop()

	player1, player2, err := setupGame(t)
	defer player1.socketClient.Socket.Close()
	defer player2.socketClient.Socket.Close()

	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 1, 1 2, 1 3\n", player1, player2)
	if err != nil {
		t.Error(err)
	}

	err = playMove("mv pass\n", player2, player1)
	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 4\n", player1, player2)
	if err != nil {
		t.Error(err)
	}

	err = playMove("mv 1 5\n", player2, player1)
	if err != nil {
		t.Error(err)
	}
}
