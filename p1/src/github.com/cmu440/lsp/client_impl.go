// Contains the implementation of a LSP client.

package lsp

import (
	"encoding/json"
	"fmt"
	"github.com/cmu440/lspnet"
)

type client struct {
	ConnectionID int
	Connection   *lspnet.UDPConn
	SeqNum       int
}

// NewClient creates, initiates, and returns a new client. This function
// should return after a connection with the server has been established
// (i.e., the client has received an Ack message from the server in response
// to its connection request), and should return a non-nil error if a
// connection could not be made (i.e., if after K epochs, the client still
// hasn't received an Ack message from the server in response to its K
// connection requests).
//
// hostport is a colon-separated string identifying the server's host address
// and port number (i.e., "localhost:9999").
func NewClient(hostport string, params *Params) (Client, error) {
	udpAddr, _ := lspnet.ResolveUDPAddr("udp", hostport)
	udpConn, _ := lspnet.DialUDP("udp", nil, udpAddr)
	client := client{0, udpConn, 0}
	connByteMessage, _ := json.Marshal(NewConnect())
	client.Connection.Write(connByteMessage)
	ackMessage, _ := client.Read()

	var ack Message
	json.Unmarshal(ackMessage, &ack)
	client.ConnectionID = ack.ConnID
	return &client, nil
}

func (c *client) ConnID() int {
	return c.ConnectionID
}

func (c *client) Read() ([]byte, error) {
	for {
		buffer := make([]byte, 1000)
		length, _ := c.Connection.Read(buffer)
		var msg Message
		json.Unmarshal(buffer[:length], &msg)
		switch msgType := msg.Type; msgType {
		case MsgAck:
			if c.SeqNum == 0 {
				return buffer[:length], nil
			}
		case MsgData:
			ackByteMessage, _ := json.Marshal(NewAck(msg.ConnID, msg.SeqNum))
			c.Connection.Write(ackByteMessage)
			return msg.Payload, nil
		default:
			fmt.Println("Default")
		}
	}
}

func (c *client) Write(payload []byte) error {
	c.SeqNum++
	byteMessage, _ := json.Marshal(NewData(c.ConnID(), c.SeqNum, len(payload), payload))
	_, err := c.Connection.Write(byteMessage)
	return err
}

func (c *client) Close() error {
	return c.Connection.Close()
}
