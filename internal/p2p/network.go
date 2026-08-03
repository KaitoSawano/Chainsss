package p2p

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	
	"eterbit/core"
)

type MessageType string

const (
	MsgTx    MessageType = "TX"
	MsgBlock MessageType = "BLOCK"
)

type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Server struct {
	Address string
	Peers   map[string]net.Conn
	Mu      sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
		Peers:   make(map[string]net.Conn),
	}
}

// StartListening membuka port TCP agar node bisa menerima koneksi dari node lain
func (s *Server, onBlockReceived func(*core.LedgerBlock), onTxReceived func(*core.Transfer)) error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("[P2P] Node listening on %s...\n", s.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn, onBlockReceived, onTxReceived)
	}
}

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

// Broadcast mengirim data (transaksi atau blok) ke semua peer yang terhubung
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

// ConnectToPeer menyambungkan node ini ke node peer lain
func (s *Server) ConnectToPeer(peerAddr string) error {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return err
	}

	s.Mu.Lock()
	s.Peers[peerAddr] = conn
	s.Mu.Unlock()

	fmt.Printf("[P2P] Successfully connected to peer: %s\n", peerAddr)
	return nil
}
