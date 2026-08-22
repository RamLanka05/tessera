package raft

import (
	"encoding/json"
	"fmt"

	"tessera/backend/internal/storage"
)

// FSM (Finite State Machine) applies committed Raft log entries
// to the local storage layer.
type FSM struct {
	storage storage.ConfigStore
}

// NewFSM creates a new FSM backed by the given ConfigStore
func NewFSM(store storage.ConfigStore) *FSM {
	return &FSM{storage: store}
}

// Apply takes a committed Raft log entry (raw bytes), unmarshals it
// into a Command, and executes the corresponding storage operation.
func (f *FSM) Apply(data []byte) error {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return fmt.Errorf("unmarshal command: %w", err)
	}

	switch cmd.Op {
	case OpCreate:
		_, err := f.storage.CreateConfig(cmd.ConfigKey, cmd.ConfigValue, cmd.Author, cmd.Message)
		if err != nil {
			return fmt.Errorf("create config: %w", err)
		}

	case OpRollback:
		_, err := f.storage.RollbackConfig(cmd.ConfigKey, cmd.TargetVID)
		if err != nil {
			return fmt.Errorf("rollback config: %w", err)
		}
	default:
		return fmt.Errorf("unknown op: %s", cmd.Op)
	}

	return nil
}