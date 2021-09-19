package main

import (
	"fmt"
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/wire"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Usage: send <relay-host>:<relay-port> <file-to-send>")
	}

	addr := os.Args[1]
	filePath := os.Args[2]

	if err := run(addr, filePath); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(addr string, filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("failed to opening file:", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Fatalln("stating file:", err)
	}

	request := &client.SendRequest{
		Body:   file,
		Name:   file.Name(),
		Length: info.Size(),
	}

	con, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("new service: %w", err)
	}
	defer con.Close()

	s := client.NewService(wire.NewEncoder(con), wire.NewDecoder(con))

	response, err := s.Send(request)
	if err != nil {
		log.Fatalln("sending:", err)
	}

	fmt.Println(response.Secret)

	select {
	case err, gotError := <-response.Errors:
		if gotError {
			log.Fatalln("failed sending file:", err)
		}
	}
	return nil
}
