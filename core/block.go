package core

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"golang.org/x/crypto/sha3"
)

// LedgerBlock represents the core structural block entity containing transactional ledger data, cryptographic hashes, and consensus metadata.
type LedgerBlock struct {
	Index      uint64      `json:"index"`
	Timestamp  int64       `json:"timestamp"`
	PrevHash   []byte      `json:"prev_hash"`
	Hash       []byte      `json:"hash"`
	Transfers  []*Transfer `json:"transfers"`
	Miner      string      `json:"miner"`
	Nonce      uint64      `json:"nonce"`
	Difficulty uint32      `json:"difficulty"`
}

// ConsensusEngine coordinates the proof-of-work mining process, hash target difficulty evaluation, and block validation parameters.
type ConsensusEngine struct {
	TargetDifficulty uint32
}

// NewConsensusEngine initializes and returns a new ConsensusEngine instance configured with the specified difficulty target.
func NewConsensusEngine(difficulty uint32) *ConsensusEngine {
	return &ConsensusEngine{TargetDifficulty: difficulty}
}

// AssembleBlockData serializes and concatenates block headers, transactional payloads, and a candidate nonce into a unified byte array for hashing.
func (ce *ConsensusEngine) AssembleBlockData(b *LedgerBlock, nonce uint64) []byte {
	var rawTxData []byte
	for _, tx := range b.Transfers {
		rawTxData = append(rawTxData, tx.Signature...)
	}
	return bytes.Join([][]byte{
		b.PrevHash,
		rawTxData,
		[]byte(strconv.FormatUint(b.Index, 16)),
		[]byte(strconv.FormatInt(b.Timestamp, 16)),
		[]byte(strconv.FormatUint(nonce, 16)),
	}, []byte{})
}

// Mine executes an iterative proof-of-work search loop, testing candidate nonces until a hash meeting the target difficulty is discovered.
func (ce *ConsensusEngine) Mine(b *LedgerBlock) (uint64, []byte) {
	var nonce uint64 = 0
	hasher := sha3.New256()

	for {
		data := ce.AssembleBlockData(b, nonce)
		hasher.Reset()
		hasher.Write(data)
		hash := hasher.Sum(nil)

		if ce.validateHash(hash) {
			return nonce, hash
		}
		nonce++
	}
}

// validateHash checks whether the given cryptographic hash satisfies the required leading zero-byte difficulty constraints.
func (ce *ConsensusEngine) validateHash(hash []byte) bool {
	hashStr := hex.EncodeToString(hash)
	for i := uint32(0); i < ce.TargetDifficulty; i++ {
		if hashStr[i] != '0' {
			return false
		}
	}
	return true
}
