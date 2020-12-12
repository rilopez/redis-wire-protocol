package server

import (
	"log"
	"sync"
	"time"
)

// Start creates a tcp connection listener to accept connections at `port`
func Start(port uint, httpPort uint, serverMaxClients uint) {
	log.Printf("starting server demons  with \n  - port:%d\n - httpPort:%d\n -serverMaxClients: %d\n",
		port, httpPort, serverMaxClients)

	core := newCore(time.Now, port, serverMaxClients)
	httpd := newHttpd(core, httpPort)
	var wg sync.WaitGroup
	wg.Add(1)
	go core.run(&wg)
	go httpd.run()

	wg.Wait()

}
