// Copyright (c) 2026 Aldian Okto. All rights reserved.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"eterbit/core"
	"eterbit/node"
	"eterbit/storage/wallet"

	_ "github.com/cloudflare/circl/sign/dilithium/mode3"
)

// MempoolFile designates the absolute or relative file path utilized for persisting unconfirmed network transaction queues locally.
const MempoolFile = "eterbit_data/mempool.json"

// main serves as the primary entry point for the command-line interface application, parsing operational arguments and routing execution flows accordingly.
func main() {
	// Initialize distinct command-line flag sets for various administrative and operational subcommands.
	walletCreateCmd := flag.NewFlagSet("create", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	explorerCmd := flag.NewFlagSet("explorer", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)

	// Define specific parameter bindings for individual command flags.
	walletName := walletCreateCmd.String("name", "keystore.json", "Custom filename for the wallet")

	sendRecipient := sendCmd.String("to", "", "Recipient destination address")
	sendAmount := sendCmd.Uint64("amount", 0, "Transfer value amount")
	sendFee := sendCmd.Uint64("fee", 2, "Transaction fee")
	walletSource := sendCmd.String("wallet", "keystore.json", "Wallet filename to use")

	// Validate whether adequate command-line arguments have been provided by the executing operator.
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Switch execution branch based upon the primary subcommand supplied in system arguments.
	switch os.Args[1] {
	case "create":
		walletCreateCmd.Parse(os.Args[2:])
		handleCreateWallet(*walletName)
	case "balance":
		balanceCmd.Parse(os.Args[2:])
		handleCheckBalance()
	case "send":
		sendCmd.Parse(os.Args[2:])
		handleSendTx(*sendRecipient, *sendAmount, *sendFee, *walletSource)
	case "node":
		nodeCmd.Parse(os.Args[2:])
		handleRunNode()
	case "explorer":
		explorerCmd.Parse(os.Args[2:])
		handleExploreBlockchain()
	case "mine":
		mineCmd.Parse(os.Args[2:])
		handleManualMine()
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage outputs the standard command-line manual instructions and available command options to standard output.
func printUsage() {
	// Render comprehensive structural documentation regarding available console commands.
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT BLOCKCHAIN CLI MANAGER (BITCOIN-LIKE ARCHITECTURE)")
	fmt.Println("================================================================================")
	fmt.Println("Available commands:")
	fmt.Println("  go run eterbit.go create -name <file.json>")
	fmt.Println("  go run eterbit.go balance")
	fmt.Println("  go run eterbit.go send -to <addr> -amount <val> [-fee <val>] [-wallet <file>]")
	fmt.Println("  go run eterbit.go node")
	fmt.Println("  go run eterbit.go mine")
	fmt.Println("  go run eterbit.go explorer")
	fmt.Println("================================================================================")
}

// handleCreateWallet generates a new cryptographic wallet instance and persists its state securely to disk.
func handleCreateWallet(filename string) {
	// Ensure that the target data directory infrastructure exists prior to file creation operations.
	os.MkdirAll("eterbit_data", 0755)
	filePath := filepath.Join("eterbit_data", filename)

	// Verify whether a wallet configuration file already occupies the designated storage path.
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("[WALLET] Wallet file '%s' already exists!\n", filename)
		return
	}

	// Invoke cryptographic wallet generation routines to obtain public/private key pairs and system addresses.
	addr, privKey, pubBytes, err := wallet.CreateOrLoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[WALLET] Failed: %v\n", err)
		return
	}

	// Encode raw cryptographic byte arrays into standardized hexadecimal string representations.
	pubHex := hex.EncodeToString(pubBytes)
	privBytes, _ := privKey.MarshalBinary()
	privHex := hex.EncodeToString(privBytes)

	// Persist the generated credentials to the specified target storage location.
	wallet.SaveWalletCustom(filePath, addr, pubHex, privHex)

	// Output successful initialization diagnostics to the standard output channel.
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT NEW WALLET CREATED")
	fmt.Println("================================================================================")
	fmt.Printf(" Filename : %s\n", filePath)
	fmt.Printf(" Address  : %s\n", addr)
	fmt.Println("--------------------------------------------------------------------------------")
}

// handleCheckBalance queries the persistent ledger database to enumerate all registered account balances and associated nonces.
func handleCheckBalance() {
	// Initialize the underlying state ledger mechanism referencing the storage directory and network difficulty level.
	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_VIEWER")
	fmt.Println("================================================================================")
	fmt.Println(" REGISTERED ACCOUNT BALANCES IN LEVELDB")
	fmt.Println("================================================================================")
	
	// Evaluate whether any account states currently exist within the ledger database mapping.
	if len(ledger.State) == 0 {
		fmt.Println(" No accounts currently recorded in the state ledger.")
		return
	}
	
	// Iterate through every active account entry within the ledger state registry (Formatted with 8 decimals).
	for addr, acc := range ledger.State {
		fmt.Printf(" Address: %s | Balance: %.8f Coins | Nonce: %d\n", addr, node.ToDecimal(acc.Balance), acc.Nonce)
	}
	fmt.Println("================================================================================")
}

// saveMempoolToDisk serializes the active transaction pool collection and writes the resulting data structure directly to storage.
func saveMempoolToDisk(mempool []*core.Transfer) {
	// Guarantee that the local data directory structure is fully provisioned.
	os.MkdirAll("eterbit_data", 0755)
	
	// Marshal the transaction array into a cleanly formatted JSON byte structure.
	data, _ := json.MarshalIndent(mempool, "", "  ")
	
	// Write the serialized data payload directly into the target disk file path.
	os.WriteFile(MempoolFile, data, 0644)
}

// loadMempoolFromDisk reads the serialized transaction pool dataset from disk and unmarshals it into memory.
func loadMempoolFromDisk() []*core.Transfer {
	var mempool []*core.Transfer
	
	// Read raw byte contents from the designated mempool storage file.
	data, err := os.ReadFile(MempoolFile)
	if err != nil {
		return mempool
	}
	
	// Unmarshal the raw JSON byte array back into a collection of transfer transaction pointers.
	json.Unmarshal(data, &mempool)
	return mempool
}

// handleSendTx constructs, signs, and broadcasts a new value transfer transaction into the network mempool architecture.
func handleSendTx(recipient string, amount uint64, fee uint64, walletFile string) {
	// Validate parameter completeness before executing transaction construction.
	if recipient == "" || amount == 0 {
		fmt.Println("[CLI] Incomplete arguments! Use -to and -amount.")
		return
	}

	if walletFile == "" {
		walletFile = "keystore.json"
	}

	// Resolve the comprehensive filesystem path for the requested keystore file.
	filePath := filepath.Join("eterbit_data", walletFile)
	addrMiner, _, _, _ := wallet.LoadWalletCustom(filePath)
	
	// Initialize the structural ledger context for processing local state interactions.
	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	
	// Load the sender cryptographic wallet data structures from local storage.
	addrA, privKeyA, pubBytesA, err := wallet.LoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[CLI] Failed to load wallet from %s: %v\n", filePath, err)
		return
	}

	// Bootstrap default ledger account balance entries for test execution continuity.
	ledger.State[addrA] = &node.AccountState{
		Balance: node.InitialAirdrop,
		Nonce:   0,
	}
	
	// Ensure compatibility across alternative address string prefix formats within the ledger state.
	if len(addrA) > 4 && addrA[:4] == "etrb" {
		ledger.State[addrA[4:]] = &node.AccountState{
			Balance: node.InitialAirdrop,
			Nonce:   0,
		}
	} else {
		ledger.State["etrb"+addrA] = &node.AccountState{
			Balance: node.InitialAirdrop,
			Nonce:   0,
		}
	}

	// Compute a dynamically adjusted unique nonce parameter utilizing high-resolution nanosecond timestamps.
	currentNonce := ledger.State[addrA].Nonce + uint64(time.Now().UnixNano()%100000)

	fmt.Printf("[CLI] Constructing transaction from %s (via %s) to %s (Amount: %.8f, Fee: %.8f)...\n", addrA, walletFile, recipient, node.ToDecimal(amount), node.ToDecimal(fee))

	// Instantiate a cryptographic transfer object using private key signing mechanisms.
	tx := core.NewTransfer(privKeyA, pubBytesA, recipient, amount, fee, currentNonce)
	
	// Append the newly created transaction into the local persistent disk mempool queue.
	existingMempool := loadMempoolFromDisk()
	existingMempool = append(existingMempool, tx)
	saveMempoolToDisk(existingMempool)

	// Print operational confirmation messages regarding successful transaction broadcasting.
	fmt.Printf("[MEMPOOL] Transaction broadcasted to network pool! ID: %s...\n", tx.ComputeID()[:16])
	fmt.Println("[CLI] Transaction waiting for node validator to mine into a block.")
}

