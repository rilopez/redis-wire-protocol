#!/usr/bin/env bash
#
#  spin up a tcp server for redis protocol
#
#   usage
#      scripts/serve.sh  <options>
#
#   the following options are available:
# 
#        -max-clients uint
#                maximun number of active client connections  using the thermomatic protocol (default 1000)
#        -port uint
#                port number to listen for TCP connections of clients implementing the  thermomatic protocol (default 1337)# 
set -euo pipefail

go run cmd/server.go "$@"  > server-output.txt 2>server.log


