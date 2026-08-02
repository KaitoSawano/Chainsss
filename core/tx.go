package core

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
	"golang.org/x/crypto/sha3"
)

type Transfer struct {
	SenderPubKey []byte `json:"sender_pub_key"`
	Recipient    string `json:"recipient"`
	Value        uint64 `json:"value"`
	Fee          uint64 `json:"fee"`
	Nonce        uint64 `json:"nonce"`
	Signature    []byte `json:"signature"`
}

func NewTransfer(priv *mode3.PrivateKey, pub []byte, recipient string, value, fee, nonce uint64) *Transfer {
	tx := &Transfer{
		SenderPubKey: pub,
		Recipient:    recipient,
		Value:        value,
		Fee:          fee,
		Nonce:        nonce,
	}
	
	sig, err := priv.Sign(rand.Reader, tx.PayloadBytes(), crypto.Hash(0))
	if err != nil {
		panic(err)
	}
	tx.Signature = sig
	return tx
}

func (tx *Transfer) PayloadBytes() []byte {
	return []byte(fmt.Sprintf("%s-%d-%d-%d", tx.Recipient, tx.Value, tx.Fee, tx.Nonce))
}

func (tx *Transfer) Verify() bool {
	var pub mode3.PublicKey
	if err := pub.UnmarshalBinary(tx.SenderPubKey); err != nil {
		return false
	}
	return mode3.Verify(&pub, tx.PayloadBytes(), tx.Signature)
}

func (tx *Transfer) ComputeID() string {
	hasher := sha3.New256()
	hasher.Write(append(tx.SenderPubKey, tx.PayloadBytes()...))
	hasher.Write(tx.Signature)
	return hex.EncodeToString(hasher.Sum(nil))
}
