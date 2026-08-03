// Copyright (c) 2026 AldianOkto. All rights reserved.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core

package core

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"golang.org/x/crypto/sha3"
	"eterbit/internal/consensus"
)

// CoinUnit mendefinisikan skala 8 desimal yang konsisten dengan node
const CoinUnit = uint64(100000000)

// Mengambil batas maksimum total koin dari parameter konsensus terpusat
var consensusParams = consensus.DefaultConsensus()
const MaxEterbitSupply uint64 = 785000000 // Menyesuaikan dengan MaxSupply di internal/consensus

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

// GetBlockReward menghitung reward per blok secara dinamis dan dikalikan CoinUnit agar akurat 8 desimal
func GetBlockReward(blockHeight uint64) uint64 {
	// Mengalikan BlockReward dengan CoinUnit agar menjadi satuan terkecil (misal: 50 * 100,000,000)
	initialReward := consensusParams.BlockReward * CoinUnit
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
func (ce *ConsensusEngine) Mine(b *LedgerBlock) (uint64, []byte) {
	b.Reward = GetBlockReward(b.Index)

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

// validateHash memvalidasi hash menggunakan aturan validasi dari internal/consensus
func (ce *ConsensusEngine) validateHash(hash []byte) bool {
	hashStr := hex.EncodeToString(hash)
	// Memanggil fungsi ValidatePoW dari package internal/consensus secara langsung
	return consensus.ValidatePoW(hashStr, uint64(ce.TargetDifficulty))
}
