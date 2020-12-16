/*
Package server provides functionality for two kind of servers
 - TCP redis protocol (GET, SET, DEL commands only)
 - HTTP json healthcheck endpoints

The TCP redis server uses goroutines to handle each connected
client.  These channels are used to communicate the client data to the server server

	requests chan common.Command
	    used to by client connections to send cmd & data to server server.
	    currently the server implements SET, GET & DEL commands


These HTTP are the implemented json endpoints

  - `GET /stats`: returns a JSON document which contains runtime statistical
     information about the server (i.e. number of goroutines, bytes read per second, etc.).

*/
package server
