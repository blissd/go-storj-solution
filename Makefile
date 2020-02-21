.PHONY: send receive relay

all: send receive relay

send:
	go build github.com/blissd/golang-storj-solution/cmd/send

receive:
	go build github.com/blissd/golang-storj-solution/cmd/receive

relay:
	go build github.com/blissd/golang-storj-solution/cmd/relay