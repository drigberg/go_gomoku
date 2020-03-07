package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

func clearScreen() {
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

func decodeGob(message []byte) Request {
	var network bytes.Buffer
	network.Write(message)
	var request Request

	dec := gob.NewDecoder(&network)

	err := dec.Decode(&request)

	if err != nil {
		fmt.Println("Error decoding GOB data:", err)
	}

	return request
}

func gobToBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
