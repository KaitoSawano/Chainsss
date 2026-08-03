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

package wallet

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"eterbit/crypto"
	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// WalletData defines the structural schema for serializing cryptographic keypairs and address mappings.
type WalletData struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

const walletDir = "eterbit_data"

// SaveWalletCustom serializes wallet credential parameters and commits them to a custom file path.
func SaveWalletCustom(filePath, addr, pubHex, privHex string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	w := WalletData{
		Address:    addr,
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

// LoadWalletCustom reads and deserializes a keystore mapping from a specific custom file path.
func LoadWalletCustom(filePath string) (string, *mode3.PrivateKey, []byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, nil, err
	}

	var w WalletData
	if err := json.Unmarshal(data, &w); err != nil {
		return "", nil, nil, err
	}

	privBytes, _ := hex.DecodeString(w.PrivateKey)
	pubBytes, _ := hex.DecodeString(w.PublicKey)

	var priv mode3.PrivateKey
	if err := priv.UnmarshalBinary(privBytes); err != nil {
		return "", nil, nil, err
	}

	return w.Address, &priv, pubBytes, nil
}

// SaveWallet serializes the wallet credential parameters to the default keystore file.
func SaveWallet(addr, pubHex, privHex string) error {
	return SaveWalletCustom(filepath.Join(walletDir, "keystore.json"), addr, pubHex, privHex)
}

// LoadWallet attempts to read and deserialize the default localized keystore mapping from disk storage.
func LoadWallet() (*WalletData, error) {
	filePath := filepath.Join(walletDir, "keystore.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var w WalletData
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}

	return &w, nil
}

// CreateOrLoadWallet checks for an existing default keystore file. If found, it loads and decodes the keys;
// otherwise, it provisions a new post-quantum Dilithium keypair and persists it.
func CreateOrLoadWallet() (string, *mode3.PrivateKey, []byte, error) {
	return CreateOrLoadWalletCustom(filepath.Join(walletDir, "keystore.json"))
}

// CreateOrLoadWalletCustom provisions a new post-quantum Dilithium keypair and persists it to a custom file path.
func CreateOrLoadWalletCustom(filePath string) (string, *mode3.PrivateKey, []byte, error) {
	pub, priv, err := crypto.GenerateKey()
	if err != nil {
		return "", nil, nil, err
	}

	pubBytes, err := pub.MarshalBinary()
	if err != nil {
		return "", nil, nil, err
	}
	
	privBytes, err := priv.MarshalBinary()
	if err != nil {
		return "", nil, nil, err
	}

	addr := crypto.PubkeyToAddress(pubBytes)
	pubHex := hex.EncodeToString(pubBytes)
	privHex := hex.EncodeToString(privBytes)

	_ = SaveWalletCustom(filePath, addr, pubHex, privHex)

	return addr, priv, pubBytes, nil
}
