package main

import (
	"flag"
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//Logging messages are written to os.Stderr.
	log.SetOutput(os.Stderr)

	serverPort := flag.Uint("port", 6379, "port number to listen for TCP connections of clients implementing the redis protocol")
	serverMaxClients := flag.Uint("max-clients", 100_000, "Max number of clients accepted by the server ")

	flag.Parse()
	ready := make(chan bool)
	quit := make(chan bool)
	events := make(chan string)
	osSignal := make(chan os.Signal)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-osSignal
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		quit <- true
	}()

	go func() {
		for {
			select {
			case <-ready:
				fmt.Println("Server Ready")
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
	close(osSignal)
}
