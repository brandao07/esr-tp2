package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/brandao07/esr-tp2/src/entity"
	"github.com/brandao07/esr-tp2/src/util"
)

func setupServer(serverAddress string) net.PacketConn {
	socket, err := net.ListenPacket("udp", serverAddress)
	util.HandleError(err)

	fmt.Printf("Listening on %s\n", serverAddress)

	return socket
}

func readFromSocket(socket net.PacketConn, buffer []byte) (int, net.Addr) {
	n, sender, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	return n, sender
}

// TODO: Implement packet loss detection logic
func isPacketLossDetected(socket net.PacketConn) bool {
	fmt.Println("Verifying for potential packet loss")
	return false
}

func processRequest(socket net.PacketConn, addr net.Addr, request string, videoData []byte) {
	fmt.Printf("Received request from %s: %s\n", addr.String(), request)

	chunks := util.SplitIntoChunks(videoData, 1024)
	for i, chunk := range chunks {
		// Send packet
		util.SendPacket(socket, addr, i, chunk, entity.STREAMING)

		// Check for packet loss and retransmit if necessary
		if isPacketLossDetected(socket) {
			fmt.Printf("Packet loss detected for sequence number: %d\n", i)
			util.SendPacket(socket, addr, i, chunk, entity.STREAMING)
		}
	}

	// Send end of stream signal
	util.SendPacket(socket, addr, 0, nil, entity.FINISHED)
}

func handleUDPRequest(wg *sync.WaitGroup, node *entity.Node) {
	socket := setupServer(node.FullAddress)
	buffer := make([]byte, 1024)

	defer socket.Close()
	defer (*wg).Done()

	for {
		n, sender := readFromSocket(socket, buffer)
		request := string(buffer[:n])

		videoData, err := os.ReadFile(request)
		util.HandleError(err)

		go processRequest(socket, sender, request, videoData)
	}
}

func readBootstrapFile(filepath string) []entity.Node {
	jsonData, err := os.ReadFile(filepath)
	util.HandleError(err)

	var nodes []entity.Node
	err = json.Unmarshal(jsonData, &nodes)
	util.HandleError(err)

	return nodes
}

func getNode(bootstrapAddress, aux string) *entity.Node {
	socket, err := net.ListenPacket("udp", "") // FIXME: Use a random port
	util.HandleError(err)

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", bootstrapAddress)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte("GET_NODE"), address)
	util.HandleError(err)

	node := util.ReceiveNode(socket)

	return node
}

func processNode(socket net.PacketConn, nodes []entity.Node, address net.Addr) {
	for _, node := range nodes {
		if node.FullAddress == address.String() {
			util.SendNode(socket, address, node)
			fmt.Println("Sent node to", address.String())
			return
		}
	}
}

func handleBootstrapRequest(wg *sync.WaitGroup, isBootstrap bool) {
	defer (*wg).Done()

	if !isBootstrap {
		return
	}
	nodes := readBootstrapFile("bootstrapper.json")

	serverAddress := nodes[0].Address + ":" + nodes[0].BootstrapPort
	socket := setupServer(serverAddress)
	defer socket.Close()

	buffer := make([]byte, 2024)

	for {
		_, address := readFromSocket(socket, buffer)
		wg.Add(1)
		go processNode(socket, nodes, address)
	}
}

func Run(bootstrapAddress, aux string) {
	var wg sync.WaitGroup

	node := getNode(bootstrapAddress, aux)
	wg.Add(1)
	go handleUDPRequest(&wg, node)
	wg.Wait()
}

func RunBootstrap() {
	var wg sync.WaitGroup
	wg.Add(1)
	go handleBootstrapRequest(&wg, true)
	wg.Wait()
}
