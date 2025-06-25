package storage

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"go_blockchain/pkg/blockchain"
	pb "go_blockchain/proto/go_blockchain/proto" // Sửa đúng import
)

type BlockStorage struct {
	db *leveldb.DB
}

func NewBlockStorage(dbPath string) (*BlockStorage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}
	return &BlockStorage{db: db}, nil
}

func (s *BlockStorage) SaveBlock(b *pb.Block) error {
	data, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}
	// key := b.MerkleRoot
	return s.db.Put(b.Hash, data, nil)
}

func (s *BlockStorage) GetBlock(hash []byte) (*blockchain.Block, error) {
	data, err := s.db.Get(hash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	var block blockchain.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	return &block, nil
}
