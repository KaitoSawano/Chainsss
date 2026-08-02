package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"time"

	"eterbit/core"
	"eterbit/node"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

func main() {
	// Definisikan argumen CLI
	walletCreateCmd := flag.NewFlagSet("create", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	nodeCmd := flag.NewFlagSet("node", flag.ExitOnError)

	// Argumen untuk perintah send
	sendRecipient := sendCmd.String("to", "", "Alamat penerima (16 karakter hex pertama)")
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
	fmt.Println("  go run eterbit.go create            - Buat dompet & kunci baru")
	fmt.Println("  go run eterbit.go balance           - Cek status/saldo akun di state")
	fmt.Println("  go run eterbit.go send -to <addr> -amount <val> -fee <val> - Kirim koin")
	fmt.Println("  go run eterbit.go node              - Jalankan penambang & live mempool node")
	fmt.Println("================================================================================")
}

func handleCreateWallet() {
	pubKey, privKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("[WALLET] Gagal membuat kunci: %v\n", err)
		return
	}

	pubBytes := pubKey.Bytes()
	addr := hex.EncodeToString(pubBytes[:16])
	
	privBytes, _ := privKey.MarshalBinary()

	fmt.Println("================================================================================")
	fmt.Println(" 🔑 DOMPET BARU BERHASIL DIBUAT")
	fmt.Println("================================================================================")
	fmt.Printf(" Alamat (Address) : 0x%s\n", addr)
	fmt.Printf(" Public Key (Hex) : %s\n", hex.EncodeToString(pubBytes))
	fmt.Printf(" Private Key (Hex): %s\n", hex.EncodeToString(privBytes))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println(" ⚠️  Simpan Private Key Anda dengan aman!")
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
		fmt.Printf(" Alamat: 0x%s... | Saldo: %d Koin | Nonce: %d\n", addr[:12], acc.Balance, acc.Nonce)
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
	
	// Untuk demo CLI, kita buat dompet pengirim sementara yang otomatis diberi saldo
	pubKeyA, privKeyA, _ := mode3.GenerateKey(rand.Reader)
	pubBytesA := pubKeyA.Bytes()
	addrA := hex.EncodeToString(pubBytesA[:16])

	// Beri saldo awal ke dompet demo ini jika belum ada
	if _, exists := ledger.State[addrA]; !exists {
		ledger.State[addrA] = &node.AccountState{Balance: 1000, Nonce: 0}
	}

	currentNonce := ledger.State[addrA].Nonce
	fmt.Printf("[CLI] Membuat transaksi dari 0x%s ke 0x%s...\n", addrA[:12], recipient)

	tx := core.NewTransfer(privKeyA, pubBytesA, recipient, amount, fee, currentNonce)
	
	// Masukkan ke mempool (bisa diproses node nanti)
	if ledger.AddToMempool(tx) {
		fmt.Printf("[CLI] ✅ Transaksi berhasil dimasukkan ke mempool! ID: %s\n", tx.ComputeID()[:16])
	}
}

func handleRunNode() {
	fmt.Println("[SYS] Memulai Live Node Eterbit dengan LevelDB Storage...")
	fmt.Println("--------------------------------------------------------------------------------")

	pubKeyM, _, _ := mode3.GenerateKey(rand.Reader)
	addrMiner := hex.EncodeToString(pubKeyM.Bytes()[:16])

	ledger := node.InitializeLedger("eterbit_data", 3, addrMiner)
	ledger.StartLiveWorker(4 * time.Second)

	fmt.Printf("[NODE] Penambang aktif dengan alamat: 0x%s\n", addrMiner[:12])
	fmt.Println("[NODE] Node berjalan. Tekan Ctrl+C untuk menghentikan.")
	fmt.Println("--------------------------------------------------------------------------------")

	select {}
}
