package common

import "github.com/google/uuid"

// CommandID command id type
type CommandID int

const (
	// LOGIN used when client connect to our server
	LOGIN CommandID = iota
	// LOGOUT used to indicate the client to log itself out
	LOGOUT
	// KILL used to indicate a client to terminate its reading loop
	KILL
	// READING sent by the server after a succesfull login
	WELCOME

	// Redis commands

	// GET https://redis.io/commands/get
	GET

	// SET https://redis.io/commands/set
	SET

	// DEL https://redis.io/commands/del
	DEL
)

// Command is used to send data between clients and server core
type Command struct {
	ID              CommandID
	Sender          uuid.UUID
	CallbackChannel chan Command
	Body            []byte
}
