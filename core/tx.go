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
	"encoding/hex"
	"fmt"

	"eterbit/crypto"
	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// Transfer represents a state-transition transaction payload containing sender public key, recipient address, transfer values, nonce, and post-quantum signature.
type Transfer struct {
	SenderPubKey []byte `json:"sender_pub_key"`
	Recipient    string `json:"recipient"`
	Value        uint64 `json:"value"`
	Fee          uint64 `json:"fee"`
	Nonce        uint64 `json:"nonce"`
	Signature    []byte `json:"signature"`
}

// NewTransfer constructs a new Transfer transaction instance, computes its hash payload, and signs it using the Dilithium private key.
func NewTransfer(priv *mode3.PrivateKey, pub []byte, recipient string, value, fee, nonce uint64) *Transfer {
	tx := &Transfer{
		SenderPubKey: pub,
		Recipient:    recipient,
		Value:        value,
		Fee:          fee,
		Nonce:        nonce,
	}

	sig, err := crypto.Sign(priv, tx.PayloadBytes())
	if err != nil {
		panic(err)
	}
	tx.Signature = sig
	return tx
}

// PayloadBytes serializes the primary transactional parameters into a canonical byte slice representation for signing and verification.
func (tx *Transfer) PayloadBytes() []byte {
	return []byte(fmt.Sprintf("%s-%d-%d-%d", tx.Recipient, tx.Value, tx.Fee, tx.Nonce))
}

// Verify validates the post-quantum digital signature attached to the transaction against the sender's public key.
func (tx *Transfer) Verify() bool {
	var pub mode3.PublicKey
	if err := pub.UnmarshalBinary(tx.SenderPubKey); err != nil {
		return false
	}
	return crypto.Verify(&pub, tx.PayloadBytes(), tx.Signature)
}

// ComputeID generates a unique cryptographic hash identifier string for the transaction instance.
func (tx *Transfer) ComputeID() string {
	hasher := crypto.Hash256(append(tx.SenderPubKey, tx.PayloadBytes()...))
	return hex.EncodeToString(hasher)
}
