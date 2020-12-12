/*
Package server provides functionality for two kind of servers
 - TCP redis protocol (GET, SET, DEL commands only)
 - HTTP json healthcheck endpoints

The TCP redis server uses goroutines to handle each connected
client.  3 channels are used to communicate the client data to the server core

	commands chan common.Command
	    used to by client connections to send cmd & data to server core.
	    currently the server implements SET, GET & DEL commands
	Logouts  chan *device.Client
		used to send clients with closed or timeout connections. The reciever
		should remove the record from the connected clients map
	Logins   chan *device.Client
		after a new connection is created this channel is used
		to send a newly created  client so the receiver can store
		its reference in the connected clients map

These HTTP are the implemented json endpoints

  - `GET /stats`: returns a JSON document which contains runtime statistical
     information about the server (i.e. number of goroutines, bytes read per second, etc.).

*/
package server
