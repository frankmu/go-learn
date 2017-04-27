// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
	"fmt"
	"net"
	"strconv"
)

type request struct {
	isGet bool
	key   string
	value []byte
}
type keyValueServer struct {
	listener net.Listener
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	return &keyValueServer{nil}
}

func (kvs *keyValueServer) Start(port int) error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	fmt.Println("Listening on " + strconv.Itoa(port))
	kvs.listener = ln
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return nil
		}
		go handleRequest(conn)
	}
	return nil
}

func (kvs *keyValueServer) Close() {
	kvs.listener.Close()
}

func (kvs *keyValueServer) Count() int {

	return -1
}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println(buf)
	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()
}
