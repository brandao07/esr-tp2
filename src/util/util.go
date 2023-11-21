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

func ContainsString(s []string, search string) bool {
	for _, val := range s {
		if val == search {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, strToRemove string) []string {
	var result []string

	for _, str := range slice {
		if str != strToRemove {
			result = append(result, str)
		}
	}

	return result
}
