package resp

import (
	"fmt"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"strings"
)

/*
Deserialize implements a very simple & limited RESP parser following this assumptions
 - supports only SET,GET, DEL commands
 - return an error if  it does not find SET, GET ,DEL at the beginning of the left trimmed string
*/
func Deserialize(serializedCMD string) (common.CommandID, common.CommandArguments, error) {
	trimmed := strings.TrimLeft(serializedCMD, " ")

	cmdStr := trimmed[:3]
	cmdArgStr := trimmed[4:]
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
	return nil, nil
}

func parseGETArguments(str string) (cmdArgs common.CommandArguments, err error) {
	return nil, nil
}

func parseDELArguments(str string) (cmdArgs common.CommandArguments, err error) {
	return nil, nil
}
