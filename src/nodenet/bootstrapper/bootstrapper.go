package bootstrapper

import (
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/brandao07/esr-tp2/src/nodenet"
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

func SendNode(socket net.PacketConn, addr net.Addr, node nodenet.Node) {
	// Send node
	_, err := socket.WriteTo(node.EncodeNode(), addr)
	util.HandleError(err)
}

func processNode(socket net.PacketConn, nodes []nodenet.Node, id []byte, addr net.Addr) {
	for _, node := range nodes {
		if node.Id == string(id) {
			// Send the node information using the provided socket to the address represented by fullAddr
			SendNode(socket, addr, node)
			log.Printf("BOOTSTRAPPER: node found %s\n", node.Address+":"+node.Port)
			return
		}
	}
	log.Println("BOOTSTRAPPER: node not found")
	socket.WriteTo([]byte("NOT_FOUND"), addr)
}

func handleBootstrapRequest(addr string, nodes []nodenet.Node, readySignal chan<- struct{}) {
	socket := nodenet.SetupSocket(addr)
	defer socket.Close()
	log.Printf("BOOTSTRAPPER: Listening on %s\n", addr)

	// notify the main thread that the bootstrap server is ready
	close(readySignal)

	var buffer []byte
	for {
		buffer = make([]byte, 2024)
		n, addr := nodenet.ReadFromSocket(socket, buffer)
		go processNode(socket, nodes, buffer[:n], addr)
	}
}

func Run(filePath string) {
	nodes := readNodesFromFile(filePath)
	bootAddr := nodes[0].Address + ":" + "10001"
	bootstrapReady := make(chan struct{})
	go handleBootstrapRequest(bootAddr, nodes, bootstrapReady)

	// Wait for the bootstrap server to be ready
	<-bootstrapReady
	nodenet.StartNode(&nodes[0])
}
