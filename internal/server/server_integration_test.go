package server

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestStart(t *testing.T) {
	ready := make(chan bool, 1)
	go Start(6379, 100000, ready)

	<-ready
	fmt.Println("server is ready")
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()
	errInt := rdb.Set(ctx, "x", 1, 0).Err()
	if errInt != nil {
		t.Errorf("expected no error , got %v", errInt)
	}

	val, err := rdb.Get(ctx, "x").Result()
	if err != nil {
		t.Errorf("expected no error , got %v", err)
	}
	if val != "1" {
		t.Errorf(`want val "1" , got %v `, val)
	}
}
