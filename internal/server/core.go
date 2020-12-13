package server

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
	"time"

	"github.com/rilopez/redis-wire-protocol/internal/common"
	"github.com/rilopez/redis-wire-protocol/internal/device"
)

// core maintains a map of clients and communication channels
type core struct {
	clients          map[uuid.UUID]*connectedClient
	commands         chan common.Command
	port             uint
	serverMaxClients uint
	now              func() time.Time
	mux              sync.Mutex
}

type connectedClient struct {
	callbackChannel  chan common.Command
	lastReadingEpoch int64
	lastReading      *device.Reading
}

// NewCore allocates a Core struct
func newCore(now func() time.Time, port uint, serverMaxClients uint) *core {
	return &core{
		clients:          make(map[uuid.UUID]*connectedClient),
		commands:         make(chan common.Command),
		now:              now,
		port:             port,
		serverMaxClients: serverMaxClients,
	}
}

func (c *core) numConnectedDevices() int {
	c.mux.Lock()
	numActiveClients := len(c.clients)
	c.mux.Unlock()
	return numActiveClients
}

func (c *core) listenConnections(wg *sync.WaitGroup) {
	address := fmt.Sprintf(":%d", c.port)
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
		numActiveClients := c.numConnectedDevices()
		if uint(numActiveClients) >= c.serverMaxClients {
			// Limit the number of active clients to prevent resource exhaustion
			log.Printf("ERR reached serverMaxClients:%d, there are already %d connected clients", c.serverMaxClients, numActiveClients)
			conn.Close()

		} else {
			log.Printf("client connection from %v", conn.RemoteAddr())
			client, err := device.NewClient(
				conn,
				c.commands,
				c.now,
			)
			if err != nil {
				conn.Close()
				log.Printf("ERR trying to create a client worker for the connection, %v", err)
				continue
			}
			wg.Add(1)
			go client.Read(wg)
		}

	}

}

// Run handles channels inbound communications from connected clients
func (c *core) run(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	go c.listenConnections(wg)

	for {
		var err error
		select {
		case cmd := <-c.commands:
			switch cmd.ID {
			case common.LOGIN:
				err = c.register(cmd.Sender, cmd.CallbackChannel)
			case common.LOGOUT:
				err = c.deregister(cmd.Sender)
			case common.SET:
				err = c.handleSET(cmd.Sender, cmd.Arguments)
			default:
				err = fmt.Errorf("Unknown Command %d", cmd.ID)
			}
		}
		if err != nil {
			log.Printf("ERR %v", err)
		}

	}

}

func (c *core) clientByID(ID uuid.UUID) (*connectedClient, bool) {
	c.mux.Lock()
	dev, exists := c.clients[ID]
	c.mux.Unlock()

	return dev, exists
}

func (c *core) handleSET(ID uuid.UUID, payload interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ERR recovering from hadleReading panic %v", r)
		}
	}()

	log.Printf("payload  %v", payload)
	reading := &device.Reading{}
	client, exists := c.clientByID(ID)
	if !exists {
		return fmt.Errorf("Client with IMEI %d does not exists", ID)
	}
	client.lastReadingEpoch = c.now().UnixNano()
	client.lastReading = reading

	fmt.Println(formatReadingOutput(ID, client.lastReadingEpoch, client.lastReading))

	return nil
}

func formatReadingOutput(ID uuid.UUID, lastReadingEpoch int64, lastReading *device.Reading) string {
	return fmt.Sprintf("%d,%d,%f,%f,%f,%f,%f",
		lastReadingEpoch,
		ID,
		lastReading.Temperature,
		lastReading.Altitude,
		lastReading.Latitude,
		lastReading.Longitude,
		lastReading.BatteryLevel)

}

func (c *core) register(ID uuid.UUID, callbackChannel chan common.Command) error {

	_, exists := c.clientByID(ID)

	if exists {
		log.Printf("DEBUG trying to kill connected dup device %v", ID)
		callbackChannel <- common.Command{ID: common.KILL}
		log.Printf("DEBUG KILL cmd sent  %v", ID)
		return fmt.Errorf("imei %d already logged in", ID)
	}
	c.mux.Lock()
	c.clients[ID] = &connectedClient{
		callbackChannel: callbackChannel,
	}
	c.mux.Unlock()
	callbackChannel <- common.Command{ID: common.WELCOME}
	log.Printf("device with IMEI %d connected succesfuly", ID)

	return nil
}

func (c *core) deregister(ID uuid.UUID) error {
	log.Printf("DEBUG trying to deregister device with IMEI %d ", ID)
	_, exists := c.clientByID(ID)
	if !exists {
		return fmt.Errorf("ERR imei %d is not logged in", ID)
	}

	c.mux.Lock()
	delete(c.clients, ID)
	c.mux.Unlock()
	log.Printf("device with IMEI %d desconnected succesfuly", ID)
	return nil
}
