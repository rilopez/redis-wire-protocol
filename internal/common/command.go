package common

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

	// DEREGISTER used to indicate the client to log itself out
	DEREGISTER
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
