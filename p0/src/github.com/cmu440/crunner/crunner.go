package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
)

// To test your server implementation, you might find it helpful to implement a
// simple 'client runner' program. The program could be very simple, as long as
// it is able to connect with and send messages to your server and is able to
// read and print out the server's response to standard output. Whether or
// not you add any code to this file will not affect your grade.
func main() {
	// connect to this socket
	conn, err := net.Dial("tcp", defaultHost+":"+strconv.Itoa(defaultPort))
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()
	fmt.Println("Connecting to " + defaultHost + ":" + strconv.Itoa(defaultPort))
	var wg sync.WaitGroup
	wg.Add(2)
	go handleWrite(conn, &wg)
	go handleRead(conn, &wg)
	wg.Wait()
}
func handleWrite(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i > 0; i-- {
		_, e := conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))
		if e != nil {
			fmt.Println("Error to send message because of ", e.Error())
			break
		}
	}
}
func handleRead(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	reader := bufio.NewReader(conn)
	for i := 1; i <= 10; i++ {
		line, _ := reader.ReadString(byte('\n'))

		fmt.Print(line)
	}
}
