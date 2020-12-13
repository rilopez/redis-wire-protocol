package resp

import (
	"encoding/csv"
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"strings"
)

/*
Deserialize implements a very simple & limited RESP parser following this assumptions
 - supports only SET,GET, DEL commands
 - supports only double quotes
return an error if  it does not find SET, GET ,DEL at the beginning of the left trimmed string
*/
func Deserialize(serializedCMD string) (common.CommandID, common.CommandArguments, error) {
	trimmed := strings.TrimLeft(serializedCMD, " ")

	cmdStr := trimmed[:3]
	cmdArgStr := trimmed[4:]
	cmdArgStr = strings.TrimLeft(cmdArgStr, " ")
	cmdArgStr = strings.TrimRight(cmdArgStr, " ")
	var cmd common.CommandID
	var cmdArgs common.CommandArguments
	var err error
	switch strings.ToUpper(cmdStr) {
	case "GET":
		cmd = common.GET
		cmdArgs, err = parseGETArguments(cmdArgStr)
	case "SET":
		cmd = common.SET
		cmdArgs, err = parseSETArguments(cmdArgStr)
	case "DEL":
		cmd = common.DEL
		cmdArgs, err = parseDELArguments(cmdArgStr)
	default:
		return common.UNKNOWN, nil, fmt.Errorf("unsupported command %s", cmdStr)
	}

	return cmd, cmdArgs, err
}

func parseSETArguments(str string) (cmdArgs common.CommandArguments, err error) {
	args, err := splitArgs(str)
	if err != nil {
		return nil, err
	}

	if len(args) < 2 {
		return nil, fmt.Errorf("invalid number of args for SET command : %v", args)
	}

	//TODO  parse  options SET key value [EX seconds|PX milliseconds|KEEPTTL] [NX|XX] [GET]
	//  EX seconds -- Set the specified expire time, in seconds.
	//	PX milliseconds -- Set the specified expire time, in milliseconds.
	//	NX -- Only set the key if it does not already exist.
	//	XX -- Only set the key if it already exist.
	//	KEEPTTL -- Retain the time to live associated with the key.
	//	GET -- Return the old value stored at key, or nil when key did not exist.

	return common.SETArguments{Key: args[0], Value: args[1]}, nil

}

func parseGETArguments(str string) (cmdArgs common.CommandArguments, err error) {
	trimmed := strings.TrimLeft(str, " ")
	args, err := splitArgs(trimmed)
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of args for GET command : %s", str)
	}

	return common.GETArguments{Key: args[0]}, nil
}

func parseDELArguments(str string) (cmdArgs common.CommandArguments, err error) {
	trimmed := strings.TrimLeft(str, " ")
	args, err := splitArgs(trimmed)
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("invalid number of args for DEL command: %s", str)
	}

	return common.DELArguments{Keys: args}, nil
}

func splitArgs(str string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(str))
	r.Comma = ' '
	return r.Read()
}
