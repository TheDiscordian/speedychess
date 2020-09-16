GOPATH = $(shell go env GOPATH)

default:
	go mod edit -dropreplace github.com/TheDiscordian/speedychess
	$(MAKE) build

build:
	protoc chesspb.proto --go_out=../ --proto_path=../chesspb/
	python3 ../chesspb/helpergen.py
	python3 ../flags/genconsts.py true
	go fmt github.com/TheDiscordian/speedychess/server github.com/TheDiscordian/speedychess/chess github.com/TheDiscordian/speedychess/flags github.com/TheDiscordian/speedychess/chesspb
	go build

dev:
	go mod edit -replace github.com/TheDiscordian/speedychess=$(GOPATH)/src/github.com/TheDiscordian/speedychess
	$(MAKE) build	
