#!/usr/bin/env bash
#
#  creates an automated  thermomatic client 
#
#   usage
#      scripts/client.sh  <options>
#
#   the following options are available:
#      -imei string
#             Device IMEI number
#      -readings uint
#             Number of automatic readings the automated client will send,  if equals 0  it sends an infite number of readings (default 5)
#      -reading-rate uint
#             Number of milliseconds between each reading (default 25)
#      -server-address string
#             Address (host:port) of the Thermomatic server (default "localhost:1337")
#      -type string
#             Automated simulated client type, it could be random, slow, too slow  (default "random")
#  
#   Random IMEIs (https://dyrk.org/tools/imei/): 
#         999755843373863
#         448324242329542    
#         304412928289834
set -euo pipefail

go run main.go client "$@" > client-output.txt 2>client.log


