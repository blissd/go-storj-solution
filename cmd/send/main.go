package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 3 {
		log.Fatalln("Usage: send <relay-host>:<relay-port> <file-to-send>")
	}

	addr := os.Args[1]
	file := os.Args[2]

	fmt.Println("addr:", addr, "file:", file)

}
