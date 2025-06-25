package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	pb "go_blockchain/proto/go_blockchain/proto" // Sửa đúng import
	// "fmt"
	"time"
)

type Block struct {
	Transactions      []Transaction `json:"transactions"`
	MerkleRoot        []byte        `json:"merkle_root"`
	PreviousBlockHash []byte        `json:"previous_block_hash"`
	Timestamp         int64         `json:"timestamp"`
	CurrentBlockHash  []byte        `json:"current_block_hash"`
}

func (b *Block) Hash() []byte {
	blockCopy := *b
	blockCopy.CurrentBlockHash = nil
	data, _ := json.Marshal(blockCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}

// NewBlock tạo block mới
func NewBlock(transactions []Transaction, prevHash []byte) *Block {
	block := &Block{
		Transactions:      transactions,
		PreviousBlockHash: prevHash,
		Timestamp:         time.Now().Unix(),
	}
	block.MerkleRoot = block.CalculateMerkleRoot()
	block.CurrentBlockHash = block.Hash()
	return block
}

func (b *Block) CalculateMerkleRoot() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	return merkleRoot(txHashes)
}

// Hàm tạo MerkleRoot từ mảng hash
func merkleRoot(data [][]byte) []byte {
	if len(data) == 1 {
		return data[0]
	}
	var newLevel [][]byte
	for i := 0; i < len(data); i += 2 {
		var hash []byte
		if i+1 < len(data) {
			hash = hashPair(data[i], data[i+1])
		} else {
			hash = hashPair(data[i], data[i]) // nếu lẻ thì duplicate phần tử cuối
		}
		newLevel = append(newLevel, hash)
	}
	return merkleRoot(newLevel)
}

func hashPair(a, b []byte) []byte {
	h := sha256.New()
	h.Write(a)
	h.Write(b)
	return h.Sum(nil)
}

func ComputeMerkleRoot(txs []*pb.Transaction) []byte {
	// Ví dụ đơn giản: concat all tx hashes rồi sha256
	var combined []byte
	for _, tx := range txs {
		txData, _ := json.Marshal(tx)
		combined = append(combined, txData...)
	}
	hash := sha256.Sum256(combined)
	return hash[:]
}

func HashBlock(b *pb.Block) []byte {
	// Bỏ field hash ra nếu có
	bCopy := *b
	bCopy.Hash = nil

	data, _ := json.Marshal(bCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}
