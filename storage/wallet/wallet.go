package wallet

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"eterbit/crypto"
	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// WalletData defines the structural schema for serializing cryptographic keypairs and address mappings.
type WalletData struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const walletDir = "eterbit_data"
const walletFile = "keystore.json"

// SaveWallet serializes the wallet credential parameters and commits them to a local JSON keystore file.
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

// LoadWallet attempts to read and deserialize the localized keystore mapping from disk storage.
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

// CreateOrLoadWallet checks for an existing keystore file. If found, it loads and decodes the keys;
// otherwise, it provisions a new post-quantum Dilithium keypair and persists it.
func CreateOrLoadWallet() (string, *mode3.PrivateKey, []byte, error) {
	existing, err := LoadWallet()
	if err == nil {
		privBytes, _ := hex.DecodeString(existing.PrivateKey)
		pubBytes, _ := hex.DecodeString(existing.PublicKey)

		var priv mode3.PrivateKey
		if err := priv.UnmarshalBinary(privBytes); err != nil {
			return "", nil, nil, err
		}

		return existing.Address, &priv, pubBytes, nil
	}

	pub, priv, err := crypto.GenerateKey()
	if err != nil {
		return "", nil, nil, err
	}

	pubBytes, err := pub.MarshalBinary()
	if err != nil {
		return "", nil, nil, err
	}
	
	privBytes, err := priv.MarshalBinary()
	if err != nil {
		return "", nil, nil, err
	}

	addr := crypto.PubkeyToAddress(pubBytes)
	pubHex := hex.EncodeToString(pubBytes)
	privHex := hex.EncodeToString(privBytes)

	SaveWallet(addr, pubHex, privHex)

	return addr, priv, pubBytes, nil
}
