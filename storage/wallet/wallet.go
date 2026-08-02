package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

type WalletData struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const walletDir = "storage/wallet"
const walletFile = "keystore.json"

// SaveWallet menyimpan dompet ke file JSON lokal
func SaveWallet(addr, pubHex, privHex string) error {
	if err := os.MkdirAll(walletDir, 0755); err != nil {
		return err
	}

	w := WalletData{
		Address:    addr,
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(walletDir, walletFile), data, 0600)
}

// LoadWallet memuat dompet yang tersimpan di lokal
func LoadWallet() (*WalletData, error) {
	filePath := filepath.Join(walletDir, walletFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var w WalletData
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}

	return &w, nil
}

// CreateOrLoadWallet membuat dompet baru jika belum ada, atau memuat yang lama
func CreateOrLoadWallet() (string, mode3.PublicKey, mode3.PrivateKey, error) {
	existing, err := LoadWallet()
	if err == nil {
		pubBytes, _ := hex.DecodeString(existing.PublicKey)
		privBytes, _ := hex.DecodeString(existing.PrivateKey)

		var pubKey mode3.PublicKey
		var privKey mode3.PrivateKey
		copy(pubKey.Bytes(), pubBytes)
		privKey.UnmarshalBinary(privBytes)

		return existing.Address, pubKey, privKey, nil
	}

	pubKey, privKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return "", mode3.PublicKey{}, mode3.PrivateKey{}, err
	}

	pubBytes := pubKey.Bytes()
	rawHex := hex.EncodeToString(pubBytes[:14])
	addr := "etrb" + rawHex
	pubHex := hex.EncodeToString(pubBytes)
	privBytes, _ := privKey.MarshalBinary()
	privHex := hex.EncodeToString(privBytes)

	SaveWallet(addr, pubHex, privHex)

	return addr, *pubKey, *privKey, nil
}
