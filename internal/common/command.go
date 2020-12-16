package common

import "fmt"

// CommandID command id type
type CommandID int

const (
	// UNKNOWN ...
	UNKNOWN CommandID = iota
	// Redis supported commands

	// GET https://redis.io/commands/get
	GET

	// SET https://redis.io/commands/set
	SET

	// DEL https://redis.io/commands/del
	DEL
	// INFO https://redis.io/commands/info
	INFO
	// INTERNAL_DEREGISTER used to indicate the client to log itself out
	INTERNAL_DEREGISTER
	// KILL used to indicate a client to terminate its reading loop
	KILL
	//RESPONSE ...
	RESPONSE
)

// Command is used to send data between clients and server core
type Command struct {
	CMD             CommandID
	ClientID        uint64
	CallbackChannel chan Command
	Arguments       CommandArguments
}

func (cmd Command) String() string {
	return fmt.Sprintf("%s args: %v", cmd.CMD, cmd.Arguments)
}

type CommandArguments interface{}

type RESPONSEArguments struct {
	Response string
}

type SETArguments struct {
	Key   string
	Value string
}

type GETArguments struct {
	Key string
}

type DELArguments struct {
	Keys []string
}
