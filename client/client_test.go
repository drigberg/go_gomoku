package client

import (
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"strconv"
	"testing"
)

func TestClientHandleMessage(t *testing.T) {
	newClient := New("GoGomoku")
	newClient.DisablePrint = true
	content := "this is a message"
	request := types.Request{
		Success: true,
		GameID:  3,
		UserID:  "hello",
		Action:  constants.MESSAGE,
		Data:    content,
	}
	message, err := helpers.GobToBytes(request)
	if err != nil {
		t.Errorf("Got error while encoding gob: %s", err)
	}
	if len(newClient.messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.messages), newClient.messages)
	}

	newClient.Handler(message)

	if len(newClient.messages) != 1 {
		t.Errorf("Expected 1 message, found %d: %s", len(newClient.messages), newClient.messages)
	}

	if newClient.messages[0].Author != "Opponent" {
		t.Errorf("Expected message author to be 'Opponent', got: %s", newClient.messages[0].Author)
	}

	if newClient.messages[0].Content != content {
		t.Errorf("Expected message content to be %s, got: %s", content, newClient.messages[0].Author)
	}
}

func TestClientHandleMessageMultiple(t *testing.T) {
	newClient := New("GoGomoku")
	newClient.DisablePrint = true
	messages := []string{
		"this is a message",
		"this is another message",
		"this is another another message",
	}

	if len(newClient.messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.messages), newClient.messages)
	}

	for _, message := range messages {
		request := types.Request{
			Success: true,
			GameID:  3,
			UserID:  "hello",
			Action:  constants.MESSAGE,
			Data:    message,
		}
		messageBytes, err := helpers.GobToBytes(request)
		if err != nil {
			t.Errorf("Got error while encoding gob: %s", err)
		}
		newClient.Handler(messageBytes)
	}

	if len(newClient.messages) != len(messages) {
		t.Errorf("Expected %d messages, found %d: %s", len(messages), len(newClient.messages), newClient.messages)
	}

	for i, message := range messages {
		if newClient.messages[i].Author != "Opponent" {
			t.Errorf("Expected message author to be 'Opponent', got: %s", newClient.messages[0].Author)
		}

		if newClient.messages[i].Content != message {
			t.Errorf("Expected message content to be %s, got: %s", message, newClient.messages[0].Author)
		}
	}
}

func TestClientHandleCreateSuccess(t *testing.T) {
	serverName := "GoGomoku"
	newClient := New(serverName)
	newClient.DisablePrint = true
	gameID := 3
	request := types.Request{
		Success: true,
		GameID:  gameID,
		Action:  constants.CREATE,
	}
	requestBytes, err := helpers.GobToBytes(request)
	if err != nil {
		t.Errorf("Got error while encoding gob: %s", err)
	}
	newClient.Handler(requestBytes)
	if newClient.gameOver {
		t.Error("Expected gameOver to be false")
	}

	if newClient.gameID != gameID {
		t.Errorf("Expected gameID to be %d, got %d", gameID, newClient.gameID)
	}

	if !newClient.yourTurn {
		t.Error("Expected gameOver to be true")
	}

	if len(newClient.messages) != 1 {
		t.Errorf("Expected 1 message, found %d: %s", len(newClient.messages), newClient.messages)
	}

	if newClient.messages[0].Author != serverName {
		t.Errorf("Expected message author to be %s, got: %s", serverName, newClient.messages[0].Author)
	}

	expectedMessageContent := "Created game #" + strconv.Itoa(gameID)

	if newClient.messages[0].Content != expectedMessageContent {
		t.Errorf("Expected message content to be %s, got: %s", expectedMessageContent, newClient.messages[0].Content)
	}
}
