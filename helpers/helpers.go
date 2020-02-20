package helpers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"go_gomoku/types"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func clearWindows() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func clearLinuxOrDarwin() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// ClearScreen clears the screen, based on operating system
func ClearScreen() {
	switch runtime.GOOS {
	case "linux":
	case "darwin":
		clearLinuxOrDarwin()
	case "windows":
		clearWindows()
	default:
		panic("Your platform is unsupported! Can't clear the terminal screen.")
	}
}

// DecodeGob builds a request from a gob
func DecodeGob(message []byte) types.Request {
	var network bytes.Buffer
	network.Write(message)
	var request types.Request

	dec := gob.NewDecoder(&network)

	err := dec.Decode(&request)

	if err != nil {
		fmt.Println("Error decoding GOB data:", err)
	}

	return request
}

// GobToBytes converts a gob to bytes
func GobToBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SendBackoff tries to send and keeps trying
func SendBackoff(data []byte, socketClient *types.SocketClient, i int) {
	if i > 1 {
		fmt.Println("Retrying message send! Attempt:", i)
	}
	if socketClient.Closed {
		return
	}

	select {
	case socketClient.Data <- data:
		return
	default:
		time.Sleep(500 * time.Millisecond)

		if i > 5 {
			return
		}
		SendBackoff(data, socketClient, i+1)
	}
}

// SendToClient tries to send a request to socketClient, with backoff
func SendToClient(request types.Request, socketClient *types.SocketClient) {
	data, err := GobToBytes(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	SendBackoff(data, socketClient, 1)
}
