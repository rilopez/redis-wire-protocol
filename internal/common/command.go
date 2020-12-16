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
	// CLIENT
	//  https://redis.io/commands/client-list
	//  https://redis.io/commands/client-info
	//  https://redis.io/commands/client-ind
	CLIENT

	// INTERNAL_DEREGISTER used to indicate the client to log itself out
	INTERNAL_DEREGISTER
)

type ClientSubcommand string

const (
	ClientSubcommandID   ClientSubcommand = "ID"
	ClientSubcommandINFO ClientSubcommand = "INFO"
	ClientSubcommandLIST ClientSubcommand = "LIST"
)

func (sub ClientSubcommand) IsValid() error {
	switch sub {
	case ClientSubcommandID, ClientSubcommandINFO, ClientSubcommandLIST:
		return nil
	}
	return fmt.Errorf("%s is an invalid client subcommand", sub)

}

// Command is used to send data between clients and server core
type Command struct {
	CMD       CommandID
	ClientID  uint
	Arguments CommandArguments
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

type CLIENTArguments struct {
	Subcommand ClientSubcommand
}

type DELArguments struct {
	Keys []string
}
