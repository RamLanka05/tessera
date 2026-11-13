package raft

import (
	"sync"

	"go.etcd.io/raft/v3/raftpb"
)

const inprocInboxSize = 256

// InprocNetwork routes Raft messages between nodes in the same process.
type InprocNetwork struct {
	mu           sync.Mutex
	inboxes      map[uint64]chan *raftpb.Message
	disconnected map[uint64]struct{}
}

// Register creates an inbox and returns the transport for id.
func (n *InprocNetwork) Register(id uint64) Transport {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.inboxes == nil {
		n.inboxes = make(map[uint64]chan *raftpb.Message)
	}
	if n.disconnected == nil {
		n.disconnected = make(map[uint64]struct{})
	}

	inbox := make(chan *raftpb.Message, inprocInboxSize)
	n.inboxes[id] = inbox
	return &inprocTransport{network: n, id: id, inbox: inbox}
}

func (n *InprocNetwork) Disconnect(id uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.disconnected == nil {
		n.disconnected = make(map[uint64]struct{})
	}
	n.disconnected[id] = struct{}{}
}

func (n *InprocNetwork) Reconnect(id uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.disconnected, id)
}

type inprocTransport struct {
	network *InprocNetwork
	id      uint64
	inbox   chan *raftpb.Message
}

func (t *inprocTransport) Send(msgs []*raftpb.Message) {
	t.network.mu.Lock()
	defer t.network.mu.Unlock()

	if _, disconnected := t.network.disconnected[t.id]; disconnected {
		return
	}

	for _, msg := range msgs {
		if msg == nil || msg.From == nil || msg.To == nil {
			continue
		}
		if _, disconnected := t.network.disconnected[*msg.From]; disconnected {
			continue
		}
		if _, disconnected := t.network.disconnected[*msg.To]; disconnected {
			continue
		}

		inbox, ok := t.network.inboxes[*msg.To]
		if !ok {
			continue
		}

		// Dropping messages is safe because Raft retries messages after loss.
		select {
		case inbox <- msg:
		default:
		}
	}
}

func (t *inprocTransport) Recv() <-chan *raftpb.Message {
	return t.inbox
}
