// Copyright (c) 2026 AldianOkto. All rights reserved.
// Use of this source code is governed by the Apache License.
// that can be found in the root directory of this repository.
// Project: Eterbit / Blockchain Core

package dilithium3

import (
	"crypto"
	"crypto/rand"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

// GenerateKey provisions a new post-quantum cryptographic keypair using Dilithium mode3.
func GenerateKey() (*mode3.PublicKey, *mode3.PrivateKey, error) {
	pub, priv, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// Sign generates a digital signature over a message hash using a Dilithium private key.
func Sign(priv *mode3.PrivateKey, message []byte) ([]byte, error) {
	return priv.Sign(rand.Reader, message, crypto.Hash(0))
}

// Verify checks the cryptographic validity of a signature against a public key and message.
func Verify(pub *mode3.PublicKey, message, sig []byte) bool {
	return mode3.Verify(pub, message, sig)
}
