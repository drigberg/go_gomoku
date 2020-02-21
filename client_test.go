package main

import (
	"go_gomoku/client"
	"go_gomoku/constants"
	"go_gomoku/helpers"
	"go_gomoku/types"
	"testing"
)

func TestClientHandleMessage(t *testing.T) {
	newClient := client.New("GoGomoku")
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
	if len(newClient.Messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.Messages), newClient.Messages)
	}

	newClient.Handler(message)

	if len(newClient.Messages) != 1 {
		t.Errorf("Expected 1 message, found %d: %s", len(newClient.Messages), newClient.Messages)
	}

	if newClient.Messages[0].Author != "Opponent" {
		t.Errorf("Expected message author to be 'Opponent', got: %s", newClient.Messages[0].Author)
	}

	if newClient.Messages[0].Content != content {
		t.Errorf("Expected message content to be %s, got: %s", content, newClient.Messages[0].Author)
	}
}

func TestClientHandleMessageMultiple(t *testing.T) {
	newClient := client.New("GoGomoku")
	newClient.DisablePrint = true
	messages := []string{
		"this is a message",
		"this is another message",
		"this is another another message",
	}

	if len(newClient.Messages) != 0 {
		t.Errorf("Expected no messages, found %d: %s", len(newClient.Messages), newClient.Messages)
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

	if len(newClient.Messages) != len(messages) {
		t.Errorf("Expected %d messages, found %d: %s", len(messages), len(newClient.Messages), newClient.Messages)
	}

	for i, message := range messages {
		if newClient.Messages[i].Author != "Opponent" {
			t.Errorf("Expected message author to be 'Opponent', got: %s", newClient.Messages[0].Author)
		}

		if newClient.Messages[i].Content != message {
			t.Errorf("Expected message content to be %s, got: %s", message, newClient.Messages[0].Author)
		}
	}
}
