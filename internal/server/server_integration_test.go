package server

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"go.uber.org/goleak"
	"sync"
	"testing"
	"time"
)

func TestBasicOps(t *testing.T) {
	defer goleak.VerifyNone(t)
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 2)
	port := uint(10_001)

	go Start(port, 1, ready, quit, events)

	<-ready
	fmt.Println("server is ready")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})

	ctx := context.Background()
	err := rdb.Set(ctx, "x", 1, 0).Err()
	common.ExpectNoError(t, err)

	val, err := rdb.Get(ctx, "x").Result()
	common.ExpectNoError(t, err)
	common.AssertEquals(t, val, "1")

	delResult, err := rdb.Del(ctx, "x").Result()
	common.ExpectNoError(t, err)
	common.AssertEquals(t, delResult, int64(1))

	val, err = rdb.Get(ctx, "x").Result()
	common.AssertEquals(t, err, redis.Nil)

	common.ExpectNoError(t, rdb.Close())
	common.AssertEquals(t, <-events, EventAfterDisconnect)
	quit <- true
	common.AssertEquals(t, <-events, EventSuccessfulShutdown)

}

func TestUnsupportedCommand(t *testing.T) {
	defer goleak.VerifyNone(t)
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 1)
	port := uint(10_002)
	go Start(port, 1, ready, quit, events)

	<-ready
	fmt.Println("server is ready")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})

	ctx := context.Background()

	err := rdb.Incr(ctx, "x").Err()
	wantError := "unsupported command [incr x]"
	if err == nil || err.Error() != wantError {
		t.Errorf("want error:%s , got: %s ", wantError, err)
	}
	common.ExpectNoError(t, rdb.Close())
	common.AssertEquals(t, <-events, EventAfterDisconnect)
	quit <- true
	common.AssertEquals(t, <-events, EventSuccessfulShutdown)

}

func TestClientConnectionsLifeCycle(t *testing.T) {
	defer goleak.VerifyNone(t)

	ready := make(chan bool, 1)
	quit := make(chan bool)
	events := make(chan string)
	port := uint(10_003)
	server := newServer(time.Now, port, 1, ready, quit, events)
	var wg sync.WaitGroup
	wg.Add(1)
	go server.run(&wg)
	<-ready
	fmt.Println("server is ready")

	ctx := context.Background()
	common.AssertEquals(t, server.numConnectedClients(), 0)
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})

	err := rdb.Set(ctx, "x", 1, 0).Err()
	common.ExpectNoError(t, err)

	common.AssertEquals(t, server.numConnectedClients(), 1)
	common.ExpectNoError(t, rdb.Close())
	event := <-events
	common.AssertEquals(t, event, EventAfterDisconnect)
	common.AssertEquals(t, server.numConnectedClients(), 0)

	quit <- true

	common.AssertEquals(t, <-events, EventSuccessfulShutdown)

}

func TestMaxClients(t *testing.T) {
	defer goleak.VerifyNone(t)

	ready := make(chan bool, 1)
	quit := make(chan bool)
	events := make(chan string)
	port := uint(10_004)
	server := newServer(time.Now, port, 2, ready, quit, events)
	var wg sync.WaitGroup
	wg.Add(1)
	go server.run(&wg)
	<-ready
	fmt.Println("server is ready")

	ctx := context.Background()

	common.AssertEquals(t, server.numConnectedClients(), 0)
	client1 := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})
	//defer client1.Close()

	err := client1.Set(ctx, "y", 99, 0).Err()
	common.ExpectNoError(t, err)
	client2 := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})
	//defer client2.Close()
	val, err := client2.Get(ctx, "y").Result()
	common.ExpectNoError(t, err)
	common.AssertEquals(t, val, "99")

	common.AssertEquals(t, server.numConnectedClients(), 2)
	client3 := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("localhost:%d", port),
		MaxRetries: 1,
	})

	err = client3.Get(ctx, "y").Err()
	if err == nil {
		t.Errorf("expecting error")
	}
	quit <- true
	common.AssertEquals(t, <-events, EventAfterDisconnect)
	common.AssertEquals(t, <-events, EventAfterDisconnect)
	common.AssertEquals(t, <-events, EventSuccessfulShutdown)
	common.ExpectNoError(t, client1.Close())
	common.ExpectNoError(t, client2.Close())
	common.ExpectNoError(t, client3.Close())

}

func TestServerLifecycle(t *testing.T) {
	defer goleak.VerifyNone(t)
	ready := make(chan bool, 1)
	quit := make(chan bool)
	events := make(chan string)
	t.Cleanup(func() {
		close(ready)
		close(quit)
		close(events)
	})
	port := uint(10_005)
	server := newServer(time.Now, port, 2, ready, quit, events)
	var wg sync.WaitGroup
	wg.Add(1)
	go server.run(&wg)
	<-ready
	fmt.Println("server is ready")

	common.AssertEquals(t, server.numConnectedClients(), 0)

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})

	ctx := context.Background()
	err := rdb.Set(ctx, "x", 1, 0).Err()
	common.ExpectNoError(t, err)
	common.AssertEquals(t, server.numConnectedClients(), 1)

	quit <- true

	common.AssertEquals(t, <-events, EventAfterDisconnect)
	common.AssertEquals(t, <-events, EventSuccessfulShutdown)
	common.ExpectNoError(t, rdb.Close())
}
