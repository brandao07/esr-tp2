package util

import (
	"bytes"
	"encoding/gob"
	"net"

	"github.com/brandao07/esr-tp2/src/entity"
)

func ReceiveNode(socket net.PacketConn) *entity.Node {
	// Read packet
	readBuff := make([]byte, 2024)
	n, _, err := socket.ReadFrom(readBuff)
	HandleError(err)

	// If node not on bootstrap
	if string(readBuff[:n]) == "NOT_FOUND" {
		return nil
	}

	// Decode packet
	decodeBuff := bytes.NewBuffer(readBuff)
	node := new(entity.Node)
	gobobj := gob.NewDecoder(decodeBuff)
	gobobj.Decode(node)

	return node
}

func SendNode(socket net.PacketConn, addr net.Addr, node entity.Node) {
	// Encode packet
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err := enc.Encode(node)
	HandleError(err)
	// Send packet
	_, err = socket.WriteTo(encodeBuff.Bytes(), addr)
	HandleError(err)
}
