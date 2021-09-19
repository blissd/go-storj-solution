package proxy

import (
	"github.com/go-kit/log"
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/wire"
	"io"
)

// Service manages transfers between senders and receivers.
type Service struct {
	// secrets source
	secrets Secrets

	// transfers that are in progress.
	// updated serially by functions processed from 'action' channel.
	transfers map[string]*transfer

	// action to add or remove transfers.
	// `Service` is effectively an actor.
	action chan func()

	logger log.Logger
}

func New(secrets Secrets, logger log.Logger) *Service {
	return &Service{
		secrets:   secrets,
		transfers: make(map[string]*transfer),
		action:    make(chan func()),
		logger:    logger,
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
// The Service takes ownership of an onboarded connection and will be responsible for closing it.
// Expected to be called from a go routine.
func (r *Service) Onboard(conn io.ReadWriteCloser) {
	dec := wire.NewDecoder(conn)

	var side client.Side
	{
		b, err := dec.DecodeByte()
		if err != nil {
			r.logger.Log("msg", "failed reading first byte", "err", err)
			_ = conn.Close()
			return
		}

		side = client.Side(b)
	}

	r.logger.Log("msg", "onboarding", "side", side)

	var secret string

	switch side {
	case client.MsgSend:
		// Onboarding a sender so generate and send secret for this transfer
		secret = r.secrets.Secret()
		if err := wire.NewEncoder(conn).EncodeString(secret); err != nil {
			r.logger.Log("msg", "failed sending secret", "err", err)
			_ = conn.Close()
			return
		}
	case client.MsgRecv:
		// Onboarding a receiver so read the secret for the transfer
		var err error
		if secret, err = dec.DecodeString(); err != nil {
			r.logger.Log("msg", "failed receiving secret", "err", err)
			_ = conn.Close()
			return
		}
	default:
		r.logger.Log("msg", "invalid client side", "side", side)
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
		switch ts.side {
		case client.MsgSend:
			r.logger.Log("msg", "joining", "side", ts.side, "secret", ts.secret)
			if _, ok := r.transfers[ts.secret]; ok {
				// should be very unlikely as the Service server generates Secrets!
				r.logger.Log("msg", "duplicate secret", "secret", ts.secret)
				_ = ts.conn.Close()
				return
			}
			r.transfers[ts.secret] = &transfer{secret: ts.secret, send: ts.conn}
		case client.MsgRecv:
			r.logger.Log("msg", "joining", "side", ts.side, "secret", ts.secret)
			if _, ok := r.transfers[ts.secret]; !ok {
				r.logger.Log("msg", "receiver provided unknown secret", "secret", ts.secret)
				_ = ts.conn.Close()
				return
			}
			t := r.transfers[ts.secret]
			t.recv = ts.conn

			// sender and receiver are connected so now start relaying traffic
			go t.run(r)
		default:
			r.logger.Log("msg", "failed join because client side is invalid", "side", ts.side)
			_ = ts.conn.Close()
		}
	}
}

// cleans up after ending a transfer for any reason
func (r *Service) close(secret string) {
	r.action <- func() {
		r.logger.Log("msg", "closing", "secret", secret)
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
	send io.ReadWriteCloser

	// recv is the connection to the receiver
	recv io.ReadWriteCloser
}

// transferSide a client side of a transfer
type transferSide struct {
	// conn is connection to client
	conn io.ReadWriteCloser

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
		r.logger.Log(
			"msg", "notifying sender of receiver failed",
			"secret", t.secret,
			"err", err,
		)
		return
	}

	// Now just pipe from sender to receiver
	// Note that the Service server doesn't care what messages are passed.
	if _, err := io.Copy(t.recv, t.send); err != nil {
		r.logger.Log(
			"msg", "notifying sender of receiver failed",
			"secret", t.secret,
			"err", err,
		)
	}
}
