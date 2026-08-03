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

package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"eterbit/core"
)

// MessageType defines the classification of network messages transmitted across nodes.
type MessageType string

const (
	MsgTx    MessageType = "TX"
	MsgBlock MessageType = "BLOCK"
)

// Envelope wraps network payloads with a specific type header for transmission parsing.
type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Server represents the P2P networking node manager.
type Server struct {
	Address string
	Peers   map[string]net.Conn
	Mu      sync.Mutex
}

// NewServer initializes a new P2P network server instance.
func NewServer(address string) *Server {
	return &Server{
		Address: address,
		Peers:   make(map[string]net.Conn),
	}
}

// StartListening binds to the specified TCP address and listens for incoming node connections.
func (s *Server) StartListening(onBlockReceived func(*core.LedgerBlock), onTxReceived func(*core.Transfer)) error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("[P2P] Node networking server listening on %s...\n", s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn, onBlockReceived, onTxReceived)
	}
}

// handleConnection processes incoming message streams from established peer connections.
func (s *Server) handleConnection(conn net.Conn, onBlock func(*core.LedgerBlock), onTx func(*core.Transfer)) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	for {
		var env Envelope
		if err := decoder.Decode(&env); err != nil {
			break
		}

		switch env.Type {
		case MsgTx:
			var tx core.Transfer
			if json.Unmarshal(env.Payload, &tx) == nil && onTx != nil {
				onTx(&tx)
			}
		case MsgBlock:
			var block core.LedgerBlock
			if json.Unmarshal(env.Payload, &block) == nil && onBlock != nil {
				onBlock(&block)
			}
		}
	}
}

// Broadcast distributes transaction or block payloads to all connected network peers.
func (s *Server) Broadcast(msgType MessageType, data interface{}) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	envBytes, err := json.Marshal(Envelope{
		Type:    msgType,
		Payload: payload,
	})
	if err != nil {
		return
	}

	for addr, conn := range s.Peers {
		_, err := conn.Write(append(envBytes, '\n'))
		if err != nil {
			conn.Close()
			delete(s.Peers, addr)
		}
	}
}

// ConnectToPeer establishes an outbound TCP connection to another active network peer.
func (s *Server) ConnectToPeer(peerAddr string) error {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return err
	}

	s.Mu.Lock()
	s.Peers[peerAddr] = conn
	s.Mu.Unlock()

	fmt.Printf("[P2P] Successfully connected to remote peer: %s\n", peerAddr)
	return nil
}
