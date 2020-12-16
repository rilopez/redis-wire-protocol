/*
Package server provides functionality for a  redis protocol  over TCP. The server supports this limited set of commands

- SET key value [NX|XX] [GET]
- GET key
- DEL key [key ...]
- INFO
- CLIENT [KILL | INFO | ID | LIST]


The TCP redis server uses goroutines to handle each connected
client.  These channels are used to communicate the client data to the server server

	requests chan common.Command
	    used to by client connections to send cmd & data to server server.
	    currently the server implements SET, GET & DEL commands

    connectedClient.response chan string

*/
package server
