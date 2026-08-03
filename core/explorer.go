package core

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"eterbit/internal/consensus"
)

// BlockRecord represents the structural schema of a stored block for exploration purposes.
type BlockRecord struct {
	Index      int64      `json:"index"`
	Timestamp  int64      `json:"timestamp"`
	PrevHash   string     `json:"prev_hash"`
	Hash       string     `json:"hash"`
	Validator  string     `json:"miner"` // Menyesuaikan dengan field miner di LedgerBlock
	Transactions []Transfer `json:"transfers"`
	Difficulty uint32     `json:"difficulty"`
	Nonce      uint64     `json:"nonce"`
	Reward     uint64     `json:"reward"`
}

// InspectBlockchain opens the LevelDB storage directly and inspects committed states and blocks.
func InspectBlockchain(dataDir string) {
	fmt.Println("================================================================================")
	fmt.Println(" ETERBIT BLOCKCHAIN EXPLORER (LEVELDB)")
	fmt.Println("================================================================================")

	// Open the LevelDB database instance from the data directory
	db, err := leveldb.OpenFile(dataDir, nil)
	if err != nil {
		fmt.Printf("[EXPLORER] Failed to open LevelDB database: %v\n", err)
		fmt.Println("[EXPLORER] Tip: Ensure the node has initialized the database storage.")
		fmt.Println("================================================================================")
		return
	}
	defer db.Close()

	// Iterate through database entries to extract and decode stored block records
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	// Referensi parameter konsensus jika dibutuhkan untuk validasi tambahan
	_ = consensus.DefaultConsensus()

	count := 0
	for iter.Next() {
		key := string(iter.Key())
		
		// Filter keys starting with "block_" to parse and display structured block details
		if len(key) >= 6 && key[:6] == "block_" {
			var block BlockRecord
			if err := json.Unmarshal(iter.Value(), &block); err != nil {
				// If it's raw bytes or another format, fallback to printing key size
				fmt.Printf("📦 Key: %s | Data Size: %d bytes\n", key, len(iter.Value()))
				continue
			}

			fmt.Printf("📦 Block #[%d]\n", block.Index)
			fmt.Printf("   Hash       : %s\n", block.Hash)
			fmt.Printf("   Prev Hash  : %s\n", block.PrevHash)
			fmt.Printf("   Validator  : %s\n", block.Validator)
			fmt.Printf("   Difficulty : %d\n", block.Difficulty)
			fmt.Printf("   Nonce      : %d\n", block.Nonce)
			fmt.Printf("   Reward     : %d Coins\n", block.Reward)
			fmt.Printf("   Tx Count   : %d transactions\n", len(block.Transactions))
			fmt.Println("--------------------------------------------------------------------------------")
			for idx, tx := range block.Transactions {
				txID := tx.ComputeID()
				if len(txID) > 8 {
					txID = txID[:8]
				}
				fmt.Printf("   └─ Tx #%d ID : %s | To: %s... | Value: %d Coins\n", 
					idx, txID, tx.Recipient[:16], tx.Value)
			}
			fmt.Println("================================================================================")
			count++
		}
	}

	if count == 0 {
		fmt.Println("[EXPLORER] No structured block records found in the database.")
	}
}
