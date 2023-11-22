package nodenet

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"time"

	"github.com/brandao07/esr-tp2/src/util"
)

type PacketState string

const (
	STREAMING      PacketState = "STREAMING"
	REQUESTING     PacketState = "REQUESTING"
	STOP_STREAMING PacketState = "STOP_STREAMING"
	ABORT          PacketState = "ABORT"
)

type Packet struct {
	Id          []byte
	Data        []byte
	State       PacketState
	File        string
	Source      string
	Destination string
	Path        []string
	Timestamp   time.Time
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

func SendPacket(socket net.PacketConn, addr net.Addr, packet *Packet) {
	// Send packet
	_, err := socket.WriteTo(EncodePacket(packet), addr)
	util.HandleError(err)
}

func (pac *Packet) DecodeId() (uint64, error) {
	var id uint64
	readBuf := bytes.NewReader(pac.Id)
	err := binary.Read(readBuf, binary.LittleEndian, &id)
	return id, err
}
