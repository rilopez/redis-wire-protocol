package client

import (
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"net"
	"testing"
)

func TestNewWorker(t *testing.T) {
	conn := &net.TCPConn{}
	quit := make(<-chan bool)
	outbound := make(chan<- common.Command)
	inbound := make(chan common.Command)
	worker, err := NewWorker(conn, 123, outbound, inbound, common.FrozenInTime, quit)
	common.ExpectNoError(t, err)
	common.AssertEquals(t, worker.ID, uint64(123))
	common.AssertEquals(t, worker.quit, quit)
	common.AssertEquals(t, worker.toServer, outbound)
	common.AssertEquals(t, worker.fromServer, inbound)
	common.AssertEquals(t, worker.now().String(), common.FrozenInTime().String())
}
