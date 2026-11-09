package main

import (
	"encoding/json"
	"fmt"
	"time"

	"tessera/backend/internal/raft"
	"tessera/backend/internal/storage"
)

// propose marshals a Command, proposes it, and waits briefly for it to apply.
func propose(node *raft.RaftNode, cmd raft.Command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	if err := node.Propose(data); err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)
	return nil
}

func main() {
	fmt.Println("Starting rollback-through-Raft test...")

	db := storage.InitDB()
	defer db.Close()
	store := storage.NewPostgresStore(db)
	fsm := raft.NewFSM(store)
	node := raft.NewRaftNode(1, fsm)
	go node.Run()
	time.Sleep(2 * time.Second) // election wait

	// Unique key per run so reruns don't pile onto old versions.
	key := fmt.Sprintf("rollback-test-%d", time.Now().UnixNano())

	// 1. Two creates: v1 then v2.
	for _, v := range []string{"one", "two"} {
		err := propose(node, raft.Command{
			Op:          raft.OpCreate,
			ConfigKey:   key,
			ConfigValue: map[string]string{"version": v},
			Author:      "test2",
			Message:     "create " + v,
		})
		if err != nil {
			fmt.Printf("propose create %s failed: %v\n", v, err)
			return
		}
	}

	// 2. Find the older version's ID.
	versions, err := store.ListVersions(key)
	if err != nil {
		fmt.Printf("list versions failed: %v\n", err)
		return
	}
	fmt.Printf("found %d versions\n", len(versions))

	// ??? 1: ListVersions orders by timestamp DESC, and each Version has
	// an IsActive flag. Which element is the v1 version you want to roll
	// back to? Pick its ID.
	targetVID := /* ??? */

	// ??? 2: build and propose the rollback Command. Which fields
	// does OpRollback actually use in fsm.go?
	if err := propose(node, /* ??? */); err != nil {
		fmt.Printf("propose rollback failed: %v\n", err)
		return
	}

	// 3. Verify the active value is v1 again.
	val, err := store.GetConfig(key)
	if err != nil {
		fmt.Printf("FAILED — read back: %v\n", err)
		return
	}
	// ??? 3: GetConfig returns JSONB bytes, and Postgres reformats JSON
	// whitespace, so don't compare the raw string against a literal.
	// What do you do with val before checking it equals "one"?

	// 4. Bad rollback: version ID that doesn't exist.
	_ = propose(node, raft.Command{Op: raft.OpRollback, ConfigKey: key, TargetVID: 999999})

	// ??? 4 (no code, just predict): what should show up in the terminal
	// from the Ready loop, and what should GetConfig return now?
	val, _ = store.GetConfig(key)
	fmt.Printf("after bad rollback, active value: %s\n", val)
}