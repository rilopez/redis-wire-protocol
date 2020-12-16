package client

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

const idleTimeout = 500 * time.Millisecond

// Worker is used to handle a client connection
type Worker struct {
	ID         uint64
	conn       net.Conn
	toServer   chan<- common.Command
	fromServer chan common.Command
	quit       <-chan bool
	now        func() time.Time
}

// NewWorker allocates a Worker
func NewWorker(conn net.Conn, ID uint64, outbound chan<- common.Command, inbound chan common.Command, now func() time.Time, quit <-chan bool) (*Worker, error) {
	client := &Worker{
		ID:         ID,
		conn:       conn,
		toServer:   outbound,
		fromServer: inbound,
		quit:       quit,
		now:        now,
	}
	return client, nil
}

func (c *Worker) receiveCommandsLoop() {
	reader := bufio.NewReader(c.conn)
	writer := bufio.NewWriter(c.conn)
	tp := textproto.NewReader(reader)
	for {
		c.conn.SetReadDeadline(time.Now().Add(idleTimeout))
		cmd, err := c.readCommand(tp)
		if err != nil {

			select {
			case <-c.quit:
				log.Printf("worker with ID %d got quick signal stopping reading loop ", c.ID)
				return
			default:
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				} else {
					log.Printf("ERR  readCommand :%v ", err)
					return
				}
			}
		} else {
			c.toServer <- cmd
			cmdResponse := <-c.fromServer
			if cmdResponse.CMD == common.RESPONSE {
				v, ok := cmdResponse.Arguments.(common.RESPONSEArguments)
				if !ok {
					log.Panicf("invalid response arguments %v", cmdResponse.Arguments)
				}
				writer.WriteString(v.Response)
				writer.Flush()
			} else {
				log.Printf("ERR [worker] unsupported cmd from server %v", cmdResponse)
				return
			}
		}
	}
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

		c.toServer <- common.Command{
			CMD:      common.INTERNAL_DEREGISTER,
			ClientID: c.ID,
		}
		wg.Done()
	}()

	c.receiveCommandsLoop()
}
