# Simple in-memory cache

My implementation of a Redis-like in-memory cache server.
Client is out of scope due to time constraints, but any redis client can be used.

## Overview
Basically, a subset of Redis functionality. API-wise.    
Supports values of types `string`, `list`, `hash`.    
Supports Per-key TTL.   
Supports basic set of CRUD commands for each type.   

## Demo

Running server 
```
./memcache
```

Interaction via redis-cli
```
./redis-cli -p 9876
127.0.0.1:9876> SET X 42
OK
127.0.0.1:9876> GET X
"42"
127.0.0.1:9876> EXPIRE X 30
(integer) 1
127.0.0.1:9876> TTL X
(integer) 23
127.0.0.1:9876> TTL X
(integer) 8
127.0.0.1:9876> TTL X
(integer) -2
127.0.0.1:9876> HMSET dict a 1 b 2 c 3
OK
127.0.0.1:9876> HGETALL dict
1) "b"
2) "2"
3) "c"
4) "3"
5) "a"
6) "1"
127.0.0.1:9876> LPUSH list 1 2 3 4 5 6
(integer) 6
127.0.0.1:9876> LRANGE list 0 -1
1) "6"
2) "5"
3) "4"
4) "3"
5) "2"
6) "1"
127.0.0.1:9876> RPOP list
"1"
127.0.0.1:9876> RPOP list
"2"
127.0.0.1:9876> RPOP list
"3"
127.0.0.1:9876> RPOP no-list
(nil)
127.0.0.1:9876> RPOP list another args
(error) ARGS_NUM RPOP requires 1 args exactly, got 3
```

