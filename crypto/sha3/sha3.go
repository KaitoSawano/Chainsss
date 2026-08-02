package sha3

import (
	"golang.org/x/crypto/sha3"
)

// Hash256 computes a standard 256-bit SHA3 hash of the given data payload.
func Hash256(data []byte) []byte {
	hasher := sha3.New256()
	hasher.Write(data)
	return hasher.Sum(nil)
}
