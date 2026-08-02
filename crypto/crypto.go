package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// GenerateKey generates a new ECDSA private key utilizing the secp256k1 equivalent elliptic curve standard.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Keccak256 computes the 256-bit Keccak hash of the provided byte slice data payload.
func Keccak256(data []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// PubkeyToAddress derives a standard hex-encoded account address string from an ECDSA public key.
func PubkeyToAddress(p *ecdsa.PublicKey) string {
	pubBytes := elliptic.Marshal(p.Curve, p.X, p.Y)
	hash := Keccak256(pubBytes[1:])
	return "etrb" + hex.EncodeToString(hash[12:])
}

// Sign creates an ECDSA digital signature over the provided 32-byte message hash payload.
func Sign(hash []byte, priv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, errors.New("invalid hash length, must be exactly 32 bytes")
	}
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash)
	if err != nil {
		return nil, err
	}
	sig := make([]byte, 64)
	copy(sig[32-len(r.Bytes()):32], r.Bytes())
	copy(sig[64-len(s.Bytes()):64], s.Bytes())
	return sig, nil
}

// Verify checks the validity of an ECDSA signature against a public key and message hash.
func Verify(pubKey *ecdsa.PublicKey, hash, sig []byte) bool {
	if len(sig) != 64 || len(hash) != 32 {
		return false
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return ecdsa.Verify(pubKey, hash, r, s)
}
