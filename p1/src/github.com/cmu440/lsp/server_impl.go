// Contains the implementation of a LSP server.

package lsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cmu440/lspnet"
	"strconv"
)

type clientConnection struct {
	Addr   *lspnet.UDPAddr
	SeqNum int
}

type server struct {
	Connection        *lspnet.UDPConn
	clients           map[int]*clientConnection
	addrMap           map[string]int
	ConnectionNum     int
	readMessageQueue  chan *Message
	writeMessageQueue chan *Message
	newClient         chan *lspnet.UDPAddr
}

// NewServer creates, initiates, and returns a new server. This function should
// NOT block. Instead, it should spawn one or more goroutines (to handle things
// like accepting incoming client connections, triggering epoch events at
// fixed intervals, synchronizing events using a for-select loop like you saw in
// project 0, etc.) and immediately return. It should return a non-nil error if
// there was an error resolving or listening on the specified port number.
func NewServer(port int, params *Params) (Server, error) {
	udpAddr, _ := lspnet.ResolveUDPAddr("udp", lspnet.JoinHostPort("localhost", strconv.Itoa(port)))
	udpConn, _ := lspnet.ListenUDP("udp", udpAddr)
	s := server{
		udpConn,
		make(map[int]*clientConnection),
		make(map[string]int),
		1,
		make(chan *Message),
		make(chan *Message),
		make(chan *lspnet.UDPAddr),
	}
	fmt.Println("Server started", lspnet.JoinHostPort("localhost", strconv.Itoa(port)))
	go handleRoutine(&s)
	go handleAccept(&s)
	return &s, nil
}

func (s *server) Read() (int, []byte, error) {
	message := <-s.readMessageQueue
	return message.ConnID, message.Payload, nil
}

func (s *server) Write(connID int, payload []byte) error {
	client, exists := s.clients[connID]
	if exists {
		s.writeMessageQueue <- NewData(connID, client.SeqNum, len(payload), payload)
	}
	return nil
}

func (s *server) CloseConn(connID int) error {
	return errors.New("not yet implemented")
}

func (s *server) Close() error {
	return s.Connection.Close()
}

func handleAccept(s *server) {
	for {
		select {
		default:
			buffer := make([]byte, 1000)
			length, addr, _ := s.Connection.ReadFromUDP(buffer)
			var msg Message
			json.Unmarshal(buffer[:length], &msg)
			switch msgType := msg.Type; msgType {
			case MsgAck:
				client, existed := s.clients[msg.ConnID]
				if existed {
					client.SeqNum = msg.SeqNum + 1
				}
			case MsgData:
				ack, _ := json.Marshal(NewAck(msg.ConnID, msg.SeqNum))
				s.Connection.WriteToUDP(ack, addr)
				s.readMessageQueue <- &msg
			case MsgConnect:
				s.newClient <- addr
			}
		}
	}
}

func handleRoutine(s *server) {
	for {
		select {
		case addr := <-s.newClient:
			_, exists := s.addrMap[addr.String()]
			if !exists {
				ack, _ := json.Marshal(NewAck(s.ConnectionNum, 0))
				s.clients[s.ConnectionNum] = &clientConnection{
					addr,
					0,
				}
				s.addrMap[addr.String()] = s.ConnectionNum

				s.ConnectionNum++
				s.Connection.WriteToUDP(ack, addr)
			}
		case msg := <-s.writeMessageQueue:
			client := s.clients[msg.ConnID]
			byteMessage, _ := json.Marshal(msg)
			s.Connection.WriteToUDP(byteMessage, client.Addr)
		default:

		}
	}
}
