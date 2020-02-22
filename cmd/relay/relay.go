package main

import (
	"github.com/blissd/golang-storj-solution/pkg/session"
	"github.com/blissd/golang-storj-solution/pkg/wire"
	"io"
	"log"
	"net"
)

// a sender or receiver connecting to the relay
type client struct {
	// connection to client
	conn net.Conn
	// send or recv
	side byte
	// secret identifies transfer
	secret string
}

// an ongoing transfer between sender and receiver
type tx struct {
	// unique key for transfer
	secret string
	send   net.Conn
	recv   net.Conn
}

// Copies bytes from sender to receiver
func (t *tx) Run(r *Relay) {
	defer r.close(t.secret)
	s := session.Attach(t.send)

	// Send "receiver is ready" message to sender so that the
	// sender can start sending bytes.
	wire.EncodeByte(session.MsgRecv)
	if err := s.SendRecvReady(); err != nil {
		log.Println("send recv ready failed:", err)
		return
	}

	// Now just pipe from sender to receiver
	// Note that the relay server doesn't care what messages are passed.
	if _, err := io.Copy(t.recv, t.send); err != nil {
		log.Println("tx.Run for:", t.secret, " failed with:", err)
	}
}

// Manages transfers
type Relay struct {
	// Ongoing transfers
	transfers map[string]tx

	// Generator of transfer session secrets
	secrets Secrets

	// Actions to add or remove transfers.
	// `relay` is effectively an actor.
	action chan func()
}

func NewRelay(secrets Secrets) *Relay {
	return &Relay{
		transfers: make(map[string]tx),
		action:    make(chan func()),
		secrets:   secrets,
	}
}

// Process actions to update relay state, such as clients joining and leaving a transfer
// Functions sent to r.action must be non-blocking.
func (r *Relay) Run() {
	for a := range r.action {
		a()
	}
}

// Onboards a new connection for a sender or a receiver.
// For a sender a secret will be generated and sent to the sender.
// For a receiver a secret will be read from the connection.
// A valid client then joins a transfer, either creating it for a sender
// or being associated with an existing transform for a receiver.
func (r *Relay) onboard(conn net.Conn) {
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
		secret = r.secrets.Secret()
		log.Println("generated secret is", secret)
		err = s.SendSecret(secret)
		if err != nil {
			log.Println("send secret in onboard:", err)
			conn.Close()
			return
		}
	case session.MsgRecv:
		log.Println("receiving secret")
		secret, err = s.RecvSecret()
		if err != nil {
			log.Println("recv secret in onboard:", err)
			conn.Close()
			return
		}
	default:
		log.Println("invalid client type in onboard:", clientType)
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

// Joins a new client, either starting a new session for a sender or
// connecting a receiver to an existing session.
// If a receiver has an unknown secret, then their connection is closed.
func (r *Relay) join(c client) {
	r.action <- func() {
		log.Println("join for client:", c)
		switch c.side {
		case session.MsgSend:
			log.Println("onboarding sender for", c.secret)
			if _, ok := r.transfers[c.secret]; ok {
				// should be very unlikely as the relay server generates secrets!
				log.Println("skipping duplicate send client:", c.secret)
				c.conn.Close()
				return
			}
			r.transfers[c.secret] = tx{secret: c.secret, send: c.conn}
		case session.MsgRecv:
			log.Println("onboarding receiver for", c.secret)
			if _, ok := r.transfers[c.secret]; !ok {
				log.Println("skipping recv client because no active tx:", c.secret)
				c.conn.Close()
				return
			}
			t := r.transfers[c.secret]
			t.recv = c.conn

			// Not a map of pointers, so setting t.recv doesn't update r.transfers[c.secret].recv!
			// Should I use a map of pointers to tx?
			r.transfers[c.secret] = t

			// sender and receiver are connect so now start relaying traffic
			log.Println("relay traffic")
			go t.Run(r)
		default:
			log.Println("skipping client because side is invalid:", c.side)
		}
	}
}

// cleans up after ending a transfer for any reason
func (r *Relay) close(secret string) {
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
