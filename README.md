# redi-xmas


**redi-xmas** is a simple key/value store that support the [Redis Protocol specification]. 
It was created  for the NS1 code challenge   

```
## Task definition

 Using only its stdlib, create an in-memory key-value store that supports Redis wire protocol’s `GET`, `DEL` and `SET`. 
 Both keys and values may be of any primitive type.

## Success Metrics
While running your solution, the following commands respond as such:
$ <start solution>

$ redis-cli set x 1
OK
$ redis-cli get x
“1”
$ redis-cli del x
(integer) 1
$ redis-cli get x
(nil)

 
- OS resources do not leak under sustained use (specifically memory and file descriptors)
- Unsupported commands receive a response indicating they are not supported
- Multiple clients can simultaneously access the application, potentially the same key
Resources

```


## How to run & test

```bash
scripts/serve.sh
```
You can see logs by running `tail -f server.log` on another terminal, or simply run `go run cmd/server.go` to see logs 
and output in stdout


**Test with telnet**
```bash
telnet localhost 6379
SET x 1
+OK
GET x
+1
DEL x
:1
```

**Test with redis bechmark**
```bash
redis-benchmark -t GET,SET

====== SET ======
  100000 requests completed in 0.69 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

98.65% <= 1 milliseconds
98.72% <= 2 milliseconds
98.87% <= 3 milliseconds
99.04% <= 4 milliseconds
99.23% <= 5 milliseconds
99.43% <= 6 milliseconds
99.53% <= 7 milliseconds
99.61% <= 8 milliseconds
99.66% <= 9 milliseconds
99.74% <= 10 milliseconds
99.81% <= 11 milliseconds
99.86% <= 12 milliseconds
99.90% <= 14 milliseconds
100.00% <= 14 milliseconds
145985.41 requests per second

====== GET ======
  100000 requests completed in 0.68 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1

98.82% <= 1 milliseconds
99.15% <= 2 milliseconds
99.35% <= 3 milliseconds
99.40% <= 4 milliseconds
99.52% <= 5 milliseconds
99.65% <= 6 milliseconds
99.70% <= 7 milliseconds
99.79% <= 8 milliseconds
99.85% <= 10 milliseconds
99.85% <= 11 milliseconds
99.90% <= 12 milliseconds
100.00% <= 12 milliseconds
146842.88 requests per second

```








[Redis Protocol specification]:https://redis.io/topics/protocol

