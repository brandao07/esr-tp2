package util

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"

	"github.com/brandao07/esr-tp2/src/entity"
)

func ReceivePacket(socket net.PacketConn) (*entity.Packet, net.Addr) {
	// Read packet
	readBuff := make([]byte, 2024)
	//socket.SetReadDeadline(time.Now().Add(2 * time.Second))
	//FIXME: ReadFrom is not working properly when trying to read ACK or REQ packets (server side)
	_, addr, err := socket.ReadFrom(readBuff)
	HandleError(err)

	// Decode packet
	decodeBuff := bytes.NewBuffer(readBuff)
	pac := new(entity.Packet)
	gobobj := gob.NewDecoder(decodeBuff)
	gobobj.Decode(pac)

	return pac, addr
}

func SendPacket(socket net.PacketConn, addr net.Addr, id int, data []byte, state entity.PacketState) {
	// Convert packet id to byte array
	idBuff := new(bytes.Buffer)
	err := binary.Write(idBuff, binary.LittleEndian, uint64(id))
	HandleError(err)

	pac := entity.Packet{
		Id:    idBuff.Bytes(),
		Data:  data,
		State: []byte(state),
	}

	// Encode packet
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err = enc.Encode(pac)
	HandleError(err)

	// Send packet
	_, err = socket.WriteTo(encodeBuff.Bytes(), addr)
	HandleError(err)
}
