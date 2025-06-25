package p2p

import (
	"context"
	"log"
	"time"

	pb "go_blockchain/proto/go_blockchain/proto"

	"google.golang.org/grpc"
)

// Gửi đề xuất block từ Leader đến một peer
func SendBlockProposal(peerAddress string, block *pb.Block) (*pb.Vote, error) {
	conn, err := grpc.Dial(peerAddress, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	if err != nil {
		log.Printf("❌ Không kết nối được peer %s: %v", peerAddress, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.ProposeBlock(ctx, &pb.BlockProposal{Block: block})
	if err != nil {
		log.Printf("❌ Lỗi gửi block đến %s: %v", peerAddress, err)
		return nil, err
	}

	log.Printf("📩 Nhận vote từ %s: %v", peerAddress, resp.Accepted)
	return resp, nil
}

// Gửi vote từ Follower về Leader
func SendVote(peerAddress string, vote *pb.Vote) (*pb.Response, error) {
	conn, err := grpc.Dial(peerAddress, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	if err != nil {
		log.Printf("❌ Không kết nối được peer %s: %v", peerAddress, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := client.SubmitVote(ctx, vote)
	if err != nil {
		log.Printf("❌ Lỗi gửi vote đến %s: %v", peerAddress, err)
		return nil, err
	}

	log.Printf("📨 Leader phản hồi vote: %s", resp.Message)
	return resp, nil
}
