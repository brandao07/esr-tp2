package bootstrapper

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/brandao07/esr-tp2/src/nodenet"
)

func TestNodeUnmarshal(t *testing.T) {
	// Read JSON data from file
	jsonData, err := os.ReadFile("../../bootstrapper.json")
	if err != nil {
		t.Errorf("Error reading JSON file: %v", err)
		return
	}

	var nodes []nodenet.Node
	err = json.Unmarshal(jsonData, &nodes)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
		return
	}

	serverAddress := nodes[0].Address + ":" + "10001"
	t.Log(serverAddress)
}

func TestRunBootstrap(t *testing.T) {
	// Start the bootstrap server in a goroutine
	go func() {
		Run("../../bootstrapper.json")
	}()

	// Wait for the bootstrap server to be ready
	time.Sleep(100 * time.Millisecond)

	node := nodenet.GetNode("127.0.0.1:10001", "127.0.0.1:30000")
	if node == nil {
		t.Errorf("Node not found")
		return
	}

	time.Sleep(100 * time.Millisecond)
}
