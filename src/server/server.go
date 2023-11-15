package server

import (
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

func getNode(bootstrapAddress, serverAddress string) *entity.Node {
	// Listen for UDP packets on the auxiliary address
	socket, err := net.ListenPacket("udp", serverAddress)
	util.HandleError(err)
	defer socket.Close()

	// Resolve the UDP address of the bootstrap server
	bootstrapAddr, err := net.ResolveUDPAddr("udp", bootstrapAddress)
	util.HandleError(err)

	// Send a "GET_NODE" request to the bootstrap server
	_, err = socket.WriteTo([]byte("GET_NODE"), bootstrapAddr)
	util.HandleError(err)

	// Receive the Node information from the bootstrap server
	node := util.ReceiveNode(socket)

	return node
}

func Run(bootstrapAddress string) {
	var wg sync.WaitGroup

	node := getNode(bootstrapAddress, "")

	wg.Add(1)
	go handleUDPRequest(&wg, node)
	wg.Wait()
}
