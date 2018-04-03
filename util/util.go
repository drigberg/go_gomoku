package util

import (
	"fmt"
	"os"
	"net"
	"time"
	"bytes"
	"encoding/gob"
	"gomoku/types"
)

func CheckError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}

func AcceptConnection(conn net.Conn, handler func(net.Conn, types.Request)) {
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	defer conn.Close()

	data := HandleGob(conn)

	handler(conn, data)
}

func HandleGob(conn net.Conn) types.Request {
	fmt.Println("Receive GOB data!")

	var data types.Request

	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&data)

	if err != nil {
		fmt.Println("Error decoding GOB data:", err)
	}

	fmt.Printf("Data received: \n%#v\n", data)

	return data
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