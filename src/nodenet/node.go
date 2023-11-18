package nodenet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/brandao07/esr-tp2/src/util"
)

type Node struct {
	Id         string `json:"id"`
	Address    string `json:"address"`
	Port       string `json:"port"`
	Neighbours []Node `json:"neighbours"`
}

var videos = []string{}

func (n *Node) DecodeNode(buff []byte) {
	// Decode node
	decodeBuff := bytes.NewBuffer(buff)
	gobobj := gob.NewDecoder(decodeBuff)
	err := gobobj.Decode(n)
	util.HandleError(err)
}

func (n *Node) EncodeNode() []byte {
	// Encode node
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err := enc.Encode(n)
	util.HandleError(err)

	return encodeBuff.Bytes()
}

func SetupSocket(addr string) net.PacketConn {
	// Listen for UDP packets on the auxiliary address
	socket, err := net.ListenPacket("udp", addr)
	util.HandleError(err)

	return socket
}

func ReadFromSocket(socket net.PacketConn, buffer []byte) (int, net.Addr) {
	n, sender, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	return n, sender
}

func receiveNode(socket net.PacketConn) *Node {
	// Read node
	buff := make([]byte, 2024)
	n, _ := ReadFromSocket(socket, buff)

	// If node not on bootstrap
	if string(buff[:n]) == "NOT_FOUND" {
		return nil
	}

	node := Node{}
	node.DecodeNode(buff[:n])

	return &node
}

func GetNode(bootAddr, id string) *Node {
	// Listen for UDP packets on the auxiliary address
	socket := SetupSocket("")
	defer socket.Close()

	// Resolve the UDP address of the bootstrap server
	addr, err := net.ResolveUDPAddr("udp", bootAddr)
	util.HandleError(err)

	// Send a "GET_NODE" request to the bootstrap server
	_, err = socket.WriteTo([]byte(id), addr)
	util.HandleError(err)

	// Receive the Node information from the bootstrap server
	node := receiveNode(socket)

	if node == nil {
		util.HandleError(fmt.Errorf("node not found"))
	}

	return node
}

func requestVideoStreaming(socket net.PacketConn, serverAddr, payload string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte(payload), addr)
	log.Printf("NODE: Sending request to %s: %s\n", serverAddr, payload)
	util.HandleError(err)
}

func streamingRequestHandler(wg *sync.WaitGroup, node *Node) {
	nodeAddr := node.Address + ":" + node.Port
	socket := SetupSocket(nodeAddr)
	log.Printf("NODE: Listening on %s\n", socket.LocalAddr().String())
	buffer := make([]byte, 1024)

	defer socket.Close()
	defer (*wg).Done()

	for {
		n, _ := ReadFromSocket(socket, buffer)
		req := string(buffer[:n])
		if util.ContainsString(videos, req) {
			//TODO: Locate the thread that is streaming the video and send the video data to the client
			log.Printf("NODE: Video %s already being streamed\n", req)
		} else {
			//TODO: Send request to neighbours
			requestVideoStreaming(socket, "", req)
		}
		//TODO: replace "video data" with the actual video data

	}
}

func Run(node *Node) {
	var wg sync.WaitGroup

	wg.Add(2)
	//go startVideoStreaming()
	go streamingRequestHandler(&wg, node)
	wg.Wait()
}
