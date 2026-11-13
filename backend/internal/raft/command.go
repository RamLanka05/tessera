package raft

type OpType string

const (
	OpCreate OpType = "create"
	OpRollback OpType = "rollback"
)

type Command struct {
	Op OpType `json:"op"`
	ConfigKey string `json:"key"`
	ConfigValue interface{} `json:"value,omitempty"`
	Author string `json:"author,omitempty"`
	Message string `json:"message,omitempty"`
	TargetVID int `json:"target_vid,omitempty"`
}