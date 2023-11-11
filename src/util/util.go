package util

import (
	"log"
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
