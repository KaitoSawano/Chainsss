// Copyright (c) 2026 AldianOkto. All rights reserved.
// Copyright (c) 2026 Eterbit Core.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core
//
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at. <http://www.apache.org/licenses/LICENSE-2.0>
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"eterbit/internal/consensus"
)

// BlockRecord represents the structural schema of a stored block for exploration purposes.
type BlockRecord struct {
	Index        int64      `json:"index"`
	Timestamp    int64      `json:"timestamp"`
	PrevHash     string     `json:"prev_hash"`
	Hash         string     `json:"hash"`
	Validator    string     `json:"miner"` // Menyesuaikan dengan field miner di LedgerBlock
	Transactions []Transfer `json:"transfers"`
	Difficulty   uint32     `json:"difficulty"`
	Nonce        uint64     `json:"nonce"`
	Reward       uint64     `json:"reward"`
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

			// Menggunakan fungsi toDecimal atau pembagian langsung dengan CoinUnit yang ada di package core
			rewardDecimal := float64(block.Reward) / float64(CoinUnit)

			fmt.Printf("📦 Block #[%d]\n", block.Index)
			fmt.Printf("    Hash         : %s\n", block.Hash)
			fmt.Printf("    Prev Hash    : %s\n", block.PrevHash)
			fmt.Printf("    Validator    : %s\n", block.Validator)
			fmt.Printf("    Difficulty   : %d\n", block.Difficulty)
			fmt.Printf("    Nonce        : %d\n", block.Nonce)
			fmt.Printf("    Reward       : %.8f Coins\n", rewardDecimal)
			fmt.Printf("    Tx Count     : %d transactions\n", len(block.Transactions))
			fmt.Println("--------------------------------------------------------------------------------")
			for idx, tx := range block.Transactions {
				txID := tx.ComputeID()
				if len(txID) > 8 {
					txID = txID[:8]
				}
				valDecimal := float64(tx.Value) / float64(CoinUnit)
				fmt.Printf("    └─ Tx #%d ID : %s | To: %s... | Value: %.8f Coins\n", 
					idx, txID, tx.Recipient[:16], valDecimal)
			}
			fmt.Println("================================================================================")
			count++
		}
	}

	if count == 0 {
		fmt.Println("[EXPLORER] No structured block records found in the database.")
	}
}
