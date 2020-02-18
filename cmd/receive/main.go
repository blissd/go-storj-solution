package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 4 {
		log.Fatalln("Usage: receive <relay-host>:<relay-port> <secret-code> <output-directory>")
	}

	addr := os.Args[1]
	secret := os.Args[2]
	dir := os.Args[3]

	fmt.Println("addr:", addr, "secret:", secret, "dir:", dir)

}
