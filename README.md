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


## How to run

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
redis-benchmark -t GET,SET,DEL
```



[Redis Protocol specification]:https://redis.io/topics/protocol

