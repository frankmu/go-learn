// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
)

type request struct {
	isGet bool
	key   string
	value []byte
}
type client struct {
	connection net.Conn
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
	init_db()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return nil
		}
		c := &client{conn}
		go handleRequest(c)
	}
	return nil
}

func (kvs *keyValueServer) Close() {
	kvs.listener.Close()
}

func (kvs *keyValueServer) Count() int {

	return -1
}

func handleRequest(c *client) {
	reader := bufio.NewReader(c.connection)
	message, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	} else {
		fmt.Println(string(message))
		tokens := bytes.Split(message, []byte(","))
		if string(tokens[0]) == "put" {
			key := string(tokens[1])
			value := string(tokens[2])
			fmt.Println("put- " + key + ": " + value)
			put(key, tokens[2])
		} else {
			k := tokens[1][:len(tokens[1])-1]
			key := string(k)
			fmt.Println("get- " + key + ": " + string(get(key)))
			c.connection.Write(get(key))
		}
	}
	// Close the connection when you're done with it.
	c.connection.Close()
}
