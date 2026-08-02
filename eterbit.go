package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
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

	// Bind flags configuration for the transfer execution command
	sendRecipient := sendCmd.String("to", "", "Recipient destination address (etrb...)")
	sendAmount := sendCmd.Uint64("amount", 0, "Transfer value amount denominated in base units")
	sendFee := sendCmd.Uint64("fee", 5, "Transaction execution gas/network fee allocation")

	// Validate presence of operational arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Switch execution context based on the primary subcommand token
	switch os.Args[1] {
	case "create":
		walletCreateCmd.Parse(os.Args[2:])
		handleCreateWallet()
	case "balance":
		balanceCmd.Parse(os.Args[2:])
		handleCheckBalance()
	case "send":
		sendCmd.Parse(os.Args[2:])
		handleSendTx(*sendRecipient, *sendAmount, *sendFee)
	case "node":
		nodeCmd.Parse(os.Args[2:])
		handleRunNode()
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage outputs the standard command reference menu and available operational flags
// to the standard output terminal interface.
func printUsage() {
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT BLOCKCHAIN CLI MANAGER")
	fmt.Println("================================================================================")
	fmt.Println("Available commands:")
	fmt.Println("  go run eterbit.go create             - Generate or load local cryptographic wallet")
	fmt.Println("  go run eterbit.go balance            - Inspect account states and balances in LevelDB")
	fmt.Println("  go run eterbit.go send -to <addr> -amount <val> -fee <val> - Broadcast transfer tx")
	fmt.Println("  go run eterbit.go node               - Initialize validator miner and live mempool daemon")
	fmt.Println("================================================================================")
}

// handleCreateWallet checks for the existence of a persistent keystore context.
// If absent, it provisions a cryptographically secure post-quantum keypair and serializes it.
func handleCreateWallet() {
	// Attempt to resolve pre-existing local keystore mapping
	existing, err := wallet.LoadWallet()
	if err == nil {
		fmt.Println("================================================================================")
		fmt.Println(" LOCAL WALLET ALREADY EXISTS (KEYSTORE)")
		fmt.Println("================================================================================")
		fmt.Printf(" Address         : %s\n", existing.Address)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println(" Wallet securely loaded from eterbit_data/keystore.json")
		return
	}

	// Trigger generation of new asymmetric PQC parameters
	addr, privKey, pubBytes, err := wallet.CreateOrLoadWallet()
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
	fmt.Println(" ETERBIT CUSTOM WALLET GENERATED & PERSISTED SUCCESSFULLY")
	fmt.Println("================================================================================")
	fmt.Printf(" Address         : %s\n", addr)
	fmt.Printf(" Public Key (Hex) : %s\n", hex.EncodeToString(pubBytes))
	fmt.Printf(" Private Key (Hex): %s\n", hex.EncodeToString(privBytes))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println(" Automatically persisted to eterbit_data/keystore.json")
}

// handleCheckBalance initializes the persistence engine state ledger interface
// and inspects all recorded account states and nonces currently committed.
func handleCheckBalance() {
	// Instantiate ledger reference mapped to the designated database storage directory
	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_VIEWER")
	fmt.Println("================================================================================")
	fmt.Println(" REGISTERED ACCOUNT BALANCES IN LEVELDB")
	fmt.Println("================================================================================")
	
	// Evaluate state ledger cardinality
	if len(ledger.State) == 0 {
		fmt.Println(" No accounts currently recorded in the state ledger.")
		return
	}
	
	// Iterate across active state keys to display balances
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
// a state-transition transaction payload into the local node mempool queue.
func handleSendTx(recipient string, amount uint64, fee uint64) {
	// Enforce strict parameter validation checks
	if recipient == "" || amount == 0 {
		fmt.Println("[CLI] Incomplete arguments! Please utilize the -to and -amount flags.")
		fmt.Println("Example: go run eterbit.go send -to <address> -amount 100")
		return
	}

	// Initialize transaction origin ledger subsystem context
	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_SENDER")
	
	// Load localized sender wallet credentials from disk storage
	addrA, privKeyA, pubBytesA, err := wallet.CreateOrLoadWallet()
	if err != nil {
		fmt.Printf("[CLI] Failed to load local wallet: %v\n", err)
		return
	}

	// Provision default genesis-like bootstrapping balance if sender state is uninitialized
	if _, exists := ledger.State[addrA]; !exists {
		ledger.State[addrA] = &node.AccountState{Balance: 1000, Nonce: 0}
	}

	// Retrieve active transactional nonce sequence value
	currentNonce := ledger.State[addrA].Nonce
	fmt.Printf("[CLI] Constructing transaction from %s to %s...\n", addrA[:16], recipient)

	// Execute cryptographic signing mechanism and formulate transfer payload object
	tx := core.NewTransfer(privKeyA, pubBytesA, recipient, amount, fee, currentNonce)
	
	// Commit signed transaction instance into the volatile mempool structure
	if ledger.AddToMempool(tx) {
		fmt.Printf("[CLI] Transaction successfully committed to mempool! ID: %s\n", tx.ComputeID()[:16])
	}
}

// handleRunNode boots up the persistent consensus worker daemon thread,
// loading validator parameters and activating automated block generation cycles.
func handleRunNode() {
	fmt.Println("[SYS] Booting Eterbit Live Node with LevelDB Storage Engine...")
	fmt.Println("--------------------------------------------------------------------------------")

	// Resolve local miner node operator identity credentials
	addrMiner, _, pubBytesMiner, err := wallet.CreateOrLoadWallet()
	if err != nil {
		fmt.Printf("[NODE] Failed to load miner wallet: %v\n", err)
		return
	}

	// Bind validator address context to the transactional ledger storage engine
	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	
	// Launch background block production worker routines with periodic intervals
	ledger.StartLiveWorker(4 * time.Second)

	fmt.Printf("[NODE] Active validator miner address: %s (Pub: %s...)\n", addrMiner[:16], hex.EncodeToString(pubBytesMiner[:8]))
	fmt.Println("[NODE] Node operational daemon running. Press Ctrl+C to terminate process.")
	fmt.Println("--------------------------------------------------------------------------------")

	// Block execution routine indefinitely to maintain persistent node uptime
	select {}
}
