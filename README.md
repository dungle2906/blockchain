# GO BLOCKCHAIN

A simple educational blockchain project written in Go with gRPC support, digital signatures, Merkle trees, block consensus, and LevelDB for persistence. Designed for learning distributed systems and cryptographic fundamentals.

---

## Table of Contents

* [Features](#features)
* [Architecture Decisions](#architecture-decisions)
* [Libraries Used](#libraries-used)
* [Folder Structure](#folder-structure)
* [Setup & Run with Docker](#setup--run-with-docker)
* [Manual Testing (Local)](#manual-testing-local)
* [Using the CLI](#using-the-cli)
* [How It Works](#how-it-works)
* [Future Improvements](#future-improvements)

---

## Features

* ECDSA keypair generation for each user (Alice, Bob)
* Digital signature of transactions with verification
* Block creation with Merkle Root
* Block propagation and voting between 3 nodes using gRPC
* Block commit if 2/3 nodes accept
* Data persistence via LevelDB
* Fully Dockerized deployment (3 blockchain nodes)

---

## Architecture Decisions

### Language: Go

* Chosen for its strong support of concurrency (goroutines, channels), simplicity, and built-in support for cryptography, protobuf, and networking.

### Communication: gRPC

* Enables strongly typed service interfaces, easy protobuf serialization, and bidirectional streaming if needed in future.

### Consensus: Simplified Leader-Follower Vote

* One leader proposes blocks.
* Two followers vote.
* Leader commits if >= 2/3 votes are positive.
* Keeps the protocol simple and understandable for learning.

### Data Store: LevelDB

* Embedded key-value store.
* Efficient and persistent.
* Simple to use without external dependencies.

---

## Libraries Used

| Library                               | Purpose                                 |
| ------------------------------------- | --------------------------------------- |
| `crypto/ecdsa`                        | Digital signature (transaction signing) |
| `crypto/sha256`                       | Hashing for Merkle Tree & block hash    |
| `google.golang.org/grpc`              | gRPC framework                          |
| `github.com/syndtr/goleveldb/leveldb` | Storage backend                         |
| `protobuf`                            | Schema definition for blocks, txs       |

---

## Folder Structure

```
go_blockchain/
├── cmd/
│   ├── main.go          # main server for nodes
│   └── cli/main.go     # client CLI to send tx
├── pkg/
│   ├── crypto/         # ECDSA, hash, signature
│   ├── p2p/            # gRPC block & vote forwarding
│   ├── blockchain/     # block structure, merkle
│   └── storage/        # LevelDB integration
├── proto/                   # blockchain.proto + generated code
├── Dockerfile               # for building Go app
├── docker-compose.yml       # run 3 nodes
└── README.md
```

---

## Setup & Run with Docker

### 1. Build & Start the system:

```bash
docker-compose up --build
```

* Node1: Leader, port 50051
* Node2: Follower, port 50052
* Node3: Follower, port 50053

### 2. Logs will show:

* Node start
* Transactions received
* Block creation every 10 seconds (if tx exist)
* Vote results
* Block commit and storage

---

## Manual Testing (Local, Without Docker)

Open 3 terminal windows:

### Terminal 1 - Leader

```bash
$env:NODE_ID="node1"
$env:IS_LEADER="true"
$env:PEERS="localhost:50052,localhost:50053"
$env:PORT="50051"
go run cmd/main.go
```

### Terminal 2 - Follower 1

```bash
$env:NODE_ID="node2"
$env:IS_LEADER="false"
$env:PEERS="localhost:50051"
$env:PORT="50052"
go run cmd/main.go
```

### Terminal 3 - Follower 2

```bash
$env:NODE_ID="node3"
$env:IS_LEADER="false"
$env:PEERS="localhost:50051"
$env:PORT="50053"
go run cmd/main.go
```

---

## Using the CLI (Send Transaction)

```bash
go run cmd/cli/main.go
```

This simulates Alice sending 99.99 coins to Bob with digital signature.

If node replies with:

```bash
✅ Phản hồi từ node: Transaction received
```

\=> means signature & transaction worked.

---

## How It Works

1. CLI creates transaction:

   * Generates ECDSA keypair
   * Signs transaction using Alice's private key
   * Sends to node via gRPC

2. Leader node stores tx in pending list.

3. Every 10s, leader packs pending txs into block:

   * Computes Merkle Root
   * Hashes block
   * Sends to followers (via `ProposeBlock`)

4. Followers verify and vote (`Vote.Accepted=true`)

5. Leader gathers votes via `SubmitVote`:

   * If >= 2/3 accepted => commit block
   * Save block to LevelDB

---

## Future Improvements

* Implement block syncing when node rejoins
* Add transaction pool expiration
* REST API for browser interaction
* Web dashboard to visualize chain and nodes
* Real Merkle tree instead of naive hash concat
* Secure CLI key management (wallets)