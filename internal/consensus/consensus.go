package consensus

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

// ConsensusParameters defines the fixed macroeconomic and mathematical rules for the Eterbit blockchain.
type ConsensusParameters struct {
	DifficultyBits uint64 // Target difficulty prefix/bits for Proof-of-Work
	BlockReward    uint64 // Initial mining reward per block (e.g., 50 coins)
	MaxSupply      uint64 // Maximum cap for token issuance
}

// DefaultConsensus returns the standard operational consensus rules for Eterbit.
func DefaultConsensus() *ConsensusParameters {
	return &ConsensusParameters{
		DifficultyBits: 3,
		BlockReward:    50,
		MaxSupply:      285000000, // Menyesuaikan dengan parameter koin Termin/Eterbit
	}
}

// ValidatePoW verifies whether a given block header hash satisfies the target difficulty requirement.
func ValidatePoW(blockHash string, difficulty uint64) bool {
	target := createTargetPrefix(difficulty)
	return len(blockHash) >= int(difficulty) && blockHash[:int(difficulty)] == target
}

// ComputeHeaderHash calculates the cryptographic SHA-256 hash representation for block validation.
func ComputeHeaderHash(prevHash string, merkleRoot string, timestamp int64, nonce uint64) string {
	record := bytes.Join([][]byte{
		[]byte(prevHash),
		[]byte(merkleRoot),
		big.NewInt(timestamp).Bytes(),
		big.NewInt(int64(nonce)).Bytes(),
	}, []byte{})

	hash := sha256.Sum256(record)
	return hex.EncodeToString(hash[:])
}

// createTargetPrefix generates the required leading zero pattern based on the difficulty level.
func createTargetPrefix(difficulty uint64) string {
	prefix := ""
	for i := uint64(0); i < difficulty; i++ {
		prefix += "0"
	}
	return prefix
}

// VerifyBlockReward checks if the distributed block reward and transaction fees adhere to protocol limits.
func VerifyBlockReward(rewardClaimed uint64, feesCollected uint64, standardReward uint64) bool {
	return rewardClaimed <= (standardReward + feesCollected)
}
