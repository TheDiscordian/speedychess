# SpeedyChess

A simple chess library, server, client, and AI (eventually).

## Building the server

### Requirements

* go
* make
* protoc (+protoc-gen-go)
* python3

### Build

```
cd server
make
```

Note: Use `make dev` to test local changes.

## Building the client

### Requirements

* go v1.14+
* make
* protoc (+protoc-gen-go)
* python3

### Build

```
cd client
make config
make
```

You only need to run `make config` once to fetch `wasm_exec.js`.

Note: Use `make dev` to test local changes.
