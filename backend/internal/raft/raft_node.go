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
	node    raft.Node
	storage *raft.MemoryStorage
	fsm     *FSM
	id      uint64
}

// NewRaftNode creates and starts a single raft node.
// For now, peers is just this one node — no real cluster yet.
func NewRaftNode(id uint64, fsm *FSM) *RaftNode {
	storage := raft.NewMemoryStorage()

	c := &raft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	node := raft.StartNode(c, []raft.Peer{{ID: id}})

	return &RaftNode{
		node:    node,
		storage: storage,
		fsm:     fsm,
		id:      id,
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

			// 2. Send rd.Messages to peers over the network.
			//    No transport yet — just log them for now.
			for _, msg := range rd.Messages {
				log.Printf("would send message: %+v\n", msg)
			}

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
		}
	}
}

// Propose submits a new command to the raft log.
func (rn *RaftNode) Propose(data []byte) error {
	return rn.node.Propose(context.Background(), data)
}