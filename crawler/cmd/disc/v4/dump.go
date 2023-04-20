package v4

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

const jsonIndent = "    "

type scanSet map[string][]*enode.Node

func writeScanJSON(file string, scan scanSet) {
	nodesJSON, err := json.MarshalIndent(scan, "", jsonIndent)
	if err != nil {
		panic(fmt.Errorf("json error. error: %v", err))
	}
	if file == "" {
		os.Stdout.Write(nodesJSON)
		return
	}

	if err := os.WriteFile(file, nodesJSON, 0644); err != nil {
		panic(fmt.Errorf("dumping error. error: %v", err))
	}
}
