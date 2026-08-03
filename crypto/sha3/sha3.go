// Copyright (c) 2026 AldianOkto. All rights reserved.
// Copyright (c) 2026 Eterbit Core.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core

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
