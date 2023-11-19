package main

import (
	"errors"
	"os"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/nodenet/bootstrapper"
	"github.com/brandao07/esr-tp2/src/nodenet/client"
	"github.com/brandao07/esr-tp2/src/nodenet/server"
	"github.com/brandao07/esr-tp2/src/util"
)

func validateArgs(minArgs int, errMsg string) {
	if len(os.Args) < minArgs {
		util.HandleError(errors.New(errMsg))
		os.Exit(1)
	}
}

func main() {
	validateArgs(2, "insufficient number of arguments")
	switch t := os.Args[1]; t {
	case "Bootstrap":
		bootstrapper.Run("bootstrapper.json")
	case "Node":
		validateArgs(3, "insufficient number of arguments for Node mode")
		node := nodenet.GetNode(os.Args[2], os.Args[3])
		nodenet.StartNode(node)
	case "Server":
		validateArgs(4, "insufficient number of arguments for Server mode")
		node := nodenet.GetNode(os.Args[2], os.Args[3])
		server.Run(node, os.Args[4])
	case "Client":
		validateArgs(3, "insufficient number of arguments for Client mode")
		client.Run(os.Args[2], os.Args[3])
	default:
		util.HandleError(errors.New("invalid argument for type! (Bootstrap, Server, Client)"))
	}
}
