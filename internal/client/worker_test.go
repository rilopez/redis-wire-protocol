package client

import (
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"net"
	"testing"
)

func TestNewWorker(t *testing.T) {
	conn := &net.TCPConn{}
	quit := make(<-chan bool)
	request := make(chan<- common.Command)
	response := make(chan string)
	worker, err := NewWorker(conn, 123, request, response, common.FrozenInTime, quit)
	common.ExpectNoError(t, err)
	common.AssertEquals(t, worker.ID, uint(123))
	common.AssertEquals(t, worker.quit, quit)
	common.AssertEquals(t, worker.request, request)
	common.AssertEquals(t, worker.response, response)
	common.AssertEquals(t, worker.now().String(), common.FrozenInTime().String())
}
