package proxy

import (
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/wire"
	"io"
	"log"
	"net"
)

// Service manages transfers between senders and receivers.
type Service struct {
	// secrets source
	secrets Secrets

	// transfers that are in progress
	transfers map[string]*transfer

	// action to add or remove transfers.
	// `Service` is effectively an actor.
	action chan func()
}

func New(secrets Secrets) *Service {
	return &Service{
		secrets:   secrets,
		transfers: make(map[string]*transfer),
		action:    make(chan func()),
	}
}

// Run processes actions to update relay proxy state, such as clients joining and leaving a transfer.
// Functions sent to r.action must be non-blocking.
// Expected to be called from a go routine.
func (r *Service) Run() {
	for a := range r.action {
		a()
	}
}

// Onboard adds a sender or receiver to the Service proxy.
// For a sender a Secret will be generated and sent to the sender.
// For a receiver a Secret will be read from the connection.
// A valid client then joins a transfer, either creating it for a sender
// or being associated with an existing transform for a receiver.
// Expected to be called from a go routine.
func (r *Service) Onboard(conn net.Conn) {
	dec := wire.NewDecoder(conn)

	var side client.Side
	{
		b, err := dec.DecodeByte()
		if err != nil {
			log.Println("failed reading first byte:", err)
			_ = conn.Close()
			return
		}

		side = client.Side(b)
	}

	log.Println("onboarding for", side)

	var secret string

	switch side {
	case client.MsgSend:
		// Onboarding a sender so generate and send secret for this transfer
		log.Println("sending Secret")
		secret = r.secrets.Secret()
		log.Println("generated Secret is", secret)
		if err := wire.NewEncoder(conn).EncodeString(secret); err != nil {
			log.Println("send Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	case client.MsgRecv:
		// Onboarding a receiver so read the secret for the transfer
		log.Println("receiving Secret")
		var err error
		if secret, err = dec.DecodeString(); err != nil {
			log.Println("recv Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	default:
		log.Println("invalid client type in onboard:", side)
		_ = conn.Close()
		return
	}

	ts := transferSide{
		conn:   conn,
		side:   side,
		secret: secret,
	}
	r.join(ts)
}

// Joins a new side of the transfer, either starting a new client for a sender or
// connecting a receiver to an existing client.
// If a receiver has an unknown Secret, then their connection is closed.
func (r *Service) join(ts transferSide) {
	r.action <- func() {
		log.Println("join for client:", ts.secret)
		switch ts.side {
		case client.MsgSend:
			log.Println("joining sender for:", ts.secret)
			if _, ok := r.transfers[ts.secret]; ok {
				// should be very unlikely as the Service server generates Secrets!
				log.Println("skipping duplicate send client:", ts.secret)
				_ = ts.conn.Close()
				return
			}
			r.transfers[ts.secret] = &transfer{secret: ts.secret, send: ts.conn}
		case client.MsgRecv:
			log.Println("joining receiver for", ts.secret)
			if _, ok := r.transfers[ts.secret]; !ok {
				log.Println("skipping recv client because no active tx:", ts.secret)
				_ = ts.conn.Close()
				return
			}
			t := r.transfers[ts.secret]
			t.recv = ts.conn

			// sender and receiver are connected so now start relaying traffic
			go t.run(r)
		default:
			log.Println("skipping client because side is invalid:", ts.side)
			_ = ts.conn.Close()
		}
	}
}

// cleans up after ending a transfer for any reason
func (r *Service) close(secret string) {
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

// transfer an ongoing transfer between sender and receiver
type transfer struct {
	// secret is a unique key for transfer
	secret string

	// send is connection from the sender
	send net.Conn

	// recv is the connection to the receiver
	recv net.Conn
}

// transferSide a client side of a transfer
type transferSide struct {
	// conn is connection to client
	conn net.Conn

	// side of the transfer, either sender or receiver
	side client.Side

	// secret identifies transfer
	secret string
}

// Copies bytes from sender to receiver
func (t *transfer) run(r *Service) {
	defer r.close(t.secret)

	// Send "receiver is ready" message to sender so that the
	// sender can start sending bytes.
	enc := wire.NewEncoder(t.send)
	if err := enc.EncodeByte(byte(client.MsgRecv)); err != nil {
		log.Println("send recv ready failed:", err)
		return
	}

	// Now just pipe from sender to receiver
	// Note that the Service server doesn't care what messages are passed.
	if _, err := io.Copy(t.recv, t.send); err != nil {
		log.Println("tx.run for:", t.secret, " failed with:", err)
	}
}
