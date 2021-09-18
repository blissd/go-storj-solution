.PHONY: send receive relay clean

build: send receive relay

send:
	go build go-storj-solution/cmd/send

receive:
	go build go-storj-solution/cmd/receive

relay:
	go build go-storj-solution/cmd/relay

test:
	CGO_ENABLED=0 go test ./...

clean:
	go clean --cache
	rm -f send receive relay