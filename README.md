# Distributed counter

This is a distributed service, consisting of multiple isolated processes
which can count the number of items, grouped by tenants that are delivered
through an HTTP restful interface.

| Resource                 | Description|
|:-------------------------|:-----------|
| `POST /items` | add new items|
| `GET /items/tenantID/count` | return number of items for given tenant| 


## Setup

Build & run

```shell
$ make up
```

or if you want to debug

```shell
$ make dev
```

Run tests

```shell
$ make test 
```  

Show logs form containers

```shell
$ make log 
```  

Run simulation

```shell
$ make simulate 
``` 

You can copy example config file to .env and change config
```shell
$ cp .env.dist .env
``` 
```.env
# Http port on which coordinator server will listen
HTTP_PORT=8080

# Debugger port
DEBUG_PORT=40000
```

## Design

### System consists of
- Coordinator which provides an RESTful API.
- `N` number of counters which can be called only from coordinator itself.

### Flow

Counters must make a decision if they can save the items. If one or more counters refuse all counters will receive request to forget about previous message.  

### Possible improvements
- RPC or sockets could be used instead of HTTP for communication between coordinator and counters.
- Different distributed algorithm like paxos or raft could be implemented for our system to be fully partition tolerant.