package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"

	"eterbit/crypto"
)

// Transfer represents a state-transition transaction payload containing sender public key, recipient address, transfer values, nonce, and signature.
type Transfer struct {
	SenderPubKey []byte `json:"sender_pub_key"`
	Recipient    string `json:"recipient"`
	Value        uint64 `json:"value"`
	Fee          uint64 `json:"fee"`
	Nonce        uint64 `json:"nonce"`
	Signature    []byte `json:"signature"`
}

// NewTransfer constructs a new Transfer transaction instance, computes its Keccak hash payload, and signs it using the ECDSA private key.
func NewTransfer(priv *ecdsa.PrivateKey, pub []byte, recipient string, value, fee, nonce uint64) *Transfer {
	tx := &Transfer{
		SenderPubKey: pub,
		Recipient:    recipient,
		Value:        value,
		Fee:          fee,
		Nonce:        nonce,
	}

	hash := crypto.Keccak256(tx.PayloadBytes())
	sig, err := crypto.Sign(hash, priv)
	if err != nil {
		panic(err)
	}
	tx.Signature = sig
	return tx
}

// PayloadBytes serializes the primary transactional parameters into a canonical byte slice representation for hashing.
func (tx *Transfer) PayloadBytes() []byte {
	return []byte(fmt.Sprintf("%s-%d-%d-%d", tx.Recipient, tx.Value, tx.Fee, tx.Nonce))
}

// Verify validates the digital signature attached to the transaction against the sender's public key using local crypto primitives.
func (tx *Transfer) Verify() bool {
	if len(tx.SenderPubKey) == 0 {
		return false
	}
	x := new(bigIntFromBytes(tx.SenderPubKey[:len(tx.SenderPubKey)/2])) // disesuaikan dengan parsing pubkey
	// Sederhanakan rekonstruksi public key untuk verifikasi
	pubCurve := elliptic.P256()
	pubX, pubY := elliptic.Unmarshal(pubCurve, tx.SenderPubKey)
	if pubX == nil {
		return false
	}
	pubKey := &ecdsa.PublicKey{Curve: pubCurve, X: pubX, Y: pubY}

	hash := crypto.Keccak256(tx.PayloadBytes())
	return crypto.Verify(pubKey, hash, tx.Signature)
}

// ComputeID generates a unique cryptographic hash identifier string for the transaction instance.
func (tx *Transfer) ComputeID() string {
	hasher := crypto.Keccak256(append(tx.SenderPubKey, tx.PayloadBytes()...))
	return hex.EncodeToString(hasher)
}
