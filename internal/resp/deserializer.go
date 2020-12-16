package resp

import (
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"net/textproto"
	"strconv"
	"strings"
)

/*
DeserializeCMD implements a very simple & limited RESP parser following this assumptions
 - supports only SET,GET, DEL commands
 - supports only double quotes
return an error if  it does not find SET, GET ,DEL at the beginning of the left trimmed string
*/
func DeserializeCMD(reader *textproto.Reader) (common.CommandID, common.CommandArguments, error) {
	arrayHeaderLine, err := reader.ReadLine()
	if err != nil {
		return common.UNKNOWN, nil, err
	}
	if arrayHeaderLine[0] != '*' {
		return common.UNKNOWN, nil, fmt.Errorf("expecting first byte to be *, got %c", arrayHeaderLine[0])
	}
	numItems, err := strconv.Atoi(string(arrayHeaderLine[1:]))
	if err != nil {
		return common.UNKNOWN, nil, fmt.Errorf("invalid array size characters %s", arrayHeaderLine[1:])
	}
	var bulkStringArray []string

	for i := 0; i < numItems; i++ {
		stringHeaderLine, err := reader.ReadLine()
		if err != nil {
			return common.UNKNOWN, nil, err
		}
		if stringHeaderLine[0] != '$' {
			return common.UNKNOWN, nil, fmt.Errorf("expecting first byte to be $, got %c", stringHeaderLine[0])
		}
		numBytes, err := strconv.Atoi(string(stringHeaderLine[1:]))
		str, err := reader.ReadLine()
		if err != nil {
			return common.UNKNOWN, nil, err
		}
		if len(str) != numBytes {
			return common.UNKNOWN, nil, fmt.Errorf("invalid string bytes len %d expecting %d ", len(str), numBytes)
		}
		bulkStringArray = append(bulkStringArray, str)
	}

	if bulkStringArray == nil && len(bulkStringArray) == 0 {
		return common.UNKNOWN, nil, fmt.Errorf("no command read")
	}

	return bulkStringArrayToCommand(bulkStringArray, err)
}

func bulkStringArrayToCommand(bulkStringArray []string, err error) (common.CommandID, common.CommandArguments, error) {
	cmdStr := bulkStringArray[0]
	var cmd common.CommandID
	var cmdArgs common.CommandArguments

	args := bulkStringArray[1:]
	switch strings.ToUpper(cmdStr) {
	case "GET":
		cmd = common.GET
		cmdArgs, err = parseGETArguments(args)
	case "SET":
		cmd = common.SET
		cmdArgs, err = parseSETArguments(args)
	case "DEL":
		cmd = common.DEL
		cmdArgs, err = parseDELArguments(args)
	case "INFO":
		cmd = common.INFO
	case "CLIENT":
		cmd = common.CLIENT
		cmdArgs, err = parseCLIENTArguments(args)
	default:
		return common.UNKNOWN, bulkStringArray, nil
	}

	return cmd, cmdArgs, err
}

func parseSETArguments(args []string) (cmdArgs common.CommandArguments, err error) {

	if len(args) < 2 {
		return nil, fmt.Errorf("invalid number of args for SET command : %v", args)
	}

	//TODO  parse  options SET  EX seconds -- Set the specified expire time, in seconds.
	//TODO  parse  options SET	PX milliseconds -- Set the specified expire time, in milliseconds.
	//TODO  parse  options SET	NX -- Only set the key if it does not already exist.
	//TODO  parse  options SET	XX -- Only set the key if it already exist.
	//TODO  parse  options SET	KEEPTTL -- Retain the time to live associated with the key.
	//	GET -- Return the old value stored at key, or nil when key did not exist.

	options := args[2:]
	optionGET := false

	for _, flag := range options {
		switch strings.ToUpper(flag) {
		case "GET":
			optionGET = true
		}

	}

	return common.SETArguments{Key: args[0], Value: args[1], OptionGET: optionGET}, nil

}

func parseGETArguments(args []string) (cmdArgs common.CommandArguments, err error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of args for GET command : %v", args)
	}
	return common.GETArguments{Key: args[0]}, nil
}

func parseCLIENTArguments(args []string) (cmdArgs common.CommandArguments, err error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of args for CLIENT command : %v", args)
	}
	subCMD := common.ClientSubcommand(strings.ToUpper(args[0]))
	if err := subCMD.IsValid(); err != nil {
		return nil, err
	}
	return common.CLIENTArguments{Subcommand: subCMD}, nil
}

func parseDELArguments(args []string) (cmdArgs common.CommandArguments, err error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("invalid number of args for DEL command: %v", args)
	}

	return common.DELArguments{Keys: args}, nil
}