// handleRunNode initiates a continuous background validation daemon process that periodically polls and processes pending transactions.
func handleRunNode() {
	fmt.Println("[SYS] Booting Eterbit Live Node (Bitcoin Core Style)...")
	
	// Load the default node mining beneficiary address from the core keystore configuration.
	addrMiner, _, _, err := wallet.LoadWalletCustom("eterbit_data/keystore.json")
	if err != nil {
		fmt.Printf("[NODE] Failed to load default miner wallet: %v\n", err)
		return
	}

	// Initialize the persistent ledger instance tied to the active validator address.
	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	
	// Spawn an asynchronous background worker goroutine to manage recurring block generation loops.
	go func() {
		for {
			// Suspend execution for a fixed interval before checking the disk mempool queue.
			time.Sleep(3 * time.Second)
			
			// Retrieve any pending transactions stored within the shared disk mempool layer.
			diskMempool := loadMempoolFromDisk()
			if len(diskMempool) > 0 {
				// Acquire thread synchronization locks prior to modifying internal ledger states.
				ledger.Mu.Lock()
				ledger.Mempool = diskMempool
				ledger.Mu.Unlock()

				fmt.Println("[NODE] Pending transactions detected in mempool. Starting Proof-of-Work...")
				
				// Execute the computational block mining procedure.
				ledger.MineBlock()

				// Flush and clear the disk mempool storage file following successful block commitment.
				saveMempoolToDisk([]*core.Transfer{})
			}
		}
	}()

	// Output operational status diagnostics for the running background node daemon.
	fmt.Printf("[NODE] Active validator miner: %s\n", addrMiner)
	fmt.Println("[NODE] Node operational and listening. Press Ctrl+C to terminate.")
	
	// Block the main execution thread indefinitely to maintain background daemon activity.
	select {}
}

