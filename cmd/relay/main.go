package main

import (
	"github.com/blissd/golang-storj-solution/pkg/session"
	"github.com/blissd/golang-storj-solution/pkg/wire"
	"log"
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

	secrets := NewRandomSecrets(6, time.Now().UnixNano())
	r := NewRelay()
	go r.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go onboard(r, secrets, conn)
	}
}

// Onboards a new connection for a sender or a receiver.
// For a sender a secret will be generated and sent to the sender.
// For a receiver a secret will be read from the connection.
// A valid client then joins a transfer, either creating it for a sender
// or being associated with an existing transform for a receiver.
func onboard(r *Relay, secrets Secrets, conn net.Conn) {

	dec := wire.NewDecoder(conn)
	clientType, err := dec.DecodeByte()
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
		secret = secrets.Secret()
		log.Println("generated secret is", secret)
		err = wire.NewEncoder(conn).EncodeString(secret)
		if err != nil {
			log.Println("send secret in onboard:", err)
			conn.Close()
			return
		}
	case session.MsgRecv:
		log.Println("receiving secret")
		secret, err = dec.DecodeString()
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
