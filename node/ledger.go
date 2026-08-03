package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"eterbit/core"
	"eterbit/internal/consensus"
	"eterbit/storage"
)

// AccountState represents the account balance and transaction sequence nonce.
type AccountState struct {
	Balance uint64 `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}

// LedgerCore manages the blockchain chain state, mempool transaction queue, and block validation engine.
type LedgerCore struct {
	Mu           sync.RWMutex
	Chain        []*core.LedgerBlock
	State        map[string]*AccountState
	Mempool      []*core.Transfer
	Engine       *core.ConsensusEngine
	MinerAddress string
	Storage      *storage.Database
	StopSignal   chan bool
}

// InitializeLedger initializes or loads the local ledger database state from the specified storage path.
func InitializeLedger(dbPath string, initialDifficulty uint32, minerAddr string) *LedgerCore {
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}

	coreLedger := &LedgerCore{
		Chain:        make([]*core.LedgerBlock, 0),
		State:        make(map[string]*AccountState),
		Mempool:      make([]*core.Transfer, 0),
		Engine:       core.NewConsensusEngine(initialDifficulty),
		MinerAddress: minerAddr,
		Storage:      db,
		StopSignal:   make(chan bool),
	}

	if !coreLedger.LoadFromDisk() {
		fmt.Println("[DB] Database is empty. Spawning Genesis Block...")
		coreLedger.SpawnGenesis()
	} else {
		fmt.Println("[DB] Blockchain successfully loaded from LevelDB storage!")
	}

	return coreLedger
}

// LoadFromDisk loads existing blockchain blocks from LevelDB disk storage and rebuilds the account state.
func (lc *LedgerCore) LoadFromDisk() bool {
	lastIdx, exists := lc.Storage.GetLastIndex()
	if !exists {
		return false
	}

	for i := uint64(0); i <= lastIdx; i++ {
		data, err := lc.Storage.GetBlock(i)
		if err != nil {
			break
		}
		var block core.LedgerBlock
		if err := json.Unmarshal(data, &block); err == nil {
			lc.Chain = append(lc.Chain, &block)
			lc.RebuildState(&block)
		}
	}
	return len(lc.Chain) > 0
}

// RebuildState updates the account balances and nonces based on the transactions within a block.
func (lc *LedgerCore) RebuildState(block *core.LedgerBlock) {
	for _, tx := range block.Transfers {
		sender := hex.EncodeToString(tx.SenderPubKey[:16])
		if _, ok := lc.State[sender]; !ok {
			lc.State[sender] = &AccountState{Balance: 10000, Nonce: 0}
		}
		
		// Validasi aman dari underflow uint64
		if lc.State[sender].Balance >= (tx.Value + tx.Fee) {
			lc.State[sender].Balance -= (tx.Value + tx.Fee)
		} else {
			lc.State[sender].Balance = 0
		}
		lc.State[sender].Nonce++

		if _, ok := lc.State[tx.Recipient]; !ok {
			lc.State[tx.Recipient] = &AccountState{Balance: 0, Nonce: 0}
		}
		lc.State[tx.Recipient].Balance += tx.Value
	}

	if block.Miner != "SYSTEM_GENESIS" && block.Miner != "" {
		var feeTotal uint64 = 0
		for _, tx := range block.Transfers {
			feeTotal += tx.Fee
		}
		
		totalRewardAdded := block.Reward + feeTotal
		if totalRewardAdded > 0 {
			if _, ok := lc.State[block.Miner]; !ok {
				lc.State[block.Miner] = &AccountState{Balance: 0, Nonce: 0}
			}
			lc.State[block.Miner].Balance += totalRewardAdded
		}
	}
}

// SpawnGenesis creates and persists the initial genesis block of the blockchain network.
func (lc *LedgerCore) SpawnGenesis() {
	genesis := &core.LedgerBlock{
		Index:      0,
		Timestamp:  time.Now().Unix(),
		PrevHash:   make([]byte, 32),
		Transfers:  []*core.Transfer{},
		Miner:      "SYSTEM_GENESIS",
		Nonce:      0,
		Difficulty: lc.Engine.TargetDifficulty,
	}
	_, genesis.Hash = lc.Engine.Mine(genesis)
	lc.Chain = append(lc.Chain, genesis)
	lc.Storage.SaveBlock(0, genesis)
}

// AddToMempool validates and inserts a transaction payload into the pending mempool queue.
func (lc *LedgerCore) AddToMempool(tx *core.Transfer) bool {
	lc.Mu.Lock()
	defer lc.Mu.Unlock()

	if !tx.Verify() {
		fmt.Println("[MEMPOOL] Invalid transaction cryptographic signature!")
		return false
	}

	sender := hex.EncodeToString(tx.SenderPubKey[:16])
	acc, exists := lc.State[sender]
	
	if !exists {
		lc.State[sender] = &AccountState{Balance: 10000, Nonce: tx.Nonce}
		acc = lc.State[sender]
	} else if acc.Balance < (tx.Value + tx.Fee) {
		acc.Balance = 10000
	}

	if tx.Nonce != acc.Nonce {
		acc.Nonce = tx.Nonce
	}

	lc.Mempool = append(lc.Mempool, tx)
	fmt.Printf("[MEMPOOL] Transaction successfully queued (ID: %s...)\n", tx.ComputeID()[:12])
	return true
}

// StartLiveWorker starts the background worker daemon to periodically mine blocks from pending mempool transactions or empty blocks.
func (lc *LedgerCore) StartLiveWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(lc.Mempool) > 0 {
					lc.MineBlock()
				}
			case <-lc.StopSignal:
				ticker.Stop()
				return
			}
		}
	}()
}

// MineBlock packages mempool transactions, executes proof-of-work mining, and appends the new block to the ledger.
func (lc *LedgerCore) MineBlock() {
	lc.Mu.Lock()
	parent := lc.Chain[len(lc.Chain)-1]
	validTx := make([]*core.Transfer, 0)
	var feeTotal uint64 = 0

	if len(lc.Mempool) > 0 {
		for _, tx := range lc.Mempool {
			sender := hex.EncodeToString(tx.SenderPubKey[:16])
			
			if _, ok := lc.State[sender]; !ok {
				lc.State[sender] = &AccountState{Balance: 10000, Nonce: 0}
			}
			acc := lc.State[sender]

			if acc.Balance >= (tx.Value + tx.Fee) {
				acc.Balance -= (tx.Value + tx.Fee)
			} else {
				acc.Balance = 0
			}
			acc.Nonce++

			if _, ok := lc.State[tx.Recipient]; !ok {
				lc.State[tx.Recipient] = &AccountState{Balance: tx.Value, Nonce: 0}
			} else {
				lc.State[tx.Recipient].Balance += tx.Value
			}

			feeTotal += tx.Fee
			validTx = append(validTx, tx)
			fmt.Printf("[MINER] -> Force Processed Tx: %d Coins to %s\n", tx.Value, tx.Recipient)
		}
		lc.Mempool = make([]*core.Transfer, 0)
	}
	lc.Mu.Unlock()

	newBlock := &core.LedgerBlock{
		Index:      parent.Index + 1,
		Timestamp:  time.Now().Unix(),
		PrevHash:   parent.Hash,
		Transfers:  validTx,
		Miner:      lc.MinerAddress,
		Difficulty: lc.Engine.TargetDifficulty,
	}

	fmt.Printf("[MINER] Mining Block #%d with %d transactions (Difficulty: %d)...\n", newBlock.Index, len(validTx), newBlock.Difficulty)
	
	startTime := time.Now()
	nonce, hash := lc.Engine.Mine(newBlock)
	duration := time.Since(startTime)

	newBlock.Nonce = nonce
	newBlock.Hash = hash

	lc.Mu.Lock()
	lc.Chain = append(lc.Chain, newBlock)
	lc.Storage.SaveBlock(newBlock.Index, newBlock)

	totalMinerReward := newBlock.Reward + feeTotal
	if totalMinerReward > 0 {
		if _, ok := lc.State[lc.MinerAddress]; !ok {
			lc.State[lc.MinerAddress] = &AccountState{Balance: totalMinerReward, Nonce: 0}
		} else {
			lc.State[lc.MinerAddress].Balance += totalMinerReward
		}
	}
	lc.Mu.Unlock()

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("[SUCCESS] Block #%d Mined & Saved! (Reward: %d, Fee: %d, Nonce: %d, Time: %v)\n", newBlock.Index, newBlock.Reward, feeTotal, newBlock.Nonce, duration)
	fmt.Printf("[CHAIN] Total Blocks: %d | Transactions Processed: %d\n", len(lc.Chain), len(validTx))
	fmt.Println("--------------------------------------------------------------------------------")
}
