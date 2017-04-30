// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"bytes"
	//"fmt"
	"io"
	"net"
	"strconv"
)

type request struct {
	isGet bool
	key   string
	value []byte
}
type client struct {
	connection       net.Conn
	id               int
	messageQueue     chan []byte
	quitSignal_Write chan int
	quitSignal_Read  chan int
}
type keyValueServer struct {
	listener      net.Listener
	clients       []*client
	req           chan *request
	res           chan *request
	newResponse   chan []byte
	newConnection chan net.Conn
	countClients  chan int
	clientCount   chan int
	deadClient    chan *client
	quit_accept   chan int
	quit_main     chan int
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	return &keyValueServer{
		nil,
		make([]*client, 0),
		make(chan *request),
		make(chan *request),
		make(chan []byte),
		make(chan net.Conn),
		make(chan int),
		make(chan int),
		make(chan *client),
		make(chan int),
		make(chan int)}
}

func (kvs *keyValueServer) Start(port int) error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	//fmt.Println("Listening on " + strconv.Itoa(port))
	kvs.listener = ln
	init_db()

	go handleRoutine(kvs)
	go handleAccept(kvs)
	return nil
}

func (kvs *keyValueServer) Close() {
	kvs.listener.Close()
	kvs.quit_accept <- 0
	kvs.quit_main <- 0
}

func (kvs *keyValueServer) Count() int {
	//fmt.Println("Get Count")
	kvs.countClients <- 0
	count := <-kvs.clientCount
	//fmt.Println(count)
	return count
}

func handleRoutine(kvs *keyValueServer) {
	counter := 0
	for {
		select {
		case <-kvs.quit_main:
			for _, c := range kvs.clients {
				c.connection.Close()
				c.quitSignal_Write <- 0
				c.quitSignal_Read <- 0
			}
			//fmt.Println("Quit Main")
			return
		case <-kvs.countClients:
			kvs.clientCount <- len(kvs.clients)
		case req := <-kvs.req:
			if req.isGet {
				v := get(req.key)
				kvs.res <- &request{
					value: v,
				}
			} else {
				put(req.key, req.value)
			}
		case newConnection := <-kvs.newConnection:
			c := &client{
				newConnection,
				counter,
				make(chan []byte),
				make(chan int),
				make(chan int)}
			kvs.clients = append(kvs.clients, c)
			counter++
			//fmt.Println("new connection")
			//fmt.Println(len(kvs.clients))
			go handleRequest(kvs, c)
			go sendResponse(c)
		case newResponse := <-kvs.newResponse:
			for _, c := range kvs.clients {
				c.messageQueue <- newResponse
			}
		case deadClient := <-kvs.deadClient:
			for i, c := range kvs.clients {
				if c == deadClient {
					kvs.clients = append(kvs.clients[:i], kvs.clients[i+1:]...)
				}
			}
		}

	}
}
func handleAccept(kvs *keyValueServer) {
	for {
		select {
		case <-kvs.quit_accept:
			//fmt.Println("Quit Accept")
			return
		default:
			conn, err := kvs.listener.Accept()
			if err == nil {
				kvs.newConnection <- conn
			}
		}
	}
}

func handleRequest(kvs *keyValueServer, c *client) {
	reader := bufio.NewReader(c.connection)
	for {
		select {
		case <-c.quitSignal_Read:
			//fmt.Println("Quit Main")
			return
		default:
			message, err := reader.ReadBytes('\n')
			if err == io.EOF {
				kvs.deadClient <- c
			} else if err == nil {
				tokens := bytes.Split(message, []byte(","))
				//fmt.Println(strconv.Itoa(c.id) + " - handle request: " + string(message))
				if string(tokens[0]) == "put" {
					key := string(tokens[1])
					//value := string(tokens[2])
					//fmt.Println("put- " + key + ": " + value)
					kvs.req <- &request{
						isGet: false,
						key:   key,
						value: tokens[2]}
					//put(key, tokens[2])
				} else {
					k := tokens[1][:len(tokens[1])-1]
					key := string(k)
					//fmt.Println("get- " + key + ": " + string(get(key)))
					kvs.req <- &request{
						isGet: true,
						key:   key}
					response := <-kvs.res
					kvs.newResponse <- append(append(k, ","...), response.value...)
				}
			}
		}
	}

	// Close the connection when you're done with it.
	// c.connection.Close()
}
func sendResponse(c *client) {
	for {
		select {
		case <-c.quitSignal_Write:
			//fmt.Println("Quit Main")
			return
		case message := <-c.messageQueue:
			c.connection.Write(message)
		}
	}
}
