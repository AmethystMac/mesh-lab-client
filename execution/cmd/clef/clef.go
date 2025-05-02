package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {
	clefPath := "clef"
	ipcPath := "cmd/clef"
	keystoreDir := "data/keystore"
	rules := "config/signer_rules.js"
	chainID := "12345"

	// Start Clef process
	cmd := exec.Command(
		clefPath,
		"--keystore", keystoreDir,
		"--chainid", chainID,
		"--ipcpath", ipcPath,
		"--configdir", keystoreDir,
		"--rules", rules,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start clef: %v", err)
	}
	log.Printf("Clef started with PID %d", cmd.Process.Pid)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down Clef...")
	if err := cmd.Process.Kill(); err != nil {
		log.Printf("Failed to kill Clef: %v", err)
	} else {
		log.Println("Clef stopped.")
	}
}
