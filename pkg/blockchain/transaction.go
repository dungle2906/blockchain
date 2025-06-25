package blockchain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
)

type Transaction struct {
	Sender    []byte  `json:"sender"`
	Receiver  []byte  `json:"receiver"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	Signature []byte  `json:"signature"`
}

// Hash tạo hash SHA256 của giao dịch
func (t *Transaction) Hash() []byte {
	// Copy transaction nhưng không có Signature
	txCopy := *t
	txCopy.Signature = nil
	data, _ := json.Marshal(txCopy)

	hash := sha256.Sum256(data)
	return hash[:]
}

// SignTransaction Ký transaction bằng private key
func SignTransaction(t *Transaction, privKey *ecdsa.PrivateKey) error {
	hash := t.Hash()
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	// Ghép R + S
	t.Signature = append(r.Bytes(), s.Bytes()...)
	return nil
}

// VerifyTransaction Kiểm tra signature
func VerifyTransaction(t *Transaction, pubKey *ecdsa.PublicKey) bool {
	hash := t.Hash()
	// Tách R, S
	signLen := len(t.Signature) / 2
	r := new(big.Int).SetBytes(t.Signature[:signLen])
	s := new(big.Int).SetBytes(t.Signature[signLen:])
	return ecdsa.Verify(pubKey, hash, r, s)
}