## Protocol
[RESP](https://redis.io/topics/protocol) was chosen as a protocol for communicating with clients. 
It's a simple, solid, low-overhead protocol. Another benefit is that `redis-cli` or any other Redis client
can be used to communicate with the server.
It will fail in some error cases, for example because COMMAND command is not implemented.

## Configuration
Server is configured via environment variables.   
* `SERVER_PORT` sets server port. Default is `9876`.
Example: `SERVER_PORT=7000`
* `LOG_LEVEL` sets log level. Default is `info` Possible values are [logrus](https://github.com/sirupsen/logrus) log levels. 
Example: `LOG_LEVEL=warn`
* `LOG_FORMAT` sets log format. Default is `text` Possible values are `json`, `text`.
Example: `LOG_FORMAT=json`

# Building and running locally
`build-local` directory contains simple scripts to 
build and run the server locally with [goreleaser](https://github.com/goreleaser/goreleaser).

`./build-local/build.sh` builds the project.

`./build-local/run.sh` runs the project in default configuration.

# Releases
Releases are available in [github releases](https://github.com/AntonRutkevich/simple-memcache/releases).

Just run the binary.

## Possible improvements
### Features
 * Active cleanup of expired keys. 
 Strategy similar to [Redis strategy](https://redis.io/commands/expire#how-redis-expires-keys) can be used.
 * `KEYS pattern` command. Performance must be considered carefully when implementing.
 * Disk storage.  

### Performance
Due to time constraints it sucks currently :) So there's a lot to improve.

Always measure!
Check which type of interaction you are optimizing!
Keep in mind, that server is accessed over network,
which might be orders of magnitude slower than actual request processing for simple requests.
Mem-profile methods to see if any memory allocations can be avoided.

* Concurrent reads/writes of single key.
  * Consider using sync.RWMutex.
   Current issue with it is that expired key can be deleted on access.
* Concurrent reads/writes of different keys.
  * Create map of per-key locks and avoid locking entire map.
  * With per-key locks, consider limiting the number of goroutines to GOMAXPROCS or so.
   The bottleneck is most probably CPU-bound.
* `string`, `list`, `map` specific operations
  * Marshaling empty/short lists.
  * Using [linked-list](https://golang.org/pkg/container/list/) instead of slices for lists,
 as growing and copying slices is expensive.
 Memory allocated by list entries is not cleaned up until the list key is deleted.

### Architecture 
* Wrap logrus.Logger as implementation detail
* Move resp.types to handlers level, leave the marshaling logic next to server

### Infrastructure
* Add panic handling
* Add graceful server shutdown
* Refine handling of connection read/write errors: which are fine skip and close connection?

## Commands
The list of commands is a subset of Redis commands, 
with an attempt to follow the redis commands behavior as close as possible. 
See [RESP](https://redis.io/topics/protocol) for the definition of types.

Errors returned from every command: 
* Number of arguments is different from the expected by the command.
* The request is malformed in some way.

Errors returned from every command except [Keys API](#keys-api):
* Type of the `value` stored in `key` is different from command type.

Some commands return errors specific to their arguments.

### String API
#### GET key
Gets the string value of the `key`.

Bulk String reply: value of `key`, or `(nil)` if `key` does not exist.
 
#### SET key value
Sets the `key` to hold the string `value`.

Simple String reply: "OK".

### List API
#### LPOP key
Pops the first items from `key`. 

Bulk String reply: string value of the first item at `key`, or `(nil)` if list is empty or `key` does not exist.
 
#### RPOP key
Pops the last items from `key`. 

Bulk String reply: string value of the last item in `key`, or `(nil)` if list is empty or `key` does not exist.

#### LPUSH key value [value ...]
Pushes one to many items to the beginning of the list at `key`. 
Items are pushed left-to-right, that is `LPUSH mylist 1 2 3` will result in `[3, 2, 1]` list.
Creates new list if does not exist yet.

Int reply: list size after push operation.

#### RPUSH key value [value ...]
Pushes one to many items to the end of the list at `key`. 
Items are pushed left-to-right, that is `RPUSH mylist 1 2 3` will result in `[1, 2, 3]` list.
Creates new list if does not exist yet.

Int reply: list size after push operation.

#### LRANGE key start stop
Returns elements from list `key`.
`start` and `stop` are 0-based and are both, inclusive, that is 
`LRANGE mylist 0 0` returns list of `[0th]` element, and 
`LRANGE mylist 0 1` returns list of `[0th, 1st]` elements.

If `start` index is after the list end, empty list is returned.

If `stop` index is after the list end, it is treated as the last element of the list.

If `start` or `stop` indices are negative, they are treated as counting backwards from the list end, that is `LRANGE mylist 0 -1` returns all list elements. 

Array reply: list of elements in the specified range.

### Hash API
#### HGET key field
Returns value stored in `key` map in the field `field`. 

Bulk String reply: value stored in map, or '(nil)' if `field` is not set or `key` does not exist.   

#### HMGET key field [field ...]
Returns values stored in `key` map in the fields listed.

For every non-existing `field` a `(nil)` value is returned. 
Running the command against a non-existing `key` results in a list of `(nil)` values. 

Array reply: values stored in map.   

#### HGETALL key
Returns all values stored in `key` map.

Array reply: all values stored in map.

#### HSET key field value
Sets the `field` in the `key` map to hold the `value`.
Creates new hash if does not exist yet.

Int reply: `1` if `field` is a new field, `0` if `field` was updated.

#### HMSET key field value [field value ...]
Sets the every `field` in the `key` map to hold the corresponding `value`.
Creates new hash if does not exist yet.

Simple String reply: "OK".

#### HDEL key field [field ...]
Clears all `field`s in the `key` map.

Int reply: number of fields removed.

### Keys API 
#### DEL key
Deletes the `key`. No-op if `key` does not exist.

Int reply: number of keys removed.

#### EXPIRE key seconds
Sets the expiration to the `key`, in seconds.

If key already had an expiration, it is replaced with the new value. 
Expiration can be removed by a command that overwrites the key itself, like `DEL`, `SET`. 
Commands that alter the key content (like `LPOP`, `HMSET`, etc) do not affect the expiration.

Expiration is timestamp-based, so it is sensitive to system time changes.
Expired keys are cleared from memory in a passive way currently, that is on-access.

Int reply: `1` if timeout was set, `0` if `key` does not exist.

#### TTL key
Gets the remaining time to live of the `key`, in seconds.

Int reply: ttl in seconds, or `-1` if the key has no ttl set, `-2` if the key does not exist. 
