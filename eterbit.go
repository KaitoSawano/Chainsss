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

func main() {
	// Definisikan argumen CLI
	walletCreateCmd := flag.NewFlagSet("create", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)

	// Argumen untuk perintah send
	sendRecipient := sendCmd.String("to", "", "Alamat penerima (etrb...)")
	sendAmount := sendCmd.Uint64("amount", 0, "Jumlah koin yang dikirim")
	sendFee := sendCmd.Uint64("fee", 5, "Biaya transaksi (fee)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

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

func printUsage() {
	fmt.Println("================================================================================")
	fmt.Println(" 🚀 ETERBIT BLOCKCHAIN CLI MANAGER")
	fmt.Println("================================================================================")
	fmt.Println("Gunakan perintah berikut:")
	fmt.Println("  go run eterbit.go create            - Buat atau muat dompet lokal")
	fmt.Println("  go run eterbit.go balance           - Cek status/saldo akun di state")
	fmt.Println("  go run eterbit.go send -to <addr> -amount <val> -fee <val> - Kirim koin")
	fmt.Println("  go run eterbit.go node              - Jalankan penambang & live mempool node")
	fmt.Println("================================================================================")
}

func handleCreateWallet() {
	existing, err := wallet.LoadWallet()
	if err == nil {
		fmt.Println("================================================================================")
		fmt.Println(" 📁 DOMPET LOKAL SUDAH ADA (KESTORE)")
		fmt.Println("================================================================================")
		fmt.Printf(" Alamat (Address) : %s\n", existing.Address)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println(" Dompet Anda aman tersimpan di storage/wallet/keystore.json")
		return
	}

	addr, pubKey, privKey, err := wallet.CreateOrLoadWallet()
	if err != nil {
		fmt.Printf("[WALLET] Gagal membuat kunci: %v\n", err)
		return
	}

	pubBytes := pubKey.Bytes()
	privBytes, _ := privKey.MarshalBinary()

	fmt.Println("================================================================================")
	fmt.Println(" 🔑 DOMPET KUSTOM ETERBIT BERHASIL DIBUAT & DISIMPAN")
	fmt.Println("================================================================================")
	fmt.Printf(" Alamat (Address) : %s\n", addr)
	fmt.Printf(" Public Key (Hex) : %s\n", hex.EncodeToString(pubBytes))
	fmt.Printf(" Private Key (Hex): %s\n", hex.EncodeToString(privBytes))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println(" ⚠️  Disimpan otomatis ke storage/wallet/keystore.json")
}

func handleCheckBalance() {
	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_VIEWER")
	fmt.Println("================================================================================")
	fmt.Println(" 📊 SALDO AKUN TERDAFTAR DI LEVELDB")
	fmt.Println("================================================================================")
	if len(ledger.State) == 0 {
		fmt.Println(" Belum ada akun yang tercatat di state ledger.")
		return
	}
	for addr, acc := range ledger.State {
		formattedAddr := addr
		if len(addr) >= 16 && addr[:4] != "etrb" {
			formattedAddr = "etrb" + addr
		}
		fmt.Printf(" Alamat: %s... | Saldo: %d Koin | Nonce: %d\n", formattedAddr[:16], acc.Balance, acc.Nonce)
	}
	fmt.Println("================================================================================")
}

func handleSendTx(recipient string, amount uint64, fee uint64) {
	if recipient == "" || amount == 0 {
		fmt.Println("[CLI] ❌ Argumen tidak lengkap! Gunakan flag -to dan -amount.")
		fmt.Println("Contoh: go run eterbit.go send -to <alamat> -amount 100")
		return
	}

	ledger := node.InitializeLedger("eterbit_data", 3, "SYSTEM_SENDER")
	
	// Muat dompet lokal pengirim
	addrA, pubKeyA, privKeyA, err := wallet.CreateOrLoadWallet()
	if err != nil {
		fmt.Printf("[CLI] Gagal memuat dompet lokal: %v\n", err)
		return
	}

	pubBytesA := pubKeyA.Bytes()

	if _, exists := ledger.State[addrA]; !exists {
		ledger.State[addrA] = &node.AccountState{Balance: 1000, Nonce: 0}
	}

	currentNonce := ledger.State[addrA].Nonce
	fmt.Printf("[CLI] Membuat transaksi dari %s ke %s...\n", addrA[:16], recipient)

	// Perbaikan di sini: menambahkan & sebelum privKeyA agar sesuai tipe pointer
	tx := core.NewTransfer(&privKeyA, pubBytesA, recipient, amount, fee, currentNonce)
	
	if ledger.AddToMempool(tx) {
		fmt.Printf("[CLI] ✅ Transaksi berhasil dimasukkan ke mempool! ID: %s\n", tx.ComputeID()[:16])
	}
}

func handleRunNode() {
	fmt.Println("[SYS] Memulai Live Node Eterbit dengan LevelDB Storage...")
	fmt.Println("--------------------------------------------------------------------------------")

	// Penambang menggunakan dompet lokal agar hadiah blok masuk ke akun Anda sendiri
	addrMiner, pubKeyMiner, _, err := wallet.CreateOrLoadWallet()
	if err != nil {
		fmt.Printf("[NODE] Gagal memuat dompet penambang: %v\n", err)
		return
	}

	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	ledger.StartLiveWorker(4 * time.Second)

	fmt.Printf("[NODE] Penambang aktif dengan alamat: %s (Pub: %s...)\n", addrMiner[:16], hex.EncodeToString(pubKeyMiner.Bytes()[:8]))
	fmt.Println("[NODE] Node berjalan. Tekan Ctrl+C untuk menghentikan.")
	fmt.Println("--------------------------------------------------------------------------------")

	select {}
}
