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
	JWTSecret:        "jwtsecret",
}

// Ethereum service config
var ethConf = &ethconfig.Config{
	NetworkId: 12345,
	Genesis:   loadGenesis("genesis.json"),
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
	defer stack.Close()

	// Start the Ethereum service
	if err := ethService.Start(); err != nil {
		log.Fatal("Failed to start eth service: ", err)
	}
	defer ethService.Stop()

	// Unlock the account
	// password := "testpassword"
	// accMan := stack.AccountManager()
	// accs := accMan.Wallets()
	// if len(accs) == 0 {
	//     log.Fatalf("No wallets found")
	// }
	// var unlockedAccount accounts.Account
	// for _, wallet := range accs {
	//     for _, account := range wallet.Accounts() {
	//         if account.Address.Hex() == "0xC0AB77b270768F317Aca3ad03Cb9A1c17232F2C2" {
	//             unlockedAccount = account
	//             err = wallet.Unlock(unlockedAccount, password)
	//             if err != nil {
	//                 log.Fatalf("Failed to unlock account: %v", err)
	//             }
	//         }
	//     }
	// }

	// keystore.unl

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
