package client

import (
	"bufio"
	"github.com/rilopez/redis-wire-protocol/internal/resp"
	"io"
	"log"
	"net"
	"net/textproto"
	"sync"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

// Worker is used to handle a client connection
type Worker struct {
	ID         uint64
	conn       net.Conn
	toServer   chan<- common.Command
	fromServer chan common.Command
	now        func() time.Time
}

// NewWorker allocates a Worker
func NewWorker(conn net.Conn, ID uint64, outbound chan<- common.Command, inbound chan common.Command, now func() time.Time) (*Worker, error) {
	client := &Worker{
		ID:         ID,
		conn:       conn,
		toServer:   outbound,
		fromServer: inbound,
		now:        now,
	}
	return client, nil
}

func (c *Worker) receiveCommandsLoop() {
	reader := bufio.NewReader(c.conn)
	writer := bufio.NewWriter(c.conn)
	tp := textproto.NewReader(reader)
	for {
		cmd, err := c.readCommand(tp)
		if err != nil {
			if err != nil {
				if err == io.EOF {
					log.Printf("EOF :%v ", err)
				} else {
					log.Printf("ERR  readCommand :%v ", err)
				}
				break
			}
		}
		c.toServer <- cmd
		cmdResponse := <-c.fromServer
		if cmdResponse.CMD == common.RESPONSE {
			v, ok := cmdResponse.Arguments.(common.RESPONSEArguments)
			if !ok {
				log.Panicf("invalid response arguments %v", cmdResponse.Arguments)
			}
			writer.WriteString(v.Response)
			writer.Flush()
		}
	}
	log.Println("DEBUG receiveCommandsLoop exit")
}

func (c *Worker) readCommand(reader *textproto.Reader) (common.Command, error) {
	cmd, data, err := resp.DeserializeCMD(reader)

	return common.Command{
		CMD:             cmd,
		ClientID:        c.ID,
		Arguments:       data,
		CallbackChannel: c.fromServer,
	}, err
}

func (c *Worker) Read(wg *sync.WaitGroup) {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("ERR trying to close the connection %v", err)
		}
		log.Println("DEBUG client connection closed")
		c.toServer <- common.Command{
			CMD:      common.INTERNAL_DEREGISTER,
			ClientID: c.ID,
		}
		wg.Done()
	}()

	c.receiveCommandsLoop()
	log.Print("DEBUG client Read exit")

}
