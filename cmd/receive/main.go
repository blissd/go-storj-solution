package main

import (
	"go-storj-solution/pkg/client"
	"io"
	"log"
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

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		log.Fatalln("output must be an existing directory")
	}

	s, err := client.NewService(addr)
	if err != nil {
		log.Fatalln("failed creating client:", err)
	}
	defer s.Close()

	r, name, err := s.Recv(secret)
	if err != nil {
		log.Fatalln("failed starting receive:", err)
	}

	filePath := path.Join(dir, name)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalln("failed creating output file:", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		log.Fatalln("failed receiving file:", err)
	}
}
