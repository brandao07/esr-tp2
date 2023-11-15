package server

import (
	"encoding/json"
	"os"
	"testing"
	"time"

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

	serverAddress := nodes[0].Address + ":" + nodes[0].BootstrapPort
	t.Log(serverAddress)
}

func TestRunBootstrap(t *testing.T) {
	// Start the bootstrap server in a goroutine
	go func() {
		RunBootstrap("../../bootstrapper.json")
	}()

	// Wait for the bootstrap server to be ready
	time.Sleep(100 * time.Millisecond)

	node := getNode("127.0.0.1:10001", "127.0.0.1:30000")
	if node == nil {
		t.Errorf("Node not found")
		return
	}

	time.Sleep(100 * time.Millisecond)
}
