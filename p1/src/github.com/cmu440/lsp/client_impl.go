// Contains the implementation of a LSP client.

package lsp

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/cmu440/lspnet"
)

type client struct {
	ConnectionID        int
	Connection          *lspnet.UDPConn
	SeqNum              int
	WindowSize          int
	SendSeqNum          int
	SendWindow          []bool
	SendBuffer          *list.List
	WriteMessageChannel chan *Message
	AckMessageChannel   chan *Message
	ReadMessageChannel  chan *Message
}

func NewClient(hostport string, params *Params) (Client, error) {
	udpAddr, _ := lspnet.ResolveUDPAddr("udp", hostport)
	udpConn, _ := lspnet.DialUDP("udp", nil, udpAddr)
	client := client{
		0,
		udpConn,
		0,
		params.WindowSize,
		0,
		make([]bool, params.WindowSize),
		list.New(),
		make(chan *Message),
		make(chan *Message),
		make(chan *Message),
	}

	go client.mainRoutine()
	go client.readRoutine()
	connByteMessage, _ := json.Marshal(NewConnect())
	client.Connection.Write(connByteMessage)
	client.Read()

	return &client, nil
}

func (c *client) ConnID() int {
	return c.ConnectionID
}

func (c *client) Read() ([]byte, error) {
	msg := <-c.ReadMessageChannel
	return msg.Payload, nil
}

func (c *client) Write(payload []byte) error {
	c.WriteMessageChannel <- NewData(c.ConnID(), c.SeqNum, len(payload), payload)
	c.SeqNum++
	return nil
}

func (c *client) Close() error {
	return c.Connection.Close()
}

func (c *client) mainRoutine() {
	for {
		select {
		case writeMessage := <-c.WriteMessageChannel:
			if writeMessage.Type == MsgAck {
				byteMessage, _ := json.Marshal(writeMessage)
				c.Connection.Write(byteMessage)
			} else {
				if writeMessage.SeqNum > c.SendSeqNum && writeMessage.SeqNum <= c.SendSeqNum+c.WindowSize {
					byteMessage, _ := json.Marshal(writeMessage)
					c.Connection.Write(byteMessage)
					c.SendWindow[writeMessage.SeqNum-c.SendSeqNum-1] = false
				} else {
					c.SendBuffer.PushBack(writeMessage)
				}
			}
		case ackMessage := <-c.AckMessageChannel:
			if ackMessage.SeqNum == 0 {
				c.ConnectionID = ackMessage.ConnID
				c.SeqNum++
			} else if ackMessage.SeqNum > c.SendSeqNum {
				c.SendWindow[ackMessage.SeqNum%c.WindowSize] = true
				for index := c.SendSeqNum % c.WindowSize; index < c.WindowSize; index++ {
					if c.SendWindow[index] == true {
						if e := c.SendBuffer.Front(); e != nil {
							byteMessage, _ := json.Marshal(e.Value.(*Message))
							c.Connection.Write(byteMessage)
							c.SendWindow[index] = false
							c.SendBuffer.Remove(e)
						}
						c.SendSeqNum++
					} else {
						break
					}
				}

			}
		}
	}
}

func (c *client) readRoutine() {
	for {
		select {
		default:
			buffer := make([]byte, 1000)
			length, _ := c.Connection.Read(buffer)
			var msg Message
			json.Unmarshal(buffer[:length], &msg)
			switch msgType := msg.Type; msgType {
			case MsgAck:
				c.AckMessageChannel <- &msg
				if msg.SeqNum == 0 {
					c.ReadMessageChannel <- &msg
				}
			case MsgData:
				c.WriteMessageChannel <- NewAck(msg.ConnID, msg.SeqNum)
				c.ReadMessageChannel <- &msg
			default:
				fmt.Println("Default")
			}
		}
	}
}
