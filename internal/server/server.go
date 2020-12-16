package server

import (
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/client"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"github.com/rilopez/redis-wire-protocol/internal/resp"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

type serverState int

// used to time out the connection accept loop. default 2 secs will timeout accepting connections and check the quit channel
const idleTimeout = 2 * time.Second
const (
	serverStateListening serverState = iota
	serverStateShuttingDown
	serverStateBooting
)

const (
	EventAfterDisconnect    = "AFTER_DISCONNECT"
	EventSuccessfulShutdown = "SUCCESSFUL_SHUTDOWN"
)

// Start creates a tcp connection listener to accept connections at `port`
func Start(port uint, serverMaxClients uint, ready chan<- bool, quit <-chan bool, events chan<- string) {
	log.Printf("starting server demons  with \n  - port:%d\n - -serverMaxClients: %d\n",
		port, serverMaxClients)

	core := newServer(time.Now, port, serverMaxClients, ready, quit, events)

	var wg sync.WaitGroup
	wg.Add(1)
	go core.run(&wg)

	wg.Wait()

}

// server maintains a map of clients and communication channels
type server struct {
	clients          map[uint]*connectedClient
	db               map[string]*string
	requests         chan common.Command
	ready            chan<- bool
	events           chan<- string
	quit             <-chan bool
	port             uint
	nextClientId     uint
	serverMaxClients uint
	now              func() time.Time
	mux              sync.Mutex
	state            serverState
}

type connectedClient struct {
	ID             uint
	response       chan<- string
	lastCMDEpoch   int64
	lastCMD        common.CommandID
	quit           chan<- bool
	addr           string
	connectedSince time.Time
}

func (c connectedClient) info(now func() time.Time) string {
	age := now().Sub(c.connectedSince)

	return fmt.Sprintf("id:%d addr:%s age:%f cmd:%d", c.ID, c.addr, age.Seconds(), c.lastCMD)
}

// NewCore allocates a Core struct
func newServer(now func() time.Time, port uint, serverMaxClients uint, ready chan<- bool, quit <-chan bool, events chan<- string) *server {
	return &server{
		clients:          make(map[uint]*connectedClient),
		db:               make(map[string]*string),
		requests:         make(chan common.Command),
		events:           events,
		ready:            ready,
		quit:             quit,
		now:              now,
		port:             port,
		nextClientId:     1,
		serverMaxClients: serverMaxClients,
		state:            serverStateBooting,
	}
}

func (s *server) numConnectedClients() int {
	s.mux.Lock()
	numActiveClients := len(s.clients)
	s.mux.Unlock()
	return numActiveClients
}

func (s *server) listenConnections(wg *sync.WaitGroup, quitListening chan bool) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenTCP("tcp", addr)

	if err != nil {
		log.Fatalf("ERR Failed to start tcp listener at %s,  %v", fmt.Sprintf(":%d", s.port), err)
	}
	defer func() {
		if err = ln.Close(); err != nil {
			log.Fatal(err)
		}
		wg.Done()
		log.Print("listened connections loop stopped")
	}()

	log.Printf("Server started listening for connections at %s ", fmt.Sprintf(":%d", s.port))
	s.ready <- true
	for {
		if err := ln.SetDeadline(time.Now().Add(idleTimeout)); err != nil {
			log.Fatal(err)
		}
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-quitListening:
				log.Print("listenConnections got a quit signal ")
				return
			default:

				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				} else {
					panic(err)
				}

			}
		} else {
			c, err := s.registerClient(conn)
			if err != nil {
				log.Printf("ERR trying to register a new client: %v", err)
				if err = conn.Close(); err != nil {
					log.Printf("ERR %v", err)
				}
				continue
			}

			log.Printf("client connection from %v", conn.RemoteAddr())
			wg.Add(1)
			go c.Read(wg)
		}

	}

}

