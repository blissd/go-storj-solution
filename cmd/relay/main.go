package main

import (
	"fmt"
	"github.com/blissd/golang-storj-solution/pkg/session"
	"log"
	"net"
	"os"
)

type transfer struct {
	send net.Conn
	recv net.Conn
}

type client struct {
	conn   net.Conn
	side   byte
	secret string
}

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Usage: relay :<port>")
	}

	addr := os.Args[1]
	fmt.Println("Port is", addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go handle(conn)
	}
}

func onboard(clients chan<- client) {
	transfers := make(map[string]transfer)
	println(transfers)
	for c := range clients {
		switch c.side {
		case session.MsgSend:
			if _, ok := transfers[c.secret]; ok {
				log.Println("skipping duplicate send client:", c.secret)
				continue
			}
			transfers[c.secret] = transfer{send: c.conn}
		case session.MsgRecv:
			if _, ok := transfers[c.secret]; !ok {
				log.Println("skipping recv client because no active transfer:", c.secret)
				continue
			}
			t := transfers[c.secret]
			t.recv = c.conn

			// sender and receiver are connect so now start relaying traffic
			log.Println("TODO - relay traffic")
		default:
			log.Println("skipping client because side is invalid:", c.side)
		}
	}
}

// inspects the first two bytes of a connection to determine if
// it is for a sender or receiver.
func route(conn net.Conn) {
	s := session.Attach(conn)
	b, err := s.FirstByte()
	if err != nil {
		log.Println("failed getting first byte:", err)
		return
	}

	switch b {
	case session.MsgSend:
		log.Println("sender connected")
	case session.MsgRecv:
		log.Println("receiver connected")
	default:
		log.Println("invalid start byte:", b)
		return
	}

}

func handle(conn net.Conn) {

}
