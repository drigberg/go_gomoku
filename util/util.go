package util

import (
	"fmt"
	"os"
	"net"
	"bytes"
	"encoding/gob"
	"gomoku/types"
	"bufio"
)

func CheckError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}

func AcceptConnection(conn net.Conn, handler func(*bufio.ReadWriter, types.Request)) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	for {
		response, err := rw.Read([]byte(""))
		fmt.Println("RESPONSE:", response)
		fmt.Println(err)

		data := HandleRwGob(rw)

		fmt.Println("DATA:", data)
		go handler(rw, data)
		fmt.Println("Handled!")
	}
}

func HandleRwGob(rw *bufio.ReadWriter) types.Request {
	fmt.Println("Receive GOB data!")

	var data types.Request

	decoder := gob.NewDecoder(rw)
	err := decoder.Decode(&data)

	if err != nil {
		fmt.Println("Error decoding GOB data:", err)
	}

	fmt.Printf("Data received: \n%#v\n", data)

	return data
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