package core

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

type Transfer struct {
	SenderPubKey []byte `json:"sender_pub_key"`
	Recipient    string `json:"recipient"`
	Value        uint64 `json:"value"`
	Fee          uint64 `json:"fee"`
	Nonce        uint64 `json:"nonce"`
	SignatureR   *big.Int `json:"signature_r"`
	SignatureS   *big.Int `json:"signature_s"`
}

func (tx *Transfer) ComputeID() string {
	data := append(tx.SenderPubKey, []byte(tx.Recipient)...)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (tx *Transfer) Sign(privKey *ecdsa.PrivateKey) error {
	hash := sha256.Sum256([]byte(tx.ComputeID()))
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return err
	}
	tx.SignatureR = r
	tx.SignatureS = s
	return nil
}

func (tx *Transfer) Verify() bool {
	if tx.SignatureR == nil || tx.SignatureS == nil {
		return false
	}
	// Rekonstruksi public key dari bytes
	x := new(big.Int).SetBytes(tx.SenderPubKey[:len(tx.SenderPubKey)/2])
	y := new(big.Int).SetBytes(tx.SenderPubKey[len(tx.SenderPubKey)/2:])
	pubKey := ecdsa.PublicKey{Curve: privKeyCurve(), X: x, Y: y}

	hash := sha256.Sum256([]byte(tx.ComputeID()))
	return ecdsa.Verify(&pubKey, hash[:], tx.SignatureR, tx.SignatureS)
}

func privKeyCurve() elliptic.Curve {
	// Sesuaikan dengan kurva yang digunakan di wallet.go Anda (misal elliptic.P256())
	return elliptic.P256()
}
