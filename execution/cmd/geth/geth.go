package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
)

// Node configuration
var stackConf = &node.Config{
	DataDir: "data",
	P2P: p2p.Config{
		ListenAddr:  ":12345",
		NoDiscovery: true,
	},
	HTTPHost:         "127.0.0.1",
	HTTPPort:         8551,
	HTTPCors:         []string{"*"},
	HTTPModules:      []string{"eth", "net", "web3"},
	AuthAddr:         "127.0.0.1",
	AuthPort:         8551,
	AuthVirtualHosts: []string{"*"},
	JWTSecret:        "config/jwtsecret.txt",
	// ExternalSigner: ,
}

// Ethereum service config
var ethConf = &ethconfig.Config{
	NetworkId: 12345,
	Genesis:   loadGenesis("config/genesis.json"),
}

func main() {

	stack, err := node.New(stackConf)
	if err != nil {
		log.Fatal("Failed to create node: ", err)
	}

	ethService, err := eth.New(stack, ethConf)
	if err != nil {
		log.Fatal("Failed to create eth service: ", err)
	}

	// Start the node
	if err := stack.Start(); err != nil {
		log.Fatal("Failed to start node: ", err)
	}

	log.Printf("Node started. Listening on port %d.", stackConf.HTTPPort)

	defer stack.Close()

	// Start the Ethereum service
	if err := ethService.Start(); err != nil {
		log.Fatal("Failed to start eth service: ", err)
	}

	log.Printf("Eth service started.")

	defer ethService.Stop()

	log.Println("Node started. Press Ctrl+C to stop.")
	select {}
}

// Load the genesis.json file
func loadGenesis(filePath string) *core.Genesis {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open genesis file: %v", err)
	}
	defer file.Close()

	var genesis core.Genesis
	if err := json.NewDecoder(file).Decode(&genesis); err != nil {
		log.Fatalf("Failed to decode genesis file: %v", err)
	}
	return &genesis
}
