package main

import (
	"encoding/hex"
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

// main initializes the command-line interface multiplexer, parses command arguments,
// and routes execution flow to the appropriate handler function based on user input.
func main() {
	// Initialize distinct FlagSets for modular CLI subcommand routing
	walletCreateCmd := flag.NewFlagSet("create", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	explorerCmd := flag.NewFlagSet("explorer", flag.ExitOnError)

	// Bind flags configuration for the wallet creation command
	walletName := walletCreateCmd.String("name", "keystore.json", "Custom filename for the wallet (e.g., wallet2.json)")

	// Bind flags configuration for the transfer execution command
	sendRecipient := sendCmd.String("to", "", "Recipient destination address (etrb...)")
	sendAmount := sendCmd.Uint64("amount", 0, "Transfer value amount denominated in base units")
	sendFee := sendCmd.Uint64("fee", 5, "Transaction execution gas/network fee allocation")
	walletSource := sendCmd.String("wallet", "keystore.json", "Wallet filename to use for sending transaction")

	// Validate presence of operational arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Switch execution context based on the primary subcommand token
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
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage outputs the standard command reference menu and available operational flags
// to the standard output terminal interface.
func printUsage() {
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT BLOCKCHAIN CLI MANAGER (MULTI-WALLET)")
	fmt.Println("================================================================================")
	fmt.Println("Available commands:")
	fmt.Println("  go run eterbit.go create -name <file.json>   - Generate a new custom wallet file")
	fmt.Println("  go run eterbit.go balance                    - Inspect account states and balances in LevelDB")
	fmt.Println("  go run eterbit.go send -to <addr> -amount <val> -wallet <file> - Broadcast transfer tx")
	fmt.Println("  go run eterbit.go node                       - Initialize validator miner and live mempool daemon")
	fmt.Println("  go run eterbit.go explorer                   - Inspect blockchain blocks and transaction ledger")
	fmt.Println("================================================================================")
}

// handleCreateWallet provisions a new cryptographic keypair and serializes it to a custom filename.
func handleCreateWallet(filename string) {
	os.MkdirAll("eterbit_data", 0755)
	filePath := filepath.Join("eterbit_data", filename)

	// Check if wallet file already exists
	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("[WALLET] Wallet file '%s' already exists in eterbit_data/!\n", filename)
		return
	}

	// Trigger generation of new asymmetric PQC parameters using custom path helper
	addr, privKey, pubBytes, err := wallet.CreateOrLoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[WALLET] Failed to generate cryptographic keys: %v\n", err)
		return
	}

	privBytes, err := privKey.MarshalBinary()
	if err != nil {
		fmt.Printf("[WALLET] Failed to marshal private key: %v\n", err)
		return
	}

	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT NEW WALLET GENERATED & PERSISTED SUCCESSFULLY")
	fmt.Println("================================================================================")
	fmt.Printf(" Filename         : %s\n", filePath)
	fmt.Printf(" Address          : %s\n", addr)
	fmt.Printf(" Public Key (Hex) : %s\n", hex.EncodeToString(pubBytes))
	fmt.Printf(" Private Key (Hex): %s\n", hex.EncodeToString(privBytes))
	fmt.Println("--------------------------------------------------------------------------------")
}

// handleCheckBalance initializes the persistence engine state ledger interface
// and inspects all recorded account states and nonces currently committed.
func handleCheckBalance() {
	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_VIEWER")
	fmt.Println("================================================================================")
	fmt.Println(" REGISTERED ACCOUNT BALANCES IN LEVELDB")
	fmt.Println("================================================================================")
	
	if len(ledger.State) == 0 {
		fmt.Println(" No accounts currently recorded in the state ledger.")
		return
	}
	
	for addr, acc := range ledger.State {
		formattedAddr := addr
		if len(addr) >= 16 && addr[:4] != "etrb" {
			formattedAddr = "etrb" + addr
		}
		fmt.Printf(" Address: %s... | Balance: %d Coins | Nonce: %d\n", formattedAddr[:16], acc.Balance, acc.Nonce)
	}
	fmt.Println("================================================================================")
}

// handleSendTx constructs, signs via post-quantum cryptography, and broadcasts
// a state-transition transaction payload into the local node mempool queue using a specific wallet.
func handleSendTx(recipient string, amount uint64, fee uint64, walletFile string) {
	if recipient == "" || amount == 0 {
		fmt.Println("[CLI] Incomplete arguments! Please utilize the -to and -amount flags.")
		fmt.Println("Example: go run eterbit.go send -to <address> -amount 100 -wallet wallet2.json")
		return
	}

	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_SENDER")
	
	filePath := filepath.Join("eterbit_data", walletFile)
	addrA, privKeyA, pubBytesA, err := wallet.LoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[CLI] Failed to load wallet from %s: %v\n", filePath, err)
		return
	}

	if _, exists := ledger.State[addrA]; !exists {
		ledger.State[addrA] = &node.AccountState{Balance: 1000, Nonce: 0}
	}

	currentNonce := ledger.State[addrA].Nonce
	fmt.Printf("[CLI] Constructing transaction from %s (via %s) to %s...\n", addrA[:16], walletFile, recipient)

	tx := core.NewTransfer(privKeyA, pubBytesA, recipient, amount, fee, currentNonce)
	
	if ledger.AddToMempool(tx) {
		fmt.Printf("[CLI] Transaction successfully committed to mempool! ID: %s\n", tx.ComputeID()[:16])
	}
}

// handleRunNode boots up the persistent consensus worker daemon thread,
// loading validator parameters and activating automated block generation cycles.
func handleRunNode() {
	fmt.Println("[SYS] Booting Eterbit Live Node with LevelDB Storage Engine...")
	fmt.Println("--------------------------------------------------------------------------------")

	addrMiner, _, pubBytesMiner, err := wallet.LoadWalletCustom("eterbit_data/keystore.json")
	if err != nil {
		fmt.Printf("[NODE] Failed to load default miner wallet (keystore.json): %v\n", err)
		return
	}

	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	ledger.StartLiveWorker(4 * time.Second)

	fmt.Printf("[NODE] Active validator miner address: %s (Pub: %s...)\n", addrMiner[:16], hex.EncodeToString(pubBytesMiner[:8]))
	fmt.Println("[NODE] Node operational daemon running. Press Ctrl+C to terminate process.")
	fmt.Println("--------------------------------------------------------------------------------")

	select {}
}

// handleExploreBlockchain invokes the core explorer routine to display stored block details.
func handleExploreBlockchain() {
	core.InspectBlockchain("eterbit_data")
}
