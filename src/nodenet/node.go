package nodenet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/brandao07/esr-tp2/src/util"
)

type Node struct {
	Address     string `json:"address"`
	Port        string `json:"port"`
	FullAddress string `json:"fullAddress"`
	Type        string `json:"type"`
	Neighbours  []Node `json:"neighbours"`
}

func ReadFromSocket(socket net.PacketConn, buffer []byte) (int, net.Addr) {
	n, sender, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	return n, sender
}

func receiveNode(socket net.PacketConn) *Node {
	// Read packet
	readBuff := make([]byte, 2024)
	n, _, err := socket.ReadFrom(readBuff)
	util.HandleError(err)

	// If node not on bootstrap
	if string(readBuff[:n]) == "NOT_FOUND" {
		return nil
	}

	// Decode packet
	decodeBuff := bytes.NewBuffer(readBuff)
	node := new(Node)
	gobobj := gob.NewDecoder(decodeBuff)
	gobobj.Decode(node)

	return node
}

func SendNode(socket net.PacketConn, addr net.Addr, node Node) {
	// Encode packet
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err := enc.Encode(node)
	util.HandleError(err)
	// Send packet
	_, err = socket.WriteTo(encodeBuff.Bytes(), addr)
	util.HandleError(err)
}

func GetNode(bootstrapAddress, serverAddress string) *Node {
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
	node := receiveNode(socket)

	if node == nil {
		util.HandleError(fmt.Errorf("node not found"))
	}

	return node
}

func Run(bootAddr string, id string) {

}
