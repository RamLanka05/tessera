package main

import (
	"encoding/json"
	"fmt"
	"time"

	"tessera/backend/internal/raft"
	"tessera/backend/internal/storage"
)

func main() {
	fmt.Println("Starting single-node Raft smoke test...")

	// Basic setup: initialize the Postgres store and the FSM.
	db := storage.InitDB()
	defer db.Close()
	store := storage.NewPostgresStore(db)
	fsm := raft.NewFSM(store)

	// Build and start the single-node RaftNode.
	network := &raft.InprocNetwork{}
	node := raft.NewRaftNode(1, []uint64{1}, network.Register(1), fsm)

	// Run the Ready loop in the background — it needs to keep
	// ticking/processing continuously.
	go node.Run()

	time.Sleep(2 * time.Second)

	// Build a real Command and propose it.
	cmd := raft.Command{
		Op:        raft.OpCreate,
		ConfigKey: "smoke-test-key",
		ConfigValue: map[string]string{
			"hello": "world",
		},
		Author:  "smoke-test",
		Message: "verifying single-node raft loop",
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		fmt.Printf("failed to marshal command: %v\n", err)
		return
	}

	if err := node.Propose(data); err != nil {
		fmt.Printf("failed to propose: %v\n", err)
		return
	}

	fmt.Println("Proposed command, waiting for it to commit and apply...")
	time.Sleep(100 * time.Millisecond)

	// Verify it actually landed in Postgres via the store.
	val, err := store.GetConfig("smoke-test-key")
	if err != nil {
		fmt.Printf("FAILED — could not read back config: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS — read back config value: %s\n", string(val))
}