package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"

	"eterbit/crypto"
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
// otherwise, it provisions a new custom cryptographic keypair and persists it.
func CreateOrLoadWallet() (string, *ecdsa.PrivateKey, error) {
	existing, err := LoadWallet()
	if err == nil {
		privBytes, _ := hex.DecodeString(existing.PrivateKey)
		pubBytes, _ := hex.DecodeString(existing.PublicKey)

		x := new(big.Int).SetBytes(pubBytes[:len(pubBytes)/2])
		y := new(big.Int).SetBytes(pubBytes[len(pubBytes)/2:])

		priv := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: elliptic.P256(),
				X:     x,
				Y:     y,
			},
			D: new(big.Int).SetBytes(privBytes),
		}

		return existing.Address, priv, nil
	}

	priv, err := crypto.GenerateKey()
	if err != nil {
		return "", nil, err
	}

	addr := crypto.PubkeyToAddress(&priv.PublicKey)
	pubBytes := elliptic.Marshal(priv.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	pubHex := hex.EncodeToString(pubBytes)
	privHex := hex.EncodeToString(priv.D.Bytes())

	SaveWallet(addr, pubHex, privHex)

	return addr, priv, nil
}
