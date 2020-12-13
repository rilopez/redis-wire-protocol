package device

import (
	"bufio"
	"github.com/rilopez/redis-wire-protocol/internal/resp"
	"log"
	"net"
	"net/textproto"
	"sync"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

// Client is used to handle a client connection
type Client struct {
	ID         uint64
	conn       net.Conn
	toServer   chan<- common.Command
	fromServer chan common.Command
	now        func() time.Time
}

// NewClient allocates a Client
func NewClient(conn net.Conn, ID uint64, outbound chan<- common.Command, inbound chan common.Command, now func() time.Time) (*Client, error) {
	client := &Client{
		ID:         ID,
		conn:       conn,
		toServer:   outbound,
		fromServer: inbound,
		now:        now,
	}
	return client, nil
}

func (c *Client) receiveCommandsLoop() {
	reader := bufio.NewReader(c.conn)
	tp := textproto.NewReader(reader)
	log.Print("DEBUG starting receiveCommandsLoop")
	for {
		select {
		case cmd := <-c.fromServer:
			if cmd.CMD == common.KILL {
				log.Printf("Server sent KILL cmd to connected device %d", c.ID)
				break
			}
		default:
			//Continue receiveReadings loop
		}
		//TODO support multiline commands
		/*
			All Redis commands are sent as arrays of bulk strings.
			For example, the command “SET mykey ‘my value’” would be written and sent as:
			*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$8\r\nmy value\r\n
		*/

		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("ERR during reading %v", err)
			break
		}
		cmd, err := c.deserializeCommand(line)
		if err != nil {
			if err != nil {
				log.Printf("ERR  readed command text line :%s , err: %v", err)
				break
			}
		}
		c.toServer <- cmd
	}
	log.Println("DEBUG receiveCommandsLoop exit")
}

func (c *Client) deserializeCommand(serializedCMD string) (common.Command, error) {
	log.Printf("DEBUG: deserializing: %s", serializedCMD)

	cmd, data, err := resp.Deserialize(serializedCMD)
	if err != nil {
		return common.Command{}, err
	}
	return common.Command{
		CMD:       cmd,
		ClientID:  c.ID,
		Arguments: data,
	}, nil
}

func (c *Client) Read(wg *sync.WaitGroup) {
	log.Println("DEBUG starting client Read")
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("ERR trying to close the connection %v", err)
		}
		log.Println("DEBUG client connection closed")
		c.toServer <- common.Command{
			CMD:      common.DEREGISTER,
			ClientID: c.ID,
		}
		wg.Done()
	}()

	c.receiveCommandsLoop()
	log.Print("DEBUG client Read exit")

}
