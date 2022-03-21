# SpeedyChess

A simple chess library, server, client, and AI (eventually).

NOTICE: This was a fun little project Discordian & Runite made a while back, it will likely never be complete. Enjoy!

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
