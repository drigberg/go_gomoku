package util

import (
	"fmt"
	"os"
	"bytes"
	"encoding/gob"
	"go_gomoku/types"
	"runtime"
)

var (
	Clear map[string]func()
)

func CallClear() {
	function, ok := Clear[runtime.GOOS]
	if ok {
		fmt.Print(ok)
			function()
	} else {
			panic("Your platform is unsupported! Can't clear the terminal screen.")
	}
}

func CheckError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}

func DecodeGob(message []byte) types.Request{
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

func GobToBytes(key interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(key)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}