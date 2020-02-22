package main

import (
	"github.com/blissd/golang-storj-solution/pkg/session"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
)

const secretLength = 6

// a sender or receiver connecting to the relay
type client struct {
	conn   net.Conn
	side   byte
	secret string
}

// an ongoing transfer between sender and receiver
type tx struct {
	secret string
	send   net.Conn
	recv   net.Conn
}

// Manages transfers
type relay struct {
	// ongoing transfers
	transfers map[string]tx

	// Actions to add or remove transfers.
	// `relay` is effectively an actor.
	action chan func()
}

// Process actions
func (r *relay) run() {
	for a := range r.action {
		a()
	}
}

// cleans up after ending a transfer for any reason
func (r *relay) close(secret string) {
	r.action <- func() {
		log.Println("closing:", secret)
		defer delete(r.transfers, secret)
		if t, ok := r.transfers[secret]; ok {
			if t.send != nil {
				log.Println("closing send:", secret)
				t.send.Close()
			}
			if t.recv != nil {
				log.Println("closing recv:", secret)
				t.recv.Close()
			}
		}
	}
}

// Joins a new client, either starting a new session for a sender or
// connecting a receiver to an existing session.
func (r *relay) join(c client) {
	r.action <- func() {
		log.Println("join for client:", c)
		switch c.side {
		case session.MsgSend:
			log.Println("onboarding sender for", c.secret)
			if _, ok := r.transfers[c.secret]; ok {
				log.Println("skipping duplicate send client:", c.secret)
				return
			}
			r.transfers[c.secret] = tx{secret: c.secret, send: c.conn}
		case session.MsgRecv:
			log.Println("onboarding receiver for", c.secret)
			if _, ok := r.transfers[c.secret]; !ok {
				log.Println("skipping recv client because no active tx:", c.secret)
				return
			}
			t := r.transfers[c.secret]
			t.recv = c.conn
			r.transfers[c.secret] = t

			// sender and receiver are connect so now start relaying traffic
			log.Println("relay traffic")
			go t.Run(r)
		default:
			log.Println("skipping client because side is invalid:", c.side)
		}
	}
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

	r := &relay{
		transfers: make(map[string]tx),
		action:    make(chan func()),
	}
	go r.run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go onboard(conn, r)
	}
}

func (t *tx) Run(r *relay) {
	// first send byte to sender to indicate receiver is ready
	defer r.close(t.secret)
	s := session.Attach(t.send)
	if err := s.SendRecvReady(); err != nil {
		log.Println("send recv ready failed:", err)
	}

	// Now just pipe from sender to receiver
	if _, err := io.Copy(t.recv, t.send); err != nil {
		log.Println("copy failed for:", t.secret, "with:", err)
	}
}

func onboard(conn net.Conn, r *relay) {
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
		//secret = "abc123"
		secret = generateSecret(secretLength)
		log.Println("generated secret is", secret)
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

	c := client{
		conn:   conn,
		side:   clientType,
		secret: secret,
	}
	r.join(c)
}

var letters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func generateSecret(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
