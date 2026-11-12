// package raft

// import (
// 	"bytes"
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"testing"
// 	"time"

// 	_ "github.com/lib/pq" // match whatever driver your project uses

// 	"tessera/backend/internal/storage" // fix to your real import path
// )

// type testCluster struct {
// 	net   *InprocNetwork
// 	nodes []*RaftNode // nodes[i] has ID i+1
// 	dbs   []*sql.DB   // dbs[i] belongs to nodes[i]
// }

// func newTestCluster(t *testing.T, size int) *testCluster {
// 	t.Helper()

// 	dsnFmt := os.Getenv("TESSERA_TEST_DSN_FMT")
// 	if dsnFmt == "" {
// 		t.Skip("set TESSERA_TEST_DSN_FMT, e.g. postgres://user:pass@localhost:5432/tessera_%d?sslmode=disable")
// 	}

// 	peers := make([]uint64, size)
// 	for i := range peers {
// 		peers[i] = uint64(i + 1)
// 	}

// 	c := &testCluster{net: &InprocNetwork{}}

// 	// One database per node, otherwise every write lands 3x in the same DB
// 	// and the test proves nothing. Databases must already be migrated.
// 	for _, id := range peers {
// 		db, err := sql.Open("postgres", fmt.Sprintf(dsnFmt, id))
// 		if err != nil {
// 			t.Fatalf("open db for node %d: %v", id, err)
// 		}
// 		if err := db.Ping(); err != nil {
// 			t.Fatalf("ping db for node %d: %v", id, err)
// 		}
// 		t.Cleanup(func() { db.Close() })
// 		c.dbs = append(c.dbs, db)
// 	}

// 	// Register every inbox BEFORE starting any node, so early
// 	// messages aren't dropped for an unknown destination.
// 	transports := make(map[uint64]Transport)
// 	for _, id := range peers {
// 		transports[id] = c.net.Register(id)
// 	}

// 	for i, id := range peers {
// 		fsm := NewFSM(c.dbs[i]) // adapt to your FSM constructor
// 		c.nodes = append(c.nodes, NewRaftNode(id, peers, transports[id], fsm))
// 	}
// 	for _, n := range c.nodes {
// 		go n.Run()
// 	}
// 	return c
// }

// // waitForLeader blocks until all nodes agree on one leader.
// func (c *testCluster) waitForLeader(t *testing.T, timeout time.Duration) *RaftNode {
// 	t.Helper()
// 	deadline := time.Now().Add(timeout)
// 	for time.Now().Before(deadline) {
// 		var leader uint64
// 		agreed := true
// 		for _, n := range c.nodes {
// 			l := n.Leader()
// 			if l == 0 || (leader != 0 && l != leader) {
// 				agreed = false
// 				break
// 			}
// 			leader = l
// 		}
// 		if agreed {
// 			return c.nodes[leader-1]
// 		}
// 		time.Sleep(50 * time.Millisecond)
// 	}
// 	t.Fatal("no stable leader elected before timeout")
// 	return nil
// }

// // waitForConfig polls one node's DB until the key appears.
// func waitForConfig(t *testing.T, db *sql.DB, node int, key string, timeout time.Duration) []byte {
// 	t.Helper()
// 	deadline := time.Now().Add(timeout)
// 	for time.Now().Before(deadline) {
// 		if val, err := storage.GetConfig(db, key); err == nil {
// 			return val
// 		}
// 		time.Sleep(50 * time.Millisecond)
// 	}
// 	t.Fatalf("node %d never applied key %q", node, key)
// 	return nil
// }

// func TestThreeNodeReplication(t *testing.T) {
// 	c := newTestCluster(t, 3)
// 	leader := c.waitForLeader(t, 10*time.Second)

// 	// Unique key so reruns don't collide with old data in the DBs.
// 	key := fmt.Sprintf("test-key-%d", time.Now().UnixNano())
// 	cmd := Command{
// 		Op:          OpCreate,
// 		ConfigKey:   key,
// 		ConfigValue: map[string]any{"feature": "on"},
// 		Author:      "test",
// 		Message:     "replication test",
// 	}
// 	data, err := json.Marshal(cmd)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := leader.Propose(data); err != nil {
// 		t.Fatalf("propose: %v", err)
// 	}

// 	first := waitForConfig(t, c.dbs[0], 1, key, 10*time.Second)
// 	for i := 1; i < len(c.dbs); i++ {
// 		got := waitForConfig(t, c.dbs[i], i+1, key, 10*time.Second)
// 		if !bytes.Equal(first, got) {
// 			t.Errorf("node %d has %s, node 1 has %s", i+1, got, first)
// 		}
// 	}
// }