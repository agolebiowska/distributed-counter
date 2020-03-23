# Distributed counter

[![Build Status](https://travis-ci.org/agolebiowska/distributed-counter.svg?branch=master)](https://travis-ci.org/agolebiowska/distributed-counter)  

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

Show logs from containers

```shell
$ make log 
```  

Run system simulation

```shell
$ make simulate 
``` 

You can copy example config file to .env and change values in it
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

### Overview 

The requirements are focused mostly on data consistency and handling child node failures.  
I decided to choose one matching scenario described in the given article (http://book.mixu.net/distsys/abstractions.html).
<img align="right" alt="gopher" align="center" src="http://book.mixu.net/distsys/images/CAP.png" width="20%">

- `AP` design was not the case because of the consistency requirement.
- `CP` cannot be fully achieved with only one callable main coordinator. Mechanism to choosing one of the counters as master after coordinator failure would be something worth consideration.
- `CA` and Two-Phase commit protocol approach is something I choose for this system. 

**Disadvantages**
- Request is synchronous (blocking).
- Possibility of deadlock between transactions.
- It is not partition tolerant.

The above drawback may lead to system performance bottleneck too. It is sacrifice of efficiency.
Using this approach it was also possible to achieve three of the ACID principles: `atomicity`, `consistency` and `read-write isolation`. 

### Flow

#### Add counter
- When a new counter instance is added it sends request to coordinator to obtain data from other counters.
- If a counter goes down and recover it will get data the same way. This ensures data consistency. 
<img align="center" alt="gopher" align="center" src="https://raw.githubusercontent.com/agolebiowska/distributed-counter/master/.img/3.png" width="50%">

#### Add items
- Coordinator sends unique message to `all` counters.
- Counters must make a decision if they can save items.
- If one or more counters refuse `all` will receive request to forget about previous message.
<img align="center" alt="gopher" align="center" src="https://raw.githubusercontent.com/agolebiowska/distributed-counter/master/.img/1.png" width="50%"> 
<img align="center" alt="gopher" align="center" src="https://raw.githubusercontent.com/agolebiowska/distributed-counter/master/.img/2.png" width="50%">
 
#### Get count
- To get count coordinator sends request to one random counter.
- Docker handles requests balancing in that case. It will not call dead nodes.
<img align="center" alt="gopher" align="center" src="https://raw.githubusercontent.com/agolebiowska/distributed-counter/master/.img/4.png" width="50%">

#### Health checks
- Coordinator performs health checks every 10 seconds. `Todo: make health check interval configurable.`
- If a counter not respond or respond with an error it is marked as dead and it is not query-able.
- After 4 more unsuccessful responses coordinator removes that counter. `Todo: make number of recovery tries configurable.`
- Docker performs coordinator health checks every 30 seconds.

### Possible improvements
- RPC or sockets could be used instead of HTTP for communication between coordinator and counters.
- Different distributed algorithm like paxos or raft could be implemented for our system to be partition tolerant.