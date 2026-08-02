package core

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"golang.org/x/crypto/sha3"
)

// Definisikan batas maksimum total koin yang dapat beredar (Max Supply)
const MaxEterbitSupply uint64 = 5000 // 5.000 Koin sesuai perhitungan halving

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
	Reward     uint64      `json:"reward"`
}

// GetBlockReward menghitung reward per blok secara dinamis berdasarkan tinggi rantai (halving mechanism)
func GetBlockReward(blockHeight uint64) uint64 {
	initialReward := uint64(50)  // Reward awal: 50 Eterbit per blok
	halvingInterval := uint64(50) // Interval halving

	halvings := blockHeight / halvingInterval

	if halvings >= 64 {
		return 0
	}

	return initialReward >> halvings
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
func (ce *ConsensusEngine) Mine(b *LedgerBlock) (uint64, []byte, error) {
	b.Reward = GetBlockReward(b.Index)

	// Pastikan total reward tidak melampaui Max Supply (opsional: bisa diakumulasikan dari blok sebelumnya)
	// Jika reward pada blok ini bernilai 0 karena halving sudah habis, penambangan tetap bisa lanjut (transaksi fee saja jika ada)
	if b.Reward == 0 && b.Index > (64 * 50) {
		// Batas supply maksimal tercapai sepenuhnya
	}

	var nonce uint64 = 0
	hasher := sha3.New256()

	for {
		data := ce.AssembleBlockData(b, nonce)
		hasher.Reset()
		hasher.Write(data)
		hash := hasher.Sum(nil)

		if ce.validateHash(hash) {
			return nonce, hash, nil
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
