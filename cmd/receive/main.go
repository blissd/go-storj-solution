package main

import (
	"errors"
	"fmt"
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/wire"
	"io"
	"log"
	"net"
	"os"
	"path"
)

func main() {

	if len(os.Args) != 4 {
		log.Fatalln("Usage: receive <relay-host>:<relay-port> <secret-code> <output-directory>")
	}

	addr := os.Args[1]
	secret := os.Args[2]
	dir := os.Args[3]

	if err := run(addr, secret, dir); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(addr string, secret string, dir string) error {

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return errors.New("no such directory")
	}

	con, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("new service: %w", err)
	}
	defer con.Close()

	s := client.NewService(wire.NewEncoder(con), wire.NewDecoder(con))

	r, err := s.Recv(secret)
	if err != nil {
		return fmt.Errorf("starting receive: %w", err)
	}

	filePath := path.Join(dir, r.Name)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		return fmt.Errorf("receiving file: %w", err)
	}
	return nil
}
