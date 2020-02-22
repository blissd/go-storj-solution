package main

import (
	"log"
	"math/rand"
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

	rand.Seed(time.Now().UnixNano())

	r := &relay{
		transfers: make(map[string]tx),
		action:    make(chan func()),
		secrets:   NewRandomSecrets(6),
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
