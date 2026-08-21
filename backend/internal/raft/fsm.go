package raft

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	"tessera/backend/internal/storage"
// )

// type FSM struct {
// 	storage storage.ConfigStore // your Phase 1 storage interface
// }

// func NewFSM(store storage.ConfigStore) *FSM {
// 	return &FSM{storage: store}
// }

// // Apply takes a committed Raft entry (as []byte), unmarshals it to Command,
// // and executes the corresponding storage operation.
// func (f *FSM) Apply(data []byte) error {
// 	var cmd Command
// 	if err := json.Unmarshal(data, &cmd); err != nil {
// 		return fmt.Errorf("unmarshal command: %w", err)
// 	}

// 	switch cmd.Op {
// 	case OpCreate:
// 		// ??? call f.storage.CreateConfig with the right args
// 		// reminder: CreateConfig(db, configKey, configValue, author, message) (int, error)
// 		// but you don't have db here — your storage layer needs a method signature
// 		// that doesn't require passing db as an arg each time.
// 		// Question for you: does your Phase 1 storage.ConfigStore interface already
// 		// hide the db pointer? Or do you need to refactor it?

// 	case OpRollback:
// 		// ??? similar logic for RollbackConfig

// 	default:
// 		return fmt.Errorf("unknown op: %s", cmd.Op)
// 	}

// 	return nil
// }

// // Snapshot and Restore are required by raft.FSM but we'll use MemoryStorage for now.
// // Stub them out — we'll implement persistent snapshots later.
// func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
// 	return nil, nil // not yet
// }

// func (f *FSM) Restore(snapshot raft.FSMSnapshot) error {
// 	return nil // not yet
// }