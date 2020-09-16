# SpeedyChess Client

This client is a simple WASM client for interacting with the chess server.


## Building the client

### Requirements

* go v1.14+
* make
* protoc (+protoc-gen-go)
* python3

### Build

```
make config
make
```

You only need to run `make config` once to fetch `wasm_exec.js`.

Note: Use `make dev` to test local changes.
