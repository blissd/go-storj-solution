package main

import (
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/proxy"
	"go-storj-solution/pkg/wire"
	"io"
	"log"
	"net"
)

// a sender or receiver connecting to the Relay
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
	// secret is a unique key for transfer
	secret string

	// send is connection from the sender
	send net.Conn

	// recv is the connection to the receiver
	recv net.Conn
}

// Copies bytes from sender to receiver
func (t *transfer) run(r *Relay) {
	defer r.close(t.secret)

	// Send "receiver is ready" message to sender so that the
	// sender can start sending bytes.
	enc := wire.NewEncoder(t.send)
	if err := enc.EncodeByte(client.MsgRecv); err != nil {
		log.Println("send recv ready failed:", err)
		return
	}

	// Now just pipe from sender to receiver
	// Note that the Relay server doesn't care what messages are passed.
	if _, err := io.Copy(t.recv, t.send); err != nil {
		log.Println("tx.run for:", t.secret, " failed with:", err)
	}
}

// Manages transfers between senders and receivers.
type Relay struct {
	// secrets source
	secrets proxy.Secrets

	// transfers that are in progress
	transfers map[string]*transfer

	// action to add or remove transfers.
	// `Relay` is effectively an actor.
	action chan func()
}

func New(secrets proxy.Secrets) *Relay {
	return &Relay{
		secrets:   secrets,
		transfers: make(map[string]*transfer),
		action:    make(chan func()),
	}
}

// Run processes actions to update relay proxy state, such as clients joining and leaving a transfer.
// Functions sent to r.action must be non-blocking.
// Expected to be called from a go routine.
func (r *Relay) Run() {
	for a := range r.action {
		a()
	}
}

// Onboard adds a sender or receiver to the Relay proxy.
// For a sender a Secret will be generated and sent to the sender.
// For a receiver a Secret will be read from the connection.
// A valid client then joins a transfer, either creating it for a sender
// or being associated with an existing transform for a receiver.
// Expected to be called from a go routine.
func (r *Relay) Onboard(conn net.Conn) {
	dec := wire.NewDecoder(conn)
	clientType, err := dec.DecodeByte()
	if err != nil {
		log.Println("failed reading first byte:", err)
		_ = conn.Close()
		return
	}

	log.Println("onboarding for", clientType)

	var secret string

	switch clientType {
	case client.MsgSend:
		// Onboarding a sender so generate and send secret for this transfer
		log.Println("sending Secret")
		secret = r.secrets.Secret()
		log.Println("generated Secret is", secret)
		err = wire.NewEncoder(conn).EncodeString(secret)
		if err != nil {
			log.Println("send Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	case client.MsgRecv:
		// Onboarding a receiver so read the secret for the transfer
		log.Println("receiving Secret")
		secret, err = dec.DecodeString()
		if err != nil {
			log.Println("recv Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	default:
		log.Println("invalid client type in onboard:", clientType)
		_ = conn.Close()
		return
	}

	c := clientInfo{
		conn:   conn,
		side:   clientType,
		secret: secret,
	}
	r.join(c)
}

// Joins a new client, either starting a new client for a sender or
// connecting a receiver to an existing client.
// If a receiver has an unknown Secret, then their connection is closed.
func (r *Relay) join(c clientInfo) {
	r.action <- func() {
		log.Println("join for client:", c.secret)
		switch c.side {
		case client.MsgSend:
			log.Println("joining sender for:", c.secret)
			if _, ok := r.transfers[c.secret]; ok {
				// should be very unlikely as the Relay server generates Secrets!
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
func (r *Relay) close(secret string) {
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