func (s *server) registerClient(conn net.Conn) (*client.Worker, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	numActiveClients := uint(len(s.clients))
	if numActiveClients >= s.serverMaxClients {
		// Limit the number of active clients to prevent resource exhaustion
		return nil, fmt.Errorf("ERR reached serverMaxClients:%d, there are already %d connected clients", s.serverMaxClients, numActiveClients)
	}

	if _, exists := s.clients[s.nextClientId]; exists {
		log.Panicf("duplicated client ID %d", s.nextClientId)
	}

	response := make(chan string)
	quit := make(chan bool)
	worker, err := client.NewWorker(
		conn,
		s.nextClientId,
		s.requests,
		response,
		s.now,
		quit,
	)
	if err != nil {
		return nil, fmt.Errorf("ERR trying to create a client worker for the connection, %v", err)
	}

	s.clients[worker.ID] = &connectedClient{
		connectedSince: s.now(),
		addr:           conn.RemoteAddr().String(),
		ID:             worker.ID,
		response:       response,
		quit:           quit,
	}
	s.nextClientId++

	return worker, nil

}

// Run handles channels inbound communications from connected clients
func (s *server) run(wg *sync.WaitGroup) {
	wg.Add(1)

	s.setState(serverStateListening)
	stopListening := make(chan bool)
	defer func() {
		wg.Done()
		close(stopListening)
	}()

	go s.listenConnections(wg, stopListening)

	for {
		var err error
		var response = ""
		select {
		case <-s.quit:
			log.Print("got quit signal trying to shutdown connected clients")
			stopListening <- true
			s.shutdown()
		default:
		}

		select {
		case cmd := <-s.requests:
			s.handleCMD(cmd, err, response)
		default:
		}

		if s.getState() == serverStateShuttingDown && s.numConnectedClients() == 0 {
			log.Print("no more clients connected, exit now")
			s.events <- EventSuccessfulShutdown
			return
		}
	}
}

func (s *server) setState(state serverState) {
	s.mux.Lock()
	s.state = state
	s.mux.Unlock()
}
func (s *server) getState() serverState {
	s.mux.Lock()
	state := s.state
	s.mux.Unlock()
	return state
}

func (s *server) handleCMD(cmd common.Command, err error, response string) {
	c, exists := s.clientByID(cmd.ClientID)
	if !exists {
		err = fmt.Errorf("client ID  %d does not exists", cmd.ClientID)
	}
	c.lastCMD = cmd.CMD
	c.lastCMDEpoch = s.now().UnixNano()

	switch cmd.CMD {
	case common.SET:
		response, err = s.handleSET(cmd.Arguments)
	case common.GET:
		response, err = s.handleGET(cmd.Arguments)
	case common.DEL:
		response, err = s.handleDEL(cmd.Arguments)
	case common.INFO:
		response, err = s.handleINFO()
	case common.CLIENT:
		response, err = s.handleCLIENT(cmd.Arguments, c)
	case common.UNKNOWN:
		err = fmt.Errorf("unsupported command %v", cmd.Arguments)
	default:
		err = fmt.Errorf("invalid server command %d", cmd.CMD)
	}
	if err != nil {
		log.Printf("ERR %v", err)
		response = resp.Error(err)
	}
	if len(response) != 0 {
		c.response <- response
	}
}

func (s *server) clientByID(ID uint) (*connectedClient, bool) {
	s.mux.Lock()
	dev, exists := s.clients[ID]
	s.mux.Unlock()

	return dev, exists
}

func (s *server) handleCLIENT(args common.CommandArguments, c *connectedClient) (response string, err error) {
	clientArgs, ok := args.(common.CLIENTArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid CLIENT argments %v", args)
	}
	switch clientArgs.Subcommand {
	case common.ClientSubcommandKILL:
		return "", s.disconnect(c.ID)
	case common.ClientSubcommandID:
		return resp.Integer(int(c.ID)), nil
	case common.ClientSubcommandLIST:
		var sb strings.Builder
		s.mux.Lock()
		for _, c := range s.clients {
			sb.WriteString(c.info(s.now))
			sb.WriteString("\n")
		}
		s.mux.Unlock()
		return resp.SimpleString(sb.String()), nil

	case common.ClientSubcommandINFO:
		return resp.SimpleString(fmt.Sprintf("%s\n", c.info(s.now))), nil

	default:
		return "-ERR", fmt.Errorf("unsupported CLIENT subcommand %v", args)
	}
}

