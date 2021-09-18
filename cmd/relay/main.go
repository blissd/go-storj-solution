package main

import (
	"fmt"
	"github.com/go-kit/log"
	"go-storj-solution/pkg/proxy"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: %v :<port>", os.Args[0])
		os.Exit(1)
	}

	addr := os.Args[1]

	if err := run(addr); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(addr string) error {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}
	defer l.Close()

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	//secrets := NewFixedSecret("abc123")
	secrets := proxy.NewRandomSecrets(6, time.Now().UnixNano())
	service := proxy.New(secrets, logger)

	go service.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %w", err)
		}

		go service.Onboard(conn)
	}
}
