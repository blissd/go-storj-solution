package main

import (
	"github.com/blissd/golang-storj-solution/pkg/session"
	"io"
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

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}
	defer l.Close()

	onboarding := make(chan client)
	go onboard(onboarding)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go handle(conn, onboarding)
	}
}

func onboard(clients chan client) {
	transfers := make(map[string]transfer)
	log.Println("go onboard")
	for c := range clients {
		log.Println("onboard for client:", c)
		switch c.side {
		case session.MsgSend:
			log.Println("onboarding sender for", c.secret)
			if _, ok := transfers[c.secret]; ok {
				log.Println("skipping duplicate send client:", c.secret)
				continue
			}
			transfers[c.secret] = transfer{send: c.conn}
		case session.MsgRecv:
			log.Println("onboarding receiver for", c.secret)
			if _, ok := transfers[c.secret]; !ok {
				log.Println("skipping recv client because no active transfer:", c.secret)
				continue
			}
			t := transfers[c.secret]
			t.recv = c.conn

			// sender and receiver are connect so now start relaying traffic
			log.Println("relay traffic")
			go t.Run()
		default:
			log.Println("skipping client because side is invalid:", c.side)
		}
	}
}

func (t *transfer) Run() {
	// first send byte to sender to indicate receiver is ready
	s := session.Attach(t.send)
	s.SendRecvReady()

	// Now just pipe from sender to receiver
	io.Copy(t.recv, t.send)
}

func handle(conn net.Conn, onboarding chan client) {
	s := session.Attach(conn)
	clientType, err := s.FirstByte()

	if err != nil {
		log.Println("failed reading first byte:", err)
		conn.Close()
		return
	}

	log.Println("onboarding for", clientType)

	var secret string

	switch clientType {
	case session.MsgSend:
		log.Println("sending secret")
		secret = "123abc"
		err = s.SendSecret(secret)
		if err != nil {
			log.Println("onboarding:", err)
			conn.Close()
			return
		}
	case session.MsgRecv:
		log.Println("receiving secret")
		secret, err = s.RecvSecret()
		if err != nil {
			log.Println("onboarding:", err)
			conn.Close()
			return
		}
	default:
		log.Println("must be send/recv:", clientType)
		conn.Close()
		return
	}

	log.Println("client is onboarded, sending message")
	c := client{
		conn:   conn,
		side:   clientType,
		secret: secret,
	}
	onboarding <- c
}
