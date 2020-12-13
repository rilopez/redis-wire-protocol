package server

import (
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"github.com/rilopez/redis-wire-protocol/internal/device"
	"github.com/rilopez/redis-wire-protocol/internal/resp"
	"log"
	"net"
	"sync"
	"time"
)

// Start creates a tcp connection listener to accept connections at `port`
func Start(port uint, httpPort uint, serverMaxClients uint) {
	log.Printf("starting server demons  with \n  - port:%d\n - httpPort:%d\n -serverMaxClients: %d\n",
		port, httpPort, serverMaxClients)

	core := newServer(time.Now, port, serverMaxClients)
	httpd := newHttpd(core, httpPort)
	var wg sync.WaitGroup
	wg.Add(1)
	go core.run(&wg)
	go httpd.run()

	wg.Wait()

}

// server maintains a map of clients and communication channels
type server struct {
	clients          map[uint64]*connectedClient
	db               map[string]*string
	commands         chan common.Command
	port             uint
	nextClientId     uint64
	serverMaxClients uint
	now              func() time.Time
	mux              sync.Mutex
}

type connectedClient struct {
	ID              uint64
	callbackChannel chan common.Command
	lastCMDEpoch    int64
	lastCMD         common.CommandID
}

// NewCore allocates a Core struct
func newServer(now func() time.Time, port uint, serverMaxClients uint) *server {
	return &server{
		clients:          make(map[uint64]*connectedClient),
		db:               make(map[string]*string),
		commands:         make(chan common.Command),
		now:              now,
		port:             port,
		nextClientId:     1,
		serverMaxClients: serverMaxClients,
	}
}

func (s *server) numConnectedClients() int {
	s.mux.Lock()
	numActiveClients := len(s.clients)
	s.mux.Unlock()
	return numActiveClients
}

func (s *server) listenConnections(wg *sync.WaitGroup) {
	address := fmt.Sprintf(":%d", s.port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("ERR Failed to start tcp listener at %s,  %v", address, err)
	}
	defer func() {
		ln.Close()
		wg.Done()
	}()

	log.Printf("Server started listening for connections at %s ", address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Print("DEBUG: new client connection")
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

func (s *server) registerClient(conn net.Conn) (*device.Client, error) {
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
	client, err := device.NewClient(
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
	defer wg.Done()

	go s.listenConnections(wg)

	for {
		var err error
		var response = ""
		select {
		case cmd := <-s.commands:
			client, exists := s.clientByID(cmd.ClientID)
			if !exists {
				err = fmt.Errorf("client ID  %d does not exists", cmd.ClientID)
			}
			client.lastCMD = cmd.CMD
			client.lastCMDEpoch = s.now().UnixNano()

			switch cmd.CMD {
			case common.DEREGISTER:
				err = s.deregister(client.ID)
			case common.SET:
				response, err = s.handleSET(cmd.Arguments)
			case common.GET:
				response, err = s.handleGET(client, cmd.Arguments)
			default:
				err = fmt.Errorf("unknown Command %d", cmd.CMD)
			}

			if len(response) != 0 {
				client.callbackChannel <- common.Command{
					CMD: common.RESPONSE,
					Arguments: common.RESPONSEArguments{
						Response: response,
					},
				}
			}
			if err != nil {
				log.Printf("ERR %v", err)
			}
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
		return "", fmt.Errorf("invalid SET argments %v", args)
	}

	log.Printf("SET with args  %v", setArgs)

	s.mux.Lock()
	s.db[setArgs.Key] = &setArgs.Value
	s.mux.Unlock()

	return resp.SimpleString("OK"), nil
}

func (s *server) handleGET(client *connectedClient, args common.CommandArguments) (response string, err error) {
	getArgs, ok := args.(common.GETArguments)
	if !ok {
		return "-ERR", fmt.Errorf("invalid GET argments %v", args)
	}

	log.Printf("GET with args  %v", getArgs)

	//TODO get  key value from memory

	s.mux.Lock()
	value, exists := s.db[getArgs.Key]
	s.mux.Unlock()
	if !exists {
		return "-ERR unknown key", fmt.Errorf("invalid key %s", getArgs.Key)
	}
	return resp.BulkString(*value), nil
}

func (s *server) handleDEL(args common.CommandArguments) (err error) {
	delArgs, ok := args.(common.DELArguments)
	if !ok {
		return fmt.Errorf("invalid GET argments %v", args)
	}

	s.mux.Lock()
	for _, k := range delArgs.Keys {
		delete(s.db, k)
	}
	s.mux.Unlock()
	log.Printf("DEL with args  %v", delArgs)

	//TODO get  key value from memory

	return nil
}

func (s *server) deregister(clientID uint64) error {
	log.Printf("DEBUG trying to deregister client with ID %d ", clientID)
	s.mux.Lock()
	delete(s.clients, clientID)
	s.mux.Unlock()
	log.Printf("client with ID %d desconnected succesfuly", clientID)
	return nil
}
