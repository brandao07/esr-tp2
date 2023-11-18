package nodenet

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"

	"github.com/brandao07/esr-tp2/src/util"
)

type PacketState string

const (
	STREAMING  PacketState = "STREAMING"
	REQUESTING PacketState = "REQUESTING"
)

type Packet struct {
	Id    []byte
	Data  []byte
	State []byte
	//Timestamp time.Time //TODO: Add Timestamp to Packet
}

func DecodePacket(buff []byte) *Packet {
	// Decode packet
	decodeBuff := bytes.NewBuffer(buff)
	pac := new(Packet)
	gobobj := gob.NewDecoder(decodeBuff)
	gobobj.Decode(pac)

	return pac
}

func EncodePacket(pac *Packet) []byte {
	// Encode packet
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err := enc.Encode(pac)
	util.HandleError(err)

	return encodeBuff.Bytes()
}

func SendPacket(socket net.PacketConn, addr net.Addr, id int, data []byte, state PacketState) {
	// Convert packet id to byte array
	idBuff := new(bytes.Buffer)
	err := binary.Write(idBuff, binary.LittleEndian, uint64(id))
	util.HandleError(err)

	pac := Packet{
		Id:    idBuff.Bytes(),
		Data:  data,
		State: []byte(state),
	}

	// Send packet
	_, err = socket.WriteTo(EncodePacket(&pac), addr)
	util.HandleError(err)
}

func GetPacketId(pac *Packet) (uint64, error) {
	var id uint64
	readBuf := bytes.NewReader(pac.Id)
	err := binary.Read(readBuf, binary.LittleEndian, &id)
	return id, err
}
