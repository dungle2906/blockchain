package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go_blockchain/proto/go_blockchain/proto"
	"math/big"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return privKey, nil
}

func PublicKeyToAddress(pubKey *ecdsa.PublicKey) []byte {
	// Hash public key bằng SHA256
	pubBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	hash := sha256.Sum256(pubBytes)
	return hash[:]
}

// Ký hash
func SignHash(hash []byte, privKey *ecdsa.PrivateKey) (*big.Int, *big.Int, error) {
	return ecdsa.Sign(rand.Reader, privKey, hash)
}

// Hash transaction
func HashTransaction(tx *proto.Transaction) []byte {
	txCopy := *tx
	txCopy.Signature = nil // không hash phần signature
	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}
