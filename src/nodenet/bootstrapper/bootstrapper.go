package bootstrapper

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/nodenet/server"
	"github.com/brandao07/esr-tp2/src/util"
)

func readNodesFromFile(filepath string) []nodenet.Node {
	// Read the content of the JSON file
	jsonData, err := os.ReadFile(filepath)
	util.HandleError(err)

	// Unmarshal the JSON data into a slice of entity.Node
	var nodes []nodenet.Node
	err = json.Unmarshal(jsonData, &nodes)
	util.HandleError(err)

	return nodes
}

func processNode(socket net.PacketConn, nodes []nodenet.Node, fullAddr net.Addr) {
	// Extract the IP address from the fullAddr net.Addr
	addr := strings.Split(fullAddr.String(), ":")[0]
	for _, node := range nodes {
		if node.Address == addr {
			// Send the node information using the provided socket to the address represented by fullAddr
			nodenet.SendNode(socket, fullAddr, node)
			log.Printf("BOOTSTRAPPER: node found %s\n", node.FullAddress)
			return
		}
	}
	log.Println("BOOTSTRAPPER: node not found")
	socket.WriteTo([]byte("NOT_FOUND"), fullAddr)
}

func handleBootstrapRequest(serverAddress string, nodes []nodenet.Node, readySignal chan<- struct{}) {
	socket, err := net.ListenPacket("udp", serverAddress)
	util.HandleError(err)

	log.Printf("BOOTSTRAPPER: Listening on %s\n", serverAddress)
	defer socket.Close()

	buffer := make([]byte, 2024)
	// notify the main thread that the bootstrap server is ready
	close(readySignal)

	for {
		_, address := nodenet.ReadFromSocket(socket, buffer)
		go processNode(socket, nodes, address)
	}
}

func Run(filePath string) {
	nodes := readNodesFromFile(filePath)
	serverAddress := nodes[0].Address + ":" + "10001"
	bootstrapReady := make(chan struct{})
	go handleBootstrapRequest(serverAddress, nodes, bootstrapReady)

	// Wait for the bootstrap server to be ready
	<-bootstrapReady
	server.Run(serverAddress, "n1")
}