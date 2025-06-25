package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go_blockchain/pkg/wallet"
	pb "go_blockchain/proto/go_blockchain/proto"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	// 1. Tạo key cho Alice và Bob
	alicePriv, _ := wallet.GenerateKeyPair()
	bobPriv, _ := wallet.GenerateKeyPair()

	alicePub := &alicePriv.PublicKey
	bobPub := &bobPriv.PublicKey

	// 2. Tạo giao dịch
	tx := &pb.Transaction{
		Sender:    wallet.PublicKeyToAddress(alicePub),
		Receiver:  wallet.PublicKeyToAddress(bobPub),
		Amount:    99.99,
		Timestamp: time.Now().Unix(),
	}

	// 3. Ký giao dịch bằng private key Alice
	tx.Signature = signTransaction(tx, alicePriv)

	// 4. Kết nối đến node qua gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Không thể kết nối gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)

	// 5. Gửi giao dịch
	res, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("❌ Gửi giao dịch thất bại: %v", err)
	}

	fmt.Printf("✅ Phản hồi từ node: %v\n", res.Message)
}

// signTransaction dùng để ký tx (tạm thời đơn giản hoá, đúng hơn nên hash rồi ký)
func signTransaction(tx *pb.Transaction, privKey *ecdsa.PrivateKey) []byte {
	hash := wallet.HashTransaction(tx)
	r, s, err := wallet.SignHash(hash, privKey)
	if err != nil {
		log.Fatalf("Ký thất bại: %v", err)
	}
	return append(r.Bytes(), s.Bytes()...)
}
