package main

import (
	"fmt"
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

	fmt.Println("addr:", addr, "file:", file)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Failed to opening file:", err)
	}
	defer file.Close()

	// chat with relay server to initiate a transfer session

	con, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("Failed dialing relay server:", err)
	}

	con.Write(relay.Start())
}
