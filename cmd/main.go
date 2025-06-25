package main

import (
	"context"
	"fmt"
	"go_blockchain/pkg/blockchain"
	"go_blockchain/pkg/p2p"
	"go_blockchain/pkg/storage"
	pb "go_blockchain/proto/go_blockchain/proto" // ✅ sửa đúng import
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
)

// BlockchainServer implements the gRPC BlockchainService
type BlockchainServer struct {
	pb.UnimplementedBlockchainServiceServer
	NodeID        string
	IsLeader      bool
	Peers         []string
	PendingTxs    []*pb.Transaction // Bộ nhớ tạm các giao dịch chưa xử lý
	VotesReceived map[string]int
	PendingBlocks map[string]*pb.Block
}

func main() {
	// 🔧 Đọc cấu hình từ biến môi trường
	nodeID := os.Getenv("NODE_ID")
	isLeader := os.Getenv("IS_LEADER") == "true"
	peers := splitPeers(os.Getenv("PEERS"))

	if nodeID == "" {
		log.Fatal("NODE_ID is not set")
	}

	fmt.Printf("🚀 Node %s starting... IsLeader=%v\n", nodeID, isLeader)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051" // mặc định nếu không set
	}
	lis, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &BlockchainServer{
		NodeID:        nodeID,
		IsLeader:      isLeader,
		Peers:         peers,
		PendingTxs:    []*pb.Transaction{},
		VotesReceived: make(map[string]int),
		PendingBlocks: make(map[string]*pb.Block),
	}

	pb.RegisterBlockchainServiceServer(grpcServer, server)

	if isLeader {
		go server.StartLeaderLoop()
	}

	fmt.Printf("🟢 gRPC server listening on :%s\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}

// Tách chuỗi PEERS thành danh sách peer address
func splitPeers(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// 📥 RPC: Nhận giao dịch
func (s *BlockchainServer) SendTransaction(ctx context.Context, tx *pb.Transaction) (*pb.Response, error) {
	log.Printf("💰 Transaction received: From %x To %x Amount %.2f\n", tx.Sender, tx.Receiver, tx.Amount)

	// TODO: Validate chữ ký ở đây

	s.PendingTxs = append(s.PendingTxs, tx)
	log.Printf("📥 Current pending txs: %d\n", len(s.PendingTxs))

	return &pb.Response{Success: true, Message: "Transaction received"}, nil
}

// ⏱️ Leader kiểm tra định kỳ → tạo block → gửi follower
func (s *BlockchainServer) StartLeaderLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		if len(s.PendingTxs) == 0 {
			log.Println("⏳ No pending transactions. Skipping block creation.")
			continue
		}

		log.Println("📦 Creating block from pending transactions...")

		// Tạo block (tạm thời không có hash và MerkleRoot)
		block := &pb.Block{
			Timestamp:    time.Now().Unix(),
			Transactions: s.PendingTxs,
			// TODO: tính MerkleRoot, Hash nếu cần
		}

		// ✅ Tính Merkle Root
		block.MerkleRoot = blockchain.ComputeMerkleRoot(block.Transactions)

		// ✅ Tính Hash block
		blockHash := blockchain.HashBlock(block)

		s.PendingBlocks[string(blockHash)] = block
		s.VotesReceived[string(blockHash)] = 1 // leader tự vote đồng ý

		// Reset giao dịch đã dùng
		s.PendingTxs = nil

		// Gửi block đến followers
		for _, peer := range s.Peers {
			go func(p string) {
				vote, err := p2p.SendBlockProposal(p, block)
				if err != nil {
					log.Printf("❌ Failed to get vote from %s: %v", p, err)
					return
				}
				log.Printf("🗳️ Vote from %s: %v", p, vote.Accepted)

				// TODO: Nếu đủ phiếu → commit block
			}(peer)
		}
	}
}

func (s *BlockchainServer) ProposeBlock(ctx context.Context, proposal *pb.BlockProposal) (*pb.Vote, error) {
	block := proposal.Block

	// ✅ Kiểm tra đơn giản
	if block == nil || len(block.Transactions) == 0 {
		log.Println("❌ Block invalid or empty, rejecting.")
		return &pb.Vote{
			Accepted:  false,
			BlockHash: nil,
		}, nil
	}

	log.Printf("📥 Received block proposal with %d transactions at %d", len(block.Transactions), block.Timestamp)

	// TODO: kiểm tra chữ ký giao dịch, MerkleRoot hợp lệ...

	// 🔐 Tạm thời giả lập hash
	blockHash := []byte("dummy-hash") // sau này bạn dùng hàm hash block thật

	// ✅ Vote
	vote := &pb.Vote{
		Accepted:  true,
		BlockHash: blockHash,
	}

	// Gửi vote về leader
	go func() {
		for _, peer := range s.Peers {
			if !s.IsLeader {
				_, err := p2p.SendVote(peer, vote)
				if err != nil {
					log.Printf("❌ Failed to send vote to leader %s: %v", peer, err)
				} else {
					log.Printf("📨 Sent vote to leader %s", peer)
				}
			}
		}
	}()

	// Trả về vote (gRPC trả về)
	return vote, nil
}

func (s *BlockchainServer) SubmitVote(ctx context.Context, vote *pb.Vote) (*pb.Response, error) {
	if !s.IsLeader {
		return &pb.Response{Success: false, Message: "Not leader"}, nil
	}

	blockKey := string(vote.BlockHash)

	// Nếu chưa có vote/block này → bỏ qua
	if _, ok := s.VotesReceived[blockKey]; !ok {
		return &pb.Response{Success: false, Message: "Block not tracked"}, nil
	}

	// Nếu vote bị từ chối → cũng bỏ qua
	if !vote.Accepted {
		log.Printf("❌ Vote từ chối cho block %x", vote.BlockHash)
		return &pb.Response{Success: true, Message: "Vote rejected"}, nil
	}

	s.VotesReceived[blockKey]++

	log.Printf("🗳️ Tổng phiếu đồng ý cho block %x: %d", vote.BlockHash, s.VotesReceived[blockKey])

	// Nếu đủ phiếu → commit block
	totalVotes := s.VotesReceived[blockKey]
	required := (len(s.Peers) + 1) * 2 / 3 // 2/3 tổng số node (kể cả leader)

	if totalVotes >= required {
		block := s.PendingBlocks[blockKey]
		// ✅ Lưu block vào LevelDB
		db, err := storage.NewBlockStorage("./data")
		if err != nil {
			log.Printf("❌ Không mở được LevelDB: %v", err)
			return &pb.Response{Success: false, Message: "DB error"}, nil
		}

		err = db.SaveBlock(block)
		if err != nil {
			log.Printf("❌ Lưu block thất bại: %v", err)
			return &pb.Response{Success: false, Message: "Save failed"}, nil
		}

		log.Printf("✅ Block %x đã lưu vào LevelDB", vote.BlockHash)
	}

	return &pb.Response{Success: true, Message: "Vote counted"}, nil
}
