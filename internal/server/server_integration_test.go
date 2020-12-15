package server

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"testing"
)

func TestSCodeChallengeSucess(t *testing.T) {
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	go Start(6379, 100000, ready, quit)

	<-ready
	fmt.Println("server is ready")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	err := rdb.Set(ctx, "x", 1, 0).Err()
	common.ExpectNoError(t, err)

	val, err := rdb.Get(ctx, "x").Result()
	common.ExpectNoError(t, err)

	if val != "1" {
		t.Errorf(`want val equals "1" , got %v `, val)
	}

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

func TestUnsupportedCommands(t *testing.T) {
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	go Start(6379, 100000, ready, quit)

	<-ready
	fmt.Println("server is ready")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	err := rdb.Incr(ctx, "x").Err()
	t.Error(err)
}
