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
# SET x 1
# +OK
# GET x
# +1
# DEL x
# :1
```

**Test with redis benchmark**

```bash
redis-benchmark -t set,get -n 1000000 -q                                            
# SET: 119531.44 requests per second
# GET: 135208.22 requests per second
#
# REDIS real server is at least twice as fast, here some numbers from may local
# SET: 219731.92 requests per second
# GET: 214178.62 requests per second
```

Of course, you can test running `go test ./...` , take a look to `internal/server/server_intergration_test.go` for E2E
tests.

## Assumptions & Known issues

1. Assuming that the challenge requirement of using only the stdlib applies only to production code, I took the liberty
   to use these two dependencies for the testing code:
   - [redis go client](https://github.com/go-redis/redis)
   - [uber's go routine leak testing library](https://github.com/uber-go/goleak)

[Redis Protocol specification]:https://redis.io/topics/protocol

