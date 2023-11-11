package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

func main() {
	// Test data
	id := 123
	data := []byte("test data")

	// Simulate server side
	serverSocket, _ := net.ListenPacket("udp", "127.0.0.1:0")
	defer serverSocket.Close()

	go func() {
		// Simulate server receiving a packet
		receivedPacket, _ := ReceivePacket(serverSocket)
		fmt.Printf("Received Packet: %+v\n", *receivedPacket)
	}()

	// Simulate client side
	clientSocket, _ := net.ListenPacket("udp", "")
	defer clientSocket.Close()

	// Simulate client sending a packet
	SendPacket(clientSocket, serverSocket.LocalAddr(), id, data, "REQUESTING")

	// Wait for a moment to let the server receive the packet
	time.Sleep(1 * time.Second)
}

// The following definitions are simplified versions of your entity and HandleError functions.
// Make sure to replace them with your actual implementations.
type PacketState string

type Packet struct {
	Id    []byte
	Data  []byte
	State []byte
}

func HandleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Replace this with your actual implementation of the entity.PacketState type

// Replace this with your actual implementation of the entity.PacketState type
func ReceivePacket(socket net.PacketConn) (*Packet, net.Addr) {
	// Read packet
	readBuff := make([]byte, 2024)
	socket.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, addr, err := socket.ReadFrom(readBuff)
	HandleError(err)

	// Decode packet
	decodeBuff := bytes.NewBuffer(readBuff)
	pac := new(Packet)
	gobobj := gob.NewDecoder(decodeBuff)
	err = gobobj.Decode(pac)
	HandleError(err)

	return pac, addr
}

// Replace this with your actual implementation of the entity.PacketState type
func SendPacket(socket net.PacketConn, addr net.Addr, id int, data []byte, state PacketState) {
	// Convert packet id to byte array
	idBuff := new(bytes.Buffer)
	err := binary.Write(idBuff, binary.LittleEndian, uint64(id))
	HandleError(err)
	pac := Packet{
		Id:    idBuff.Bytes(),
		Data:  data,
		State: []byte(state),
	}
	fmt.Printf("Sent Packet: %+v\n", pac)

	// Encode packet
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err = enc.Encode(pac)
	HandleError(err)

	// Send packet
	_, err = socket.WriteTo(encodeBuff.Bytes(), addr)
	HandleError(err)
}
