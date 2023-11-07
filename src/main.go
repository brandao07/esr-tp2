package main

import (
	"errors"
	"os"

	"github.com/brandao07/esr-tp2/src/client"
	"github.com/brandao07/esr-tp2/src/server"
	"github.com/brandao07/esr-tp2/src/util"
)

func main() {
	switch t := os.Args[1]; t {
	case "Server":
		server.Run()
	case "Client":
		client.Run()
	default:
		util.HandleError(errors.New("invalid argument for type! (Server, Client)"))
	}
}
