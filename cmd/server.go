package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rilopez/redis-wire-protocol/internal/server"
)

func main() {
	//Logging messages are written to os.Stderr.
	log.SetOutput(os.Stderr)

	serverPort := flag.Uint("port", 6379, "port number to listen for TCP connections of clients implementing the redis protocol")
	serverMaxClients := flag.Uint("max-clients", 100000, "Max number of clients accepted by the server ")

	flag.Parse()
	ready := make(chan bool, 1)
	quit := make(chan bool, 1)
	events := make(chan string, 100)
	go func() {
		for {
			select {
			case event := <-events:
				if event == server.EventSuccessfulShutdown {
					fmt.Println("Bye")
					return
				}
			default:

			}
		}
	}()
	server.Start(*serverPort, *serverMaxClients, ready, quit, events)
	close(events)
	close(quit)
	close(ready)
}
