package main

import (
	"strconv"
	"testing"
)

func TestClientHandleMessage(t *testing.T) {
	newClient := NewClient("GoGomoku")
	newClient.disablePrint = true
	content := "this is a message"
	request := Request{
		Success: true,
		GameID:  3,
		UserID:  "hello",
		Action:  MESSAGE,
		Data:    content,
	}
	message, err := gobToBytes(request)
	if err != nil {
		t.Errorf("Got error while encoding gob: %s", err)
	}
	if len(newClient.messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.messages), newClient.messages)
	}

	newClient.handler(message)

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
	newClient := NewClient("GoGomoku")
	newClient.disablePrint = true
	messages := []string{
		"this is a message",
		"this is another message",
		"this is another another message",
	}

	if len(newClient.messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.messages), newClient.messages)
	}

	for _, message := range messages {
		request := Request{
			Success: true,
			GameID:  3,
			UserID:  "hello",
			Action:  MESSAGE,
			Data:    message,
		}
		messageBytes, err := gobToBytes(request)
		if err != nil {
			t.Errorf("Got error while encoding gob: %s", err)
		}
		newClient.handler(messageBytes)
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
	newClient := NewClient(serverName)
	newClient.disablePrint = true
	gameID := 3
	request := Request{
		Success: true,
		GameID:  gameID,
		Action:  CREATE,
	}
	requestBytes, err := gobToBytes(request)
	if err != nil {
		t.Errorf("Got error while encoding gob: %s", err)
	}
	newClient.handler(requestBytes)
	if newClient.gameOver {
		t.Error("Expected gameOver to be false")
	}

	if newClient.GameID != gameID {
		t.Errorf("Expected gameID to be %d, got %d", gameID, newClient.GameID)
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

func TestClientHandleCreateError(t *testing.T) {
	serverName := "GoGomoku"
	newClient := NewClient(serverName)
	newClient.disablePrint = true
	request := Request{
		Success: false,
		Action:  CREATE,
	}
	requestBytes, err := gobToBytes(request)
	if err != nil {
		t.Errorf("Got error while encoding gob: %s", err)
	}
	newClient.handler(requestBytes)

	if len(newClient.messages) != 1 {
		t.Errorf("Expected 1 message, found %d: %s", len(newClient.messages), newClient.messages)
	}

	if newClient.messages[0].Author != serverName {
		t.Errorf("Expected message author to be %s, got: %s", serverName, newClient.messages[0].Author)
	}

	expectedMessageContent := "Error! Could not create game."

	if newClient.messages[0].Content != expectedMessageContent {
		t.Errorf("Expected message content to be %s, got: %s", expectedMessageContent, newClient.messages[0].Content)
	}
}
