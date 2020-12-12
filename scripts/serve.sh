#!/usr/bin/env bash
#
#  spin up a tcp/http server for thermomatic protocol 
#
#   usage
#      scripts/serve.sh  <options>
#
#   the following options are available:
# 
#        -http-port uint
#                port number to listen for HTTP connections used mainly for healthchecks (default 80)
#        -max-clients uint
#                maximun number of active client connections  using the thermomatic protocol (default 1000)
#        -port uint
#                port number to listen for TCP connections of clients implementing the  thermomatic protocol (default 1337)# 
set -euo pipefail

go run main.go server "$@"  > server-output.txt 2>server.log


