package main

import (
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

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}
	defer l.Close()

	//secrets := NewFixedSecret("abc123")
	secrets := proxy.NewRandomSecrets(6, time.Now().UnixNano())
	service := proxy.New(secrets)

	go service.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go service.Onboard(conn)
	}
}
