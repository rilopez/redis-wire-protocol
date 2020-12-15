package server

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"sync"
	"testing"
	"time"
)

func TestBasicOps(t *testing.T) {
	t.Parallel()
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 1)
	port := uint(10_001)
	go Start(port, 1, ready, quit, events)

	<-ready
	fmt.Println("server is ready")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})
	t.Cleanup(func() {
		rdb.Close()
	})
	ctx := context.Background()
	err := rdb.Set(ctx, "x", 1, 0).Err()
	common.ExpectNoError(t, err)

	val, err := rdb.Get(ctx, "x").Result()
	common.ExpectNoError(t, err)
	common.AssertEquals(t, val, "1")

	delResult, err := rdb.Del(ctx, "x").Result()
	common.ExpectNoError(t, err)
	if delResult != 1 {
		t.Errorf(`want val equals 1 , got %v `, val)
	}

	val, err = rdb.Get(ctx, "x").Result()

	if err != redis.Nil {
		t.Errorf(`want err "%s" , got %v `, redis.Nil, err)
	}

}

func TestUnsupportedCommand(t *testing.T) {
	t.Parallel()
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
	t.Cleanup(func() {
		rdb.Close()
	})
	ctx := context.Background()

	err := rdb.Incr(ctx, "x").Err()
	wantError := "unsupported command [incr x]"
	if err.Error() != wantError {
		t.Errorf("want error:%s , got: %s ", wantError, err.Error())
	}

}

func TestClientConnectionsLifeCycle(t *testing.T) {
	t.Parallel()
	//defer goleak.VerifyNone(t)

	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 1)
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
	rdb.Close()
	event := <-events
	common.AssertEquals(t, event, EventAfterDisconnect)
	common.AssertEquals(t, server.numConnectedClients(), 0)

	//quit <- true
	//serverEvent:= <-events
	//common.AssertEquals(t, serverEvent, EventSuccessfulShutdown)

}

func TestMaxClients(t *testing.T) {
	t.Parallel()
	//defer goleak.VerifyNone(t)

	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 1)
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

	err := client1.Set(ctx, "y", 99, 0).Err()
	common.ExpectNoError(t, err)
	client2 := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", port),
	})

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
	//quit <- true
	//serverEvent:= <-events
	//common.AssertEquals(t, serverEvent, EventSuccessfulShutdown)

}

func TestServerLifecycle(t *testing.T) {
	t.Skip()
	t.Parallel()
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 1)
	port := uint(10_005)
	server := newServer(time.Now, port, 2, ready, quit, events)
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

	quit <- true
	event := <-events
	common.AssertEquals(t, event, EventAfterDisconnect)
	common.AssertEquals(t, server.numConnectedClients(), 0)

}

//TODO  benchmark
