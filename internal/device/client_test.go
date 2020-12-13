package device

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
)

func TestRead(t *testing.T) {
	t.Skip("NEEDS REIMPLEMENTATION")
	timeout := time.After(1 * time.Second)
	outbound := make(chan common.Command, 1)
	expectedIMEI := uint64(490154203237518)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("ERR while trying to start testing server, %v", err)
	}

	var wg sync.WaitGroup
	go func() {
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			t.Errorf("ERR while getting a connection, %v", err)

		}
		client, err := NewClient(conn, outbound, common.FrozenInTime)
		client.Read(&wg)

	}()

	conn, err := net.Dial("tcp", ln.Addr().String())

	expectedIMEIbytes := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	conn.Write(expectedIMEIbytes)
	readingBytes := CreateRandReadingBytes()
	conn.Write(readingBytes[:])

	conn.Close()
	select {
	case <-timeout:
		t.Fatal("Timeout")
	case cmd := <-outbound:
		switch cmd.ID {
		case common.LOGOUT:
			if cmd.Sender != expectedIMEI {
				t.Errorf("expected client device with IMEI %v was not sent to logout channel  got %v", expectedIMEI, cmd.Sender)
			}

		case common.LOGIN:
			if cmd.Sender != expectedIMEI {
				t.Errorf("expecterd client device with IMEI %v was not sent to login channel  got %v", expectedIMEI, cmd.Sender)
			}

		}
	}
}
