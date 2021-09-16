.PHONY: send receive relay

build: send receive relay

send:
	go build go-storj-solution/cmd/send

receive:
	go build go-storj-solution/cmd/receive

relay:
	go build go-storj-solution/cmd/relay

test:
	go test .