// handleManualMine executes a single manual iteration of the Proof-of-Work block mining procedure using accumulated mempool records.
func handleManualMine() {
	fmt.Println("[CLI] Triggering Manual Block Mining...")
	
	// Load the default miner wallet configuration, falling back to a system identifier if unavailable.
	addrMiner, _, _, err := wallet.LoadWalletCustom("eterbit_data/keystore.json")
	if err != nil {
		addrMiner = "SYSTEM_MINER"
	}

	// Initialize the ledger environment associated with manual block construction.
	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	
	// Fetch unconfirmed transaction objects from physical disk storage.
	diskMempool := loadMempoolFromDisk()
	if len(diskMempool) > 0 {
		// Acquire mutex locking primitives to safely inject transactions into memory structures.
		ledger.Mu.Lock()
		ledger.Mempool = diskMempool
		ledger.Mu.Unlock()
	}

	// Perform the manual mining cycle to seal transactions into a new cryptographic block.
	ledger.MineBlock()
	
	// Purge the local disk mempool file after successful block sealing.
	saveMempoolToDisk([]*core.Transfer{})
}

// handleExploreBlockchain parses and displays structural blockchain blocks and metadata directly from physical storage.
func handleExploreBlockchain() {
	// Invoke core storage inspection functions to visualize the existing blockchain database layout.
	core.InspectBlockchain("eterbit_data")
}
