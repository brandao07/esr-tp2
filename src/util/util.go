package util

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/brandao07/esr-tp2/src/entity"
)

func HandleError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func SplitIntoChunks(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize

		if end > len(data) {
			end = len(data)
		}

		chunks = append(chunks, data[i:end])
	}

	return chunks
}

func GetPacketId(pac *entity.Packet) (uint64, error) {
	var id uint64
	readBuf := bytes.NewReader(pac.Id)
	err := binary.Read(readBuf, binary.LittleEndian, &id)
	return id, err
}
