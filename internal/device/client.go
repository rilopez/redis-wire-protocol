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
	writer := bufio.NewWriter(c.conn)
	tp := textproto.NewReader(reader)
	log.Print("DEBUG starting receiveCommandsLoop")
	for {
		select {
		case cmd := <-c.fromServer:
			if cmd.CMD == common.KILL {
				log.Printf("Server sent KILL cmd to connected device %d", c.ID)
				break
			} else if cmd.CMD == common.RESPONSE {
				v, ok := cmd.Arguments.(common.RESPONSEArguments)
				if !ok {
					log.Panicf("invalid response arguments %v", cmd.Arguments)
				}
				writer.WriteString(v.Response)
			}
		default:
			//Continue receiveReadings loop
		}
		cmd, err := c.readCommand(tp)
		if err != nil {
			if err != nil {
				log.Printf("ERR  readCommand :%v ", err)
				break
			}
		}
		c.toServer <- cmd
	}
	log.Println("DEBUG receiveCommandsLoop exit")
}

func (c *Client) readCommand(reader *textproto.Reader) (common.Command, error) {
	cmd, data, err := resp.DeserializeCMD(reader)
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
