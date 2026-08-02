package crypto

import (
	"encoding/hex"

	"eterbit/crypto/dilithium3"
	"eterbit/crypto/sha3"
	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// GenerateKey delegates post-quantum keypair creation to the dilithium3 module.
func GenerateKey() (*mode3.PublicKey, *mode3.PrivateKey, error) {
	return dilithium3.GenerateKey()
}

// Hash256 delegates hashing operations to the sha3 module.
func Hash256(data []byte) []byte {
	return sha3.Hash256(data)
}

// PubkeyToAddress derives a custom post-quantum network address string from a Dilithium public key bytes slice.
func PubkeyToAddress(pubBytes []byte) string {
	rawHex := hex.EncodeToString(pubBytes[:14])
	return "etrb" + rawHex
}

// Sign delegates signature generation to the dilithium3 module.
func Sign(priv *mode3.PrivateKey, message []byte) ([]byte, error) {
	return dilithium3.Sign(priv, message)
}

// Verify delegates signature verification to the dilithium3 module.
func Verify(pub *mode3.PublicKey, message, sig []byte) bool {
	return dilithium3.Verify(pub, message, sig)
}
