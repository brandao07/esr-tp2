package main

import (
	"fmt"
	"os"

	"github.com/brandao07/esr-tp2/client"
	"github.com/brandao07/esr-tp2/server"
)

func main() {
	switch t := os.Args[1]; t {
	case "Server":
		server.Run()
	case "Client":
		client.Run()
	default:
		panic(fmt.Errorf("Invalid argument for type! (Server, Client)\n"))
	}
}