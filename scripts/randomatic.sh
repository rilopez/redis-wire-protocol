#!/usr/bin/env bash
#  
#  Randomatic implements a simple TCP client that sends `n` random readings after login
#  usage
#      scripts/randomatic.sh <imei>
#
#  <imei> is a 15 digit valid imei number used to identify the client
 

imei=$1
scripts/client.sh -type=random  -imei=$imei -readings=0 -reading-rate=500