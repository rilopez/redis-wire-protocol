package client

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/resp"
	"io"
	"log"
	"net"
	"net/textproto"
	"sync"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

const idleTimeout = 20 * time.Millisecond

// Worker is used to handle a client connection
type Worker struct {
	ID       uint
	conn     net.Conn
	request  chan<- common.Command
	response <-chan string
	quit     <-chan bool
	now      func() time.Time
}

// NewWorker allocates a Worker
func NewWorker(conn net.Conn, ID uint, request chan<- common.Command, response <-chan string, now func() time.Time, quit <-chan bool) (*Worker, error) {
	if conn == nil {
		return nil, fmt.Errorf("conn can not be nil")
	}
	if request == nil {
		return nil, fmt.Errorf("request chan can not be nil")
	}
	if response == nil {
		return nil, fmt.Errorf("response chan can not be nil")
	}
	if quit == nil {
		return nil, fmt.Errorf("quit chan can not be nil")
	}
	if now == nil {
		return nil, fmt.Errorf("now function can not be nil")
	}

	client := &Worker{
		ID:       ID,
		conn:     conn,
		request:  request,
		response: response,
		quit:     quit,
		now:      now,
	}
	return client, nil
}

func (c *Worker) receiveCommandsLoop() {
	reader := bufio.NewReader(c.conn)
	writer := bufio.NewWriter(c.conn)
	tp := textproto.NewReader(reader)
	for {
		if err := c.conn.SetReadDeadline(time.Now().Add(idleTimeout)); err != nil {
			log.Fatal(err)
		}
		cmd, err := c.readCommand(tp)
		if err != nil {

			select {
			case <-c.quit:
				log.Printf("worker with ID %d got quick signal stopping reading loop ", c.ID)
				return
			default:
				if errTimeout, ok := err.(net.Error); ok && errTimeout.Timeout() {
					continue
				} else {
					if errors.Is(err, io.EOF) {
						log.Printf("ERR  client connection EOF ")
					} else {
						log.Printf("ERR  readCommand :%v ", err)
					}
					return
				}
			}
		} else {
			c.request <- cmd
			_, err := writer.WriteString(<-c.response)
			if err != nil {
				log.Printf("ERR writing to connection %v ", err)
			}
			err = writer.Flush()
			if err != nil {
				log.Printf("ERR trying to flush response %v ", err)
			}

		}
	}
}

func (c *Worker) readCommand(reader *textproto.Reader) (common.Command, error) {
	cmd, data, err := resp.DeserializeCMD(reader)

	return common.Command{
		CMD:       cmd,
		ClientID:  c.ID,
		Arguments: data,
	}, err
}

func (c *Worker) Read(wg *sync.WaitGroup) {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("ERR trying to close the connection %v", err)
		}

		c.request <- common.Command{
			CMD:       common.CLIENT,
			Arguments: common.CLIENTArguments{Subcommand: common.ClientSubcommandKILL},
			ClientID:  c.ID,
		}
		wg.Done()
	}()

	c.receiveCommandsLoop()
}
