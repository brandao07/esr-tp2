package main

import (
	"errors"
	"os"

	"github.com/brandao07/esr-tp2/src/client"
	"github.com/brandao07/esr-tp2/src/server"
	"github.com/brandao07/esr-tp2/src/test"
	"github.com/brandao07/esr-tp2/src/util"
)

func main() {
	switch t := os.Args[1]; t {
	case "Server":
		server.Run(os.Args[2])
	case "Client":
		client.Run(os.Args[2], os.Args[3])
	case "Scenario1":
		config := test.Scenario1Config{
			ServerAddress: "127.0.0.1:30000",
			RPAddress:     "127.0.0.1:10001",
		}
		test.Scenario1(config)
	case "Scenario2":
		//TODO
		panic("not implemented")
	default:
		util.HandleError(errors.New("invalid argument for type! (Server, Client, Scenario1, Scenario2)"))
	}
}
