#!/usr/bin/env bash
#
#  start a tcp server for redis protocol
#
#   usage
#      scripts/serve.sh  <options>
#
#   the following options are available:
# 
#        -max-clients uint
#                maximum number of active client connections  (default 100_000)
#        -port uint
#                port number to listen for TCP connections of clients implementing  (default 6379)#
set -euo pipefail

LOG_FILE="server.log"
SERVER_OUTPUT_FILE="server-output.txt"

echo "starting REDIS simple clone server  & redirecting std & stderr  to $LOG_FILE & $SERVER_OUTPUT_FILE"
echo "Press CTRL + C to quit"
go run cmd/server.go "$@"  > server-output.txt 2>server.log


