package raft

import (
	"context"
	"log"
	"time"

	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

// RaftNode wraps an etcd/raft Node with everything needed to run
// the Ready loop and apply committed entries to the FSM.
type RaftNode struct {
	node      raft.Node
	storage   *raft.MemoryStorage
	fsm       *FSM
	id        uint64
	transport Transport
}

func (rn *RaftNode) Leader() uint64 { return rn.node.Status().Lead }

// NewRaftNode creates and starts a Raft node with the supplied cluster peers.
func NewRaftNode(id uint64, peers []uint64, transport Transport, fsm *FSM) *RaftNode {
	storage := raft.NewMemoryStorage()

	c := &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	raftPeers := make([]raft.Peer, 0, len(peers))
	for _, peerID := range peers {
		raftPeers = append(raftPeers, raft.Peer{ID: peerID})
	}
	node := raft.StartNode(c, raftPeers)

	return &RaftNode{
		node:      node,
		storage:   storage,
		fsm:       fsm,
		id:        id,
		transport: transport,
	}
}

// Run starts the ticking + Ready loop. Blocks forever (run it in a goroutine).
func (rn *RaftNode) Run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rn.node.Tick()

		case rd := <-rn.node.Ready():
			// 1. Persist rd.Entries to storage (they're not committed yet,
			//    just need to be durable before we ack them)
			rn.storage.Append(rd.Entries)

			// 2. Send after persistence and before applying commits.
			rn.transport.Send(rd.Messages)

			// 3. Apply rd.CommittedEntries to the FSM.
			for _, entry := range rd.CommittedEntries {
				if entry.GetType() != raftpb.EntryNormal || len(entry.Data) == 0 {
					continue
				}
				err := rn.fsm.Apply(entry.Data)
				if err != nil {
					log.Printf("error applying entry: %v", err)
				}
			}

			// 4. Signal we're done processing this Ready batch.
			rn.node.Advance()

		case msg, ok := <-rn.transport.Recv():
			if !ok {
				return
			}
			if err := rn.node.Step(context.Background(), msg); err != nil {
				log.Printf("error stepping raft message: %v", err)
			}
		}
	}
}

// Propose submits a new command to the raft log.
func (rn *RaftNode) Propose(data []byte) error {
	return rn.node.Propose(context.Background(), data)
}
