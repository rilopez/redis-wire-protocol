package device

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
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
	ID       uuid.UUID
	conn     net.Conn
	outbound chan<- common.Command
	inbound  chan common.Command
	now      func() time.Time
}

// NewClient allocates a Client
func NewClient(conn net.Conn, outbound chan<- common.Command, now func() time.Time) (*Client, error) {

	client := &Client{
		ID: uuid.New(),
		conn:     conn,
		outbound: outbound,
		now:      now,
	}
	return client, nil
}

// Close terminates a connection to core.
func (c *Client) logout() error {

	c.outbound <- common.Command{
		ID:     common.LOGOUT,
		Sender: c.ID,
	}
	return nil
}

func (c *Client) receiveLoginMessage() error {
	log.Println("DEBUG: receiveLoginMessage start")
	var loginMsg [15]byte
	n, err := c.conn.Read(loginMsg[:])
	if err != nil || n < 15 {
		return fmt.Errorf("ERR trying to read IMEI, bytes read: %d, err: %v", n, err)
	}

	imei, err := decodeIMEI(loginMsg[:])
	if err != nil {
		return fmt.Errorf("ERR decoding IMEI bytes %v ", err)
	}
	c.ID = imei

	return nil
}

func (c *Client) receiveCommandsLoop() {

	reader := bufio.NewReader(c.conn)
	tp := textproto.NewReader(reader)
	log.Print("DEBUG starting receiveCommandsLoop")
	for {
		select {
		case cmd := <-c.inbound:
			if cmd.ID == common.KILL {
				log.Printf("Server sent KILL cmd to connected device %d", c.ID)
				break
			}
		default:
			//Continue receiveReadings loop
		}
		//TODO support multiline commands
		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("ERR during reading %v", err)
			c.logout()
			break
		}
		cmd := deserializeCommand(line)
		c.outbound <- cmd
	}
	log.Println("DEBUG receiveCommandsLoop exit")
}

func (c *Client)  deserializeCommand(serializedCMD string) common.Command {
	cmd, data := resp.Deserialize(serializedCMD)
	return common.Command{
		ID:     cmd,
		Sender: c.ID
		Body:   data,
	}
}


func (c *Client) Read(wg *sync.WaitGroup) {
	log.Println("DEBUG starting client Read")
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("ERR trying to close the connection %v", err)
		}
		log.Println("DEBUG client connection closed")
		wg.Done()
	}()


	c.inbound = make(chan common.Command)

	c.outbound <- common.Command{
		ID:              common.LOGIN,
		Sender:          c.ID,
		CallbackChannel: c.inbound,
	}

	c.receiveCommandsLoop()
	log.Print("DEBUG client Read exit")

}
