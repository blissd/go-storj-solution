package main

import (
	"log"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Usage: relay :<port>")
	}

	addr := os.Args[1]

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}
	defer l.Close()

	r := &relay{
		transfers: make(map[string]tx),
		action:    make(chan func()),
		secrets:   NewRandomSecrets(6, time.Now().UnixNano()),
	}
	go r.run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go r.onboard(conn)
	}
}
