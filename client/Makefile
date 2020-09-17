GOROOT = $(shell go env GOROOT)
GOPATH = $(shell go env GOPATH)

default:
	go mod edit -dropreplace github.com/TheDiscordian/speedychess
	$(MAKE) build

config:
	cp "$(GOROOT)/misc/wasm/wasm_exec.js" .
	
build:
	protoc chesspb.proto --go_out=../ --proto_path=../chesspb/
	python3 ../chesspb/helpergen.py
	python3 ../flags/genconsts.py false
	python3 ../tools/checklicense.py
	go fmt
	GOOS=js GOARCH=wasm go build -o chess.wasm

dev:
	go mod edit -replace github.com/TheDiscordian/speedychess=$(GOPATH)/src/github.com/TheDiscordian/speedychess
	$(MAKE) build	
	
