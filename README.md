# redi-xmas


## Preface


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

 

OS resources do not leak under sustained use (specifically memory and file descriptors)
Unsupported commands receive a response indicating they are not supported
Multiple clients can simultaneously access the application, potentially the same key
Resources

## How to run
```
go run main.go
```

Test with telnet
```
telnet localhost 6379
SET x 1
+OK
GET x
+1
DEL x
:1
```

Test with redis bechmark

```bash
redis-benchmark -t GET,SET,DEL
```

## Protocol specification

[Redis Protocol specification](https://redis.io/topics/protocol)

## Tasks 

- [x] Read/digest/play  [Redis Protocol specification](https://redis.io/topics/protocol)
- [ ] Create basic server skeleton to receive TCP connections with one byte package  
- [ ] redis protocol deserialization
- [ ] redis protocol serialization
- [ ] implement `SET` command
- [ ] implement `GET` command
- [ ] implement `DEL` command
