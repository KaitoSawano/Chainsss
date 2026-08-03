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

func main() {
	walletCreateCmd := flag.NewFlagSet("create", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	explorerCmd := flag.NewFlagSet("explorer", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)

	walletName := walletCreateCmd.String("name", "keystore.json", "Custom filename for the wallet")

	sendRecipient := sendCmd.String("to", "", "Recipient destination address")
	sendAmount := sendCmd.Uint64("amount", 0, "Transfer value amount")
	sendFee := sendCmd.Uint64("fee", 2, "Transaction fee")
	walletSource := sendCmd.String("wallet", "keystore.json", "Wallet filename to use")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

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

func printUsage() {
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT BLOCKCHAIN CLI MANAGER (MULTI-WALLET)")
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

func handleCreateWallet(filename string) {
	os.MkdirAll("eterbit_data", 0755)
	filePath := filepath.Join("eterbit_data", filename)

	if _, err := os.Stat(filePath); err == nil {
		fmt.Printf("[WALLET] Wallet file '%s' already exists!\n", filename)
		return
	}

	addr, privKey, pubBytes, err := wallet.CreateOrLoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[WALLET] Failed: %v\n", err)
		return
	}

	pubHex := hex.EncodeToString(pubBytes)
	privBytes, _ := privKey.MarshalBinary()
	privHex := hex.EncodeToString(privBytes)

	wallet.SaveWalletCustom(filePath, addr, pubHex, privHex)

	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT NEW WALLET CREATED")
	fmt.Println("================================================================================")
	fmt.Printf(" Filename : %s\n", filePath)
	fmt.Printf(" Address  : %s\n", addr)
	fmt.Println("--------------------------------------------------------------------------------")
}

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
		fmt.Printf(" Address: %s | Balance: %d Coins | Nonce: %d\n", addr, acc.Balance, acc.Nonce)
	}
	fmt.Println("================================================================================")
}

func handleSendTx(recipient string, amount uint64, fee uint64, walletFile string) {
	if recipient == "" || amount == 0 {
		fmt.Println("[CLI] Incomplete arguments! Use -to and -amount.")
		return
	}

	if walletFile == "" {
		walletFile = "keystore.json"
	}

	filePath := filepath.Join("eterbit_data", walletFile)
	addrMiner, _, _, _ := wallet.LoadWalletCustom(filePath)
	
	// Inisialisasi ledger dengan miner address dari wallet pengirim agar state/reward sinkron
	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	
	addrA, privKeyA, pubBytesA, err := wallet.LoadWalletCustom(filePath)
	if err != nil {
		fmt.Printf("[CLI] Failed to load wallet from %s: %v\n", filePath, err)
		return
	}

	// Pastikan saldo awal akun cukup di state ledger
	ledger.State[addrA] = &node.AccountState{
		Balance: 10000,
		Nonce:   0,
	}
	
	if len(addrA) > 4 && addrA[:4] == "etrb" {
		ledger.State[addrA[4:]] = &node.AccountState{
			Balance: 10000,
			Nonce:   0,
		}
	} else {
		ledger.State["etrb"+addrA] = &node.AccountState{
			Balance: 10000,
			Nonce:   0,
		}
	}

	fmt.Printf("[CLI] Constructing transaction from %s (via %s) to %s (Amount: %d, Fee: %d)...\n", addrA, walletFile, recipient, amount, fee)

	tx := core.NewTransfer(privKeyA, pubBytesA, recipient, amount, fee, 0)
	
	if ledger.AddToMempool(tx) {
		fmt.Printf("[CLI] Transaction successfully committed to mempool! ID: %s\n", tx.ComputeID()[:16])
		
		// AUTO-MINE: Langsung panggil MineBlock di instance yang sama agar transaksi langsung masuk blok!
		fmt.Println("[CLI] Auto-mining block to process mempool transaction...")
		ledger.MineBlock()
	}
}

func handleRunNode() {
	fmt.Println("[SYS] Booting Eterbit Live Node...")
	addrMiner, _, _, err := wallet.LoadWalletCustom("eterbit_data/keystore.json")
	if err != nil {
		fmt.Printf("[NODE] Failed to load default miner wallet: %v\n", err)
		return
	}

	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	ledger.StartLiveWorker(4 * time.Second)

	fmt.Printf("[NODE] Active validator miner: %s\n", addrMiner)
	fmt.Println("[NODE] Node operational. Press Ctrl+C to terminate.")
	select {}
}

func handleManualMine() {
	fmt.Println("[CLI] Triggering Manual Block Mining...")
	addrMiner, _, _, err := wallet.LoadWalletCustom("eterbit_data/keystore.json")
	if err != nil {
		addrMiner = "SYSTEM_MINER"
	}

	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	ledger.MineBlock()
}

func handleExploreBlockchain() {
	core.InspectBlockchain("eterbit_data")
}
