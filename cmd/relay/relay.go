package main

import (
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/wire"
	"io"
	"log"
	"net"
)

// a sender or receiver connecting to the relay
type clientInfo struct {
	// connection to client
	conn net.Conn
	// send or recv
	side byte
	// secret identifies transfer
	secret string
}

// an ongoing transfer between sender and receiver
type transfer struct {
	// unique key for transfer
	secret string
	send   net.Conn
	recv   net.Conn
}

// Copies bytes from sender to receiver
func (t *transfer) run(r *relay) {
	defer r.close(t.secret)

	// Send "receiver is ready" message to sender so that the
	// sender can start sending bytes.
	enc := wire.NewEncoder(t.send)
	if err := enc.EncodeByte(client.MsgRecv); err != nil {
		log.Println("send recv ready failed:", err)
		return
	}

	// Now just pipe from sender to receiver
	// Note that the relay server doesn't care what messages are passed.
	if _, err := io.Copy(t.recv, t.send); err != nil {
		log.Println("tx.run for:", t.secret, " failed with:", err)
	}
}

// Manages transfers
type relay struct {
	// Ongoing transfers
	transfers map[string]*transfer

	// Actions to add or remove transfers.
	// `Relay` is effectively an actor.
	action chan func()
}

func newRelay() *relay {
	return &relay{
		transfers: make(map[string]*transfer),
		action:    make(chan func()),
	}
}

// Process actions to update relay state, such as clients joining and leaving a transfer
// Functions sent to r.action must be non-blocking.
func (r *relay) Run() {
	for a := range r.action {
		a()
	}
}

// Joins a new client, either starting a new client for a sender or
// connecting a receiver to an existing client.
// If a receiver has an unknown secret, then their connection is closed.
func (r *relay) join(c clientInfo) {
	r.action <- func() {
		log.Println("join for client:", c.secret)
		switch c.side {
		case client.MsgSend:
			log.Println("joining sender for:", c.secret)
			if _, ok := r.transfers[c.secret]; ok {
				// should be very unlikely as the relay server generates secrets!
				log.Println("skipping duplicate send client:", c.secret)
				_ = c.conn.Close()
				return
			}
			r.transfers[c.secret] = &transfer{secret: c.secret, send: c.conn}
		case client.MsgRecv:
			log.Println("joining receiver for", c.secret)
			if _, ok := r.transfers[c.secret]; !ok {
				log.Println("skipping recv client because no active tx:", c.secret)
				_ = c.conn.Close()
				return
			}
			t := r.transfers[c.secret]
			t.recv = c.conn

			// sender and receiver are connected so now start relaying traffic
			go t.run(r)
		default:
			log.Println("skipping client because side is invalid:", c.side)
			_ = c.conn.Close()
		}
	}
}

// cleans up after ending a transfer for any reason
func (r *relay) close(secret string) {
	r.action <- func() {
		log.Println("closing:", secret)
		defer delete(r.transfers, secret)
		if t, ok := r.transfers[secret]; ok {
			if t.send != nil {
				_ = t.send.Close()
			}
			if t.recv != nil {
				_ = t.recv.Close()
			}
		}
	}
}
