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
	initCommandLineInterface(serverCommandHandler)
}

type serverHandler func(port uint, httpPort uint, serverMaxClients uint)

func initCommandLineInterface(handleServerCmd serverHandler) {

	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", 6379, "port number to listen for TCP connections of clients implementing the redis protocol")
	serverMaxClients := serverCmd.Uint("max-clients", 100000, "Max number of clients accepted by the server ")
	serverHTTPPort := serverCmd.Uint("http-port", 80, "port number to listen for HTTP connections used mainly for healthchecks")

	if len(os.Args) < 2 {
		fmt.Println("server subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		err := serverCmd.Parse(os.Args[2:])
		if err != nil {
			serverCmd.Usage()
		}
		handleServerCmd(*serverPort, *serverHTTPPort, *serverMaxClients)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func serverCommandHandler(port uint, httpPort uint, serverMaxClients uint) {
	server.Start(port, httpPort, serverMaxClients)
}
