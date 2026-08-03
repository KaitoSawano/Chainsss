// Copyright (c) 2026 AldianOkto. All rights reserved.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core
package storage

import (
	"encoding/json"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

// Database represents a persistent LevelDB storage engine wrapper instance.
type Database struct {
	DB *leveldb.DB
}

// NewDatabase initializes and opens a LevelDB instance at the specified directory path.
func NewDatabase(dbPath string) (*Database, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

// SaveBlock serializes a block payload and commits it to disk storage, updating the latest block index pointer.
func (d *Database) SaveBlock(index uint64, blockData interface{}) error {
	data, err := json.Marshal(blockData)
	if err != nil {
		return err
	}
	key := "block_" + strconv.FormatUint(index, 10)
	err = d.DB.Put([]byte(key), data, nil)
	if err != nil {
		return err
	}
	return d.DB.Put([]byte("last_index"), []byte(strconv.FormatUint(index, 10)), nil)
}

// GetLastIndex retrieves the highest committed block height index from the underlying database.
func (d *Database) GetLastIndex() (uint64, bool) {
	data, err := d.DB.Get([]byte("last_index"), nil)
	if err != nil {
		return 0, false
	}
	idx, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, false
	}
	return idx, true
}

// GetBlock queries and returns the serialized byte payload of a specific block by its height index.
func (d *Database) GetBlock(index uint64) ([]byte, error) {
	key := "block_" + strconv.FormatUint(index, 10)
	return d.DB.Get([]byte(key), nil)
}
