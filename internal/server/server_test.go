package server

import (
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"testing"
)

func TestNewCore(t *testing.T) {
	core := newServer(common.FrozenInTime, uint(1337), 2, nil, nil, nil)
	expectedClientsLen := 0
	actualClientsLen := core.numConnectedClients()
	if actualClientsLen != expectedClientsLen {
		t.Errorf("expected len(server.client) to equal %d but got %d", expectedClientsLen, actualClientsLen)
	}
}
