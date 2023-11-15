package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/brandao07/esr-tp2/src/entity"
)

func TestNodeUnmarshal(t *testing.T) {
	// Read JSON data from file
	jsonData, err := os.ReadFile("../../bootstrapper.json")
	if err != nil {
		t.Errorf("Error reading JSON file: %v", err)
		return
	}

	var nodes []entity.Node
	err = json.Unmarshal(jsonData, &nodes)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
		return
	}
	t.Log(nodes[0].Address)
}