func (s *server) handleSET(args common.CommandArguments) (response string, err error) {
	setArgs, ok := args.(common.SETArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid SET argments %v", args)
	}
	var prevValue *string
	needToSet := false
	response = resp.BulkString(nil)
	s.mux.Lock()
	prevValue, ok = s.db[setArgs.Key]

	if setArgs.OptionNX && !ok {
		//Only set the key if it does not already exist.
		needToSet = true
	} else if setArgs.OptionXX && ok {
		//Only set the key if it already exist.
		needToSet = true
	} else if !setArgs.OptionNX && !setArgs.OptionXX {
		needToSet = true
	}
	if needToSet {
		s.db[setArgs.Key] = &setArgs.Value
		response = resp.SimpleString("OK")
	}

	s.mux.Unlock()

	if setArgs.OptionGET {
		response = resp.BulkString(prevValue)
	}
	return
}

func (s *server) handleGET(args common.CommandArguments) (response string, err error) {
	getArgs, ok := args.(common.GETArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid GET argments %v", args)
	}
	s.mux.Lock()
	value, _ := s.db[getArgs.Key]
	s.mux.Unlock()

	return resp.BulkString(value), nil
}

func (s *server) handleDEL(args common.CommandArguments) (response string, err error) {
	delArgs, ok := args.(common.DELArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid GET argments %v", delArgs)
	}

	s.mux.Lock()
	var opStatus = 0
	for _, k := range delArgs.Keys {
		_, exists := s.db[k]
		if exists {
			opStatus = 1 //del cmd is successful if deletes at least one key
		}
		delete(s.db, k)
	}
	s.mux.Unlock()

	return resp.Integer(opStatus), nil
}

func (s *server) disconnect(clientID uint) error {
	s.mux.Lock()
	delete(s.clients, clientID)
	s.mux.Unlock()
	log.Printf("client with ID %d disconnected succesfuly", clientID)

	s.events <- EventAfterDisconnect
	return nil
}

func (s *server) handleINFO() (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("NumConnectedClients:%d\n", s.numConnectedClients()))
	sb.WriteString(fmt.Sprintf("NumCPU:%d\n", runtime.NumCPU()))
	sb.WriteString(fmt.Sprintf("NumGoroutine:%d\n", runtime.NumGoroutine()))

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	sb.WriteString("=== MemStats === \n")
	sb.WriteString(fmt.Sprintf("Alloc:%d\n", memStats.Alloc))
	sb.WriteString(fmt.Sprintf("TotalAlloc:%d\n", memStats.TotalAlloc))
	sb.WriteString(fmt.Sprintf("Sys:%d\n", memStats.Sys))
	sb.WriteString(fmt.Sprintf("Mallocs:%d\n", memStats.Mallocs))
	sb.WriteString(fmt.Sprintf("Frees:%d\n", memStats.Frees))
	sb.WriteString(fmt.Sprintf("Live Objects(Mallocs - Frees):%d\n", memStats.Mallocs-memStats.Frees))

	sb.WriteString(fmt.Sprintf("PauseTotalNs:%d\n", memStats.PauseTotalNs))
	sb.WriteString(fmt.Sprintf("NumGC:%d\n", memStats.NumGC))

	str := sb.String()
	return resp.BulkString(&str), nil
}

func (s *server) shutdown() {
	s.setState(serverStateShuttingDown)
	s.mux.Lock()
	clients := s.clients
	s.mux.Unlock()
	log.Printf("trying to kill %d connected clients", len(s.clients))
	for id, c := range clients {
		log.Printf("sending kill signal to client with ID :%d", id)
		if c.quit == nil {
			log.Printf("client callback channel is nil")
		}
		c.quit <- true
	}
}
