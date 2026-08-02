package core

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"golang.org/x/crypto/sha3"
)

type LedgerBlock struct {
	Index        uint64      `json:"index"`
	Timestamp    int64       `json:"timestamp"`
	PrevHash     []byte      `json:"prev_hash"`
	Hash         []byte      `json:"hash"`
	Transfers    []*Transfer `json:"transfers"`
	Miner        string      `json:"miner"`
	Nonce        uint64      `json:"nonce"`
	Difficulty   uint32      `json:"difficulty"`
}

type ConsensusEngine struct {
	TargetDifficulty uint32
}

func NewConsensusEngine(difficulty uint32) *ConsensusEngine {
	return &ConsensusEngine{TargetDifficulty: difficulty}
}

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

func (ce *ConsensusEngine) validateHash(hash []byte) bool {
	hashStr := hex.EncodeToString(hash)
	for i := uint32(0); i < ce.TargetDifficulty; i++ {
		if hashStr[i] != '0' {
			return false
		}
	}
	return true
}
