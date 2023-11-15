package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/brandao07/esr-tp2/src/entity"
	"github.com/brandao07/esr-tp2/src/util"
)

func readNodesFromFile(filepath string) []entity.Node {
	// Read the content of the JSON file
	jsonData, err := os.ReadFile(filepath)
	util.HandleError(err)

	// Unmarshal the JSON data into a slice of entity.Node
	var nodes []entity.Node
	err = json.Unmarshal(jsonData, &nodes)
	util.HandleError(err)

	return nodes
}

func processNode(socket net.PacketConn, nodes []entity.Node, fullAddr net.Addr) {
	// Extract the IP address from the fullAddr net.Addr
	addr := strings.Split(fullAddr.String(), ":")[0]
	fmt.Println(fullAddr)
	for _, node := range nodes {
		if node.Address == addr {
			// Send the node information using the provided socket to the address represented by fullAddr
			util.SendNode(socket, fullAddr, node)

			fmt.Println("Sent node to", node.FullAddress)
			return
		}
	}
}

func handleBootstrapRequest(serverAddress string, nodes []entity.Node, readySignal chan<- struct{}) {
	socket := setupServer(serverAddress)
	defer socket.Close()

	buffer := make([]byte, 2024)
	// add a channel here to notify the main thread that the bootstrap server is ready
	close(readySignal)

	for {

		_, address := readFromSocket(socket, buffer)

		go processNode(socket, nodes, address)
	}
}

func RunBootstrap(filePath string) {
	nodes := readNodesFromFile(filePath)
	serverAddress := nodes[0].Address + ":" + nodes[0].BootstrapPort
	bootstrapReady := make(chan struct{})
	go handleBootstrapRequest(serverAddress, nodes, bootstrapReady)

	// Wait for the bootstrap server to be ready
	<-bootstrapReady
	Run(serverAddress)
}
