# Distributed counter

| Resource                 | Description|
|:-------------------------|:-----------|
| `POST /items` | add new items|
| `GET /items/tenantID/count` | return number of items for given tenant| 


## Setup

Copy example config file to .env and set the values in it.

```shell
$ cp .env.dist .env
```

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