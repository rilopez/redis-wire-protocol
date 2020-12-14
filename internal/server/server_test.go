package server

import (
	"testing"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

func TestNewCore(t *testing.T) {
	core := newServer(common.FrozenInTime, uint(1337), 2)
	expectedClientsLen := 0
	actualClientsLen := core.numConnectedClients()
	if actualClientsLen != expectedClientsLen {
		t.Errorf("expected len(server.client) to equal %d but got %d", expectedClientsLen, actualClientsLen)
	}
}

//TODO create integration test with for Start func

//TODO create benchmark test with for Start func
