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
	clients          map[uint64]*connectedClient
	db               map[string]*string
	commands         chan common.Command
	ready            chan<- bool
	events           chan<- string
	quit             <-chan bool
	port             uint
	nextClientId     uint64
	serverMaxClients uint
	now              func() time.Time
	mux              sync.Mutex
	state            serverState
}

type connectedClient struct {
	ID              uint64
	callbackChannel chan common.Command
	lastCMDEpoch    int64
	lastCMD         common.CommandID
}

// NewCore allocates a Core struct
func newServer(now func() time.Time, port uint, serverMaxClients uint, ready chan<- bool, quit <-chan bool, events chan<- string) *server {
	return &server{
		clients:          make(map[uint64]*connectedClient),
		db:               make(map[string]*string),
		commands:         make(chan common.Command),
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
		ln.Close()
		wg.Done()
		log.Print("listened connections loop stopped")
	}()

	log.Printf("Server started listening for connections at %s ", fmt.Sprintf(":%d", s.port))
	s.ready <- true
	for {
		ln.SetDeadline(time.Now().Add(idleTimeout))
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-quitListening:
				log.Print("listenConnections got a quit signal ")
				return
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		} else {
			client, err := s.registerClient(conn)
			if err != nil {
				conn.Close()
				log.Printf("ERR trying to register a new client: %v", err)
				continue
			}

			log.Printf("client connection from %v", conn.RemoteAddr())
			wg.Add(1)
			go client.Read(wg)
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

	callbackChan := make(chan common.Command)
	client, err := client.NewWorker(
		conn,
		s.nextClientId,
		s.commands,
		callbackChan,
		s.now,
	)
	if err != nil {
		return nil, fmt.Errorf("ERR trying to create a client worker for the connection, %v", err)
	}

	s.clients[client.ID] = &connectedClient{
		ID:              client.ID,
		callbackChannel: callbackChan,
	}
	s.nextClientId++

	return client, nil

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
		case cmd := <-s.commands:
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
	client, exists := s.clientByID(cmd.ClientID)
	if !exists {
		err = fmt.Errorf("client ID  %d does not exists", cmd.ClientID)
	}
	client.lastCMD = cmd.CMD
	client.lastCMDEpoch = s.now().UnixNano()

	switch cmd.CMD {
	case common.SET:
		response, err = s.handleSET(cmd.Arguments)
	case common.GET:
		response, err = s.handleGET(cmd.Arguments)
	case common.DEL:
		response, err = s.handleDEL(cmd.Arguments)
	case common.INFO:
		response, err = s.handleINFO()

		//TODO support CLIENT
		//TODO support CLIENT LIST
		//TODO support PING
		//TODO support ECHO

	case common.UNKNOWN:
		err = fmt.Errorf("unsupported command %v", cmd.Arguments)
	case common.INTERNAL_DEREGISTER:
		err = s.disconnect(client.ID)
	default:
		err = fmt.Errorf("invalid server command %d", cmd.CMD)
	}
	if err != nil {
		log.Printf("ERR %v", err)
		response = resp.Error(err)
	}
	if len(response) != 0 && cmd.CallbackChannel != nil {
		cmd.CallbackChannel <- common.Command{
			CMD: common.RESPONSE,
			Arguments: common.RESPONSEArguments{
				Response: response,
			},
		}
	}
}

func (s *server) clientByID(ID uint64) (*connectedClient, bool) {
	s.mux.Lock()
	dev, exists := s.clients[ID]
	s.mux.Unlock()

	return dev, exists
}

func (s *server) handleSET(args common.CommandArguments) (response string, err error) {
	setArgs, ok := args.(common.SETArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid SET argments %v", args)
	}
	log.Printf("SET with args  %v", setArgs)
	s.mux.Lock()
	s.db[setArgs.Key] = &setArgs.Value
	s.mux.Unlock()

	return resp.SimpleString("OK"), nil
}

func (s *server) handleGET(args common.CommandArguments) (response string, err error) {
	getArgs, ok := args.(common.GETArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid GET argments %v", args)
	}
	log.Printf("GET with args  %v", getArgs)
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
	opStatus := 0
	for _, k := range delArgs.Keys {
		_, exists := s.db[k]
		if exists {
			opStatus = 1 //del cmd is successful if deletes at least one key
		}
		delete(s.db, k)
	}
	s.mux.Unlock()
	log.Printf("DEL with args  %v", delArgs)

	return resp.Integer(opStatus), nil
}

func (s *server) disconnect(clientID uint64) error {
	s.mux.Lock()
	delete(s.clients, clientID)
	s.mux.Unlock()
	log.Printf("client with ID %d desconnected succesfuly", clientID)

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
	log.Printf("trying to kill %d connected clients", len(s.clients))
	for id, c := range s.clients {
		log.Printf("sending kill signal to client with ID :%d", id)
		c.callbackChannel <- common.Command{
			CMD: common.KILL,
		}
	}
	s.mux.Unlock()
}
