package main

import (
	"fmt"
	"go-storj-solution/pkg/proxy"
	"log"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Usage: Service :<port>")
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

	//secrets := NewFixedSecret("abc123")
	secrets := proxy.NewRandomSecrets(6, time.Now().UnixNano())
	service := proxy.New(secrets)

	go service.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %w", err)
		}

		go service.Onboard(conn)
	}
}
