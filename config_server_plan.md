# Config Server: Tech Stack & Phase Breakdown

## Tech Stack

### Backend
- **Language:** Go
  - Why: Goroutines for concurrent I/O, mature Raft libraries, built-in performance
  - Alternative: Node.js (if you want to stick with JavaScript, but Go is stronger for distributed systems)

### Consensus & Replication
- **Library:** `etcd/raft` (Go)
  - Mature, battle-tested (used in etcd itself)
  - Handles leader election, log replication, state snapshots
  - Avoids "build Raft from scratch" rabbit hole
- **Communication:** gRPC between nodes (fast, efficient)

### Database & Storage
- **Source of Truth (Write Path):** PostgreSQL
  - Stores config versions, audit logs, version tree
  - Schema:
    ```
    configs (id, name, value_json, created_at)
    versions (id, config_id, parent_version_id, value_json, timestamp, author)
    audit_log (id, config_id, action, timestamp, user)
    active_pointers (config_name, active_version_id, last_updated)
    ```

- **Read Cache (Read Path):** Redis or in-memory KV (start simple)
  - Store flattened latest versions
  - Sub-millisecond reads
  - Cache invalidation on write

### Frontend
- **Framework:** React + TypeScript
- **State Management:** React hooks or simple Redux
- **Real-time:** WebSocket client library (e.g., socket.io)
- **Visualization:** Directed Acyclic Graph (DAG) for version tree
  - Use a library like `react-flow-renderer` or `vis-network`

### Real-Time Communication
- **Server → Client Push:** WebSockets (or Server-Sent Events as fallback)
- **Server ↔ Server:** gRPC

### Security
- **JWT Auth:** RS256 (asymmetric signing)
- **Token Generation:** Backend signs with private key
- **Token Validation:** Clients validate with public key (embedded)
- **Revocation:** In-memory blacklist on read nodes + distributed via Raft log

### Deployment & Testing
- **Containerization:** Docker
- **Local Multi-Node Testing:** Docker Compose (3-5 nodes)
- **Observability:** Prometheus (metrics) + Grafana (optional but nice)
- **Chaos Testing:** Custom bash/Go scripts to kill nodes, simulate DB failures
- **CI/CD:** GitHub Actions (linting, tests, Docker image scan)

### Deployment Target (When Ready)
- **Option A (Simple):** DigitalOcean App Platform or Fly.io (managed containers)
- **Option B (More Complex):** AWS ECS or self-managed Kubernetes (overkill for MVP, but good learning)
- Start simple; add complexity if time permits.

---

## Phase Breakdown (5-Month Timeline)

### Phase 0: Setup & Foundation (Week 1)
**Goal:** Boilerplate + schema ready to go
- Initialize Go project structure
- Set up PostgreSQL locally (Docker)
- Write schema migrations
- Initialize React project (Vite or CRA)
- Sketch out API contract (OpenAPI spec)
- Set up GitHub repo + basic GitHub Actions pipeline

**Deliverable:**
- Empty but structured Go backend (HTTP server, no logic yet)
- PostgreSQL running locally
- React dashboard skeleton

---

### Phase 1: Basic Config API & Admin Dashboard (Weeks 2–4, ~3 weeks)
**Goal:** CRUD operations, basic versioning, no consensus yet

**Backend:**
- REST API endpoints:
  - `GET /config/:name` (read latest)
  - `POST /config` (create/update)
  - `GET /config/:name/versions` (list version history)
  - `POST /config/:name/rollback/:version_id` (manual rollback)
  - `POST /auth/token` (JWT generation)
- Simple in-memory JWT validation
- PostgreSQL writes for configs + versions
- Basic error handling & logging

**Frontend:**
- Login page (submit client credentials → get JWT)
- Config list view (show all configs)
- Config detail page (view current value, version history)
- Create/edit form (basic)
- Manual rollback button (swaps `active_pointers.active_version_id`)

**Database:**
- Run migrations, seed test data
- Verify reads/writes work end-to-end

**Testing:**
- Unit tests for core API logic
- Integration test: create config → read back → rollback

**Deliverable:**
- Working API + dashboard (single-node, no consensus)
- Can CRUD configs, track versions, manually rollback
- GitHub actions CI passing

---

### Phase 2: Raft Consensus & Multi-Node Setup (Weeks 5–7, ~3 weeks)
**Goal:** Distribute consensus across 3+ nodes, prove leader election works

**Backend:**
- Integrate `etcd/raft` library
- Implement Raft state machine interface:
  - `Apply()` — Apply a log entry (config change) to local state
  - `Snapshot()` — Serialize current state for new nodes joining
  - `Restore()` — Load snapshot on startup
- gRPC endpoints for inter-node communication
  - Raft message passing (RequestVote, AppendEntries, etc.)
- Leader detection:
  - Leader handles write requests
  - Followers forward writes to leader
- Persist Raft state to PostgreSQL (log entries, snapshots)

**Testing:**
- Docker Compose file: 3 nodes running locally
- Test leader election:
  - Start 3 nodes, verify 1 becomes leader
  - Kill leader, verify election of new leader
  - Kill 2 nodes, verify 1-node minority can't commit (safety)
- Test config replication:
  - Write config on leader
  - Verify it propagates to followers
  - Verify all nodes agree on value

**Deliverable:**
- Multi-node cluster running locally
- Config writes replicate via Raft
- Post-mortem #1: "Leader Election Worked, But Took Longer Than Expected" (document any findings)

---

### Phase 3: Read Path Optimization & Real-Time Push (Weeks 8–9, ~2 weeks)
**Goal:** Sub-millisecond reads, real-time updates to clients

**Backend:**
- In-memory or Redis KV store for flattened configs
  - On every Raft commit, flatten the config JSON and store in KV
  - Read API queries KV (not the version tree DB)
- Separate read-only API endpoint:
  - `GET /api/v1/config/:name` (hits KV, ultra-fast)
  - No auth overhead (or lightweight validation)
- WebSocket server for real-time push:
  - Client connects to `/ws`
  - Server maintains active connection list
  - On config update, broadcast to all connected clients
  - Clients re-fetch config via read API

**Frontend:**
- WebSocket client library (socket.io or native WS)
- Listen for `config_updated` events
- On event, re-fetch config in memory
- Update UI (optional animation: "Config updated in real-time")

**Testing:**
- Benchmark: measure read latency (target: <1ms)
- Connect 10 mock clients via WebSocket
- Change a config, verify all 10 clients receive update within 100ms
- Load test: send 1000 reads/sec to read API, measure latency

**Deliverable:**
- Read-optimized API + KV layer working
- WebSocket push mechanism in place
- Performance benchmarks documented

---

### Phase 4: Admin Dashboard Enhancements & Panic Button (Weeks 10–11, ~2 weeks)
**Goal:** Full operational dashboard with version tree visualization and panic button

**Frontend:**
- Version tree visualization (DAG)
  - Show all versions as nodes
  - Show parent-child relationships as edges
  - Highlight active version (green), other versions (gray)
  - Click a version to preview its config
- "Panic Button" UI:
  - Large red button in top-right
  - On click: rollback active_pointer to previous stable version
  - Confirm dialog: "Are you sure?"
  - Show rollback status (success/error)
- Config detail improvements:
  - Show who changed what, when (audit log)
  - Diff viewer (compare two versions)
  - Tag versions (e.g., "stable", "beta")

**Backend:**
- Audit log endpoint: `GET /config/:name/audit`
- Diff endpoint: `GET /config/:name/versions/:v1/diff/:v2`
- Panic button endpoint: `POST /config/:name/panic` (sets `active_pointer` to previous version)

**Testing:**
- UI test: render version tree for config with 5+ versions
- Test panic button: rollback works, clients notified in real-time

**Deliverable:**
- Polished React dashboard
- Panic button operational
- Diff viewer working
- Audit trail visible

---

### Phase 5: JWT Revocation, Security Hardening & Observability (Weeks 12–14, ~2-3 weeks)
**Goal:** Production-grade security + monitoring

**Backend:**
- JWT revocation system:
  - Maintain revocation blacklist in distributed cache (Redis)
  - On login, return JWT + short TTL (5–15 min)
  - Revocation endpoint: `POST /auth/revoke/:token_jti`
    - Adds token ID to blacklist
    - Broadcasts revocation via Raft log to all nodes
  - On each request, check token against blacklist before processing
- Prometheus metrics:
  - Request latency (histogram)
  - RPC errors per node
  - Raft term, commit index (cluster health)
  - Config change rate
  - JWT token revocations
- Structured logging (JSON logs)
  - All requests logged with trace ID
  - Errors include full context

**Testing:**
- Revoke a token, verify subsequent requests fail
- Verify revocation broadcasts to all 3 nodes
- Collect Prometheus metrics, verify sensible values

**Deliverable:**
- JWT revocation working
- Prometheus endpoint scraped
- Security audit checklist completed

---

### Phase 6: Chaos Engineering & Post-Mortems (Weeks 15–18, ~3-4 weeks)
**Goal:** Intentional failures, observability, documentation

**Chaos Scenarios (write bash/Go scripts to trigger):**

1. **Kill Leader Node**
   - Start 3-node cluster
   - Kill leader (SIGKILL)
   - Measure: time to election + availability during election
   - Expected: <500ms re-election, 0 data loss
   - Post-mortem: "Leader Failure Recovery"

2. **Database Connection Loss**
   - Start cluster, write a config
   - Simulate DB timeout (kill PostgreSQL container)
   - Measure: does Raft buffer logs until DB recovers?
   - Expected: system continues in memory, persists to DB when it recovers
   - Post-mortem: "Database Partition Survival"

3. **Network Partition (Split Brain Prevention)**
   - 3-node cluster, partition into 1 vs. 2
   - Try writes on both sides
   - Expected: minority (1-node) rejects writes, majority (2-node) accepts
   - Measure: verify no conflicting versions
   - Post-mortem: "Split-Brain Prevention"

4. **High Load Surge**
   - Simulate 5000 reads/sec for 60 seconds
   - Measure: latency under load, memory growth
   - Expected: <10ms p99 latency, stable memory
   - Post-mortem: "Load Testing Findings & Bottleneck Discovery"

5. **Cascading Follower Failures**
   - Start 5-node cluster
   - Kill nodes 2, 3, 4 sequentially
   - Expected: cluster stays alive (3-node quorum)
   - Measure: availability during cascade
   - Post-mortem: "Resilience to Cascading Failures"

**For Each Chaos Test:**
- Run the scenario
- Capture logs, metrics, timings
- Write a Post-Mortem markdown file:
  - What was the failure?
  - How did the system behave?
  - What did monitoring show?
  - How did it recover?
  - What can be improved?

**Observability Improvements:**
- Add Grafana dashboard (optional but impressive)
- Dashboards: Cluster Health, Request Latency, Error Rate, Raft State
- Alert thresholds: high error rate, node down, etc.

**Deliverable:**
- 5+ chaos test scripts in `/chaos` folder
- 5+ post-mortem markdown files
- Prometheus + Grafana running locally (or screenshots)
- GitHub issue tracker populated with findings

---

## Timeline Summary

| Phase | Duration | Goal |
|-------|----------|------|
| **Phase 0** | Week 1 | Setup, schema, boilerplate |
| **Phase 1** | Weeks 2–4 | Basic CRUD API + dashboard |
| **Phase 2** | Weeks 5–7 | Raft consensus, multi-node |
| **Phase 3** | Weeks 8–9 | Read optimization, WebSocket push |
| **Phase 4** | Weeks 10–11 | Dashboard UX, panic button |
| **Phase 5** | Weeks 12–14 | JWT revocation, monitoring |
| **Phase 6** | Weeks 15–18 | Chaos testing + post-mortems |
| **Buffer** | Weeks 19–20 | Polish, documentation, bonus features |

**Total: ~20 weeks (5 months)** ✅

---

## Bonus Features (If Time Permits)

- [ ] Multi-region replication (Raft followers in different cloud regions)
- [ ] GitOps integration (export configs to GitHub, sync back)
- [ ] Config branching (like Git branches, for staging)
- [ ] Webhooks (notify external systems on config changes)
- [ ] Rate limiting per client (prevent abuse)
- [ ] Terraform provider (define configs as IaC)
- [ ] CLI tool for local development

---

## GitHub Repo Structure

```
config-server/
├── backend/
│   ├── main.go
│   ├── api/          (REST handlers)
│   ├── consensus/    (Raft state machine)
│   ├── storage/      (PostgreSQL interaction)
│   ├── grpc/         (inter-node communication)
│   ├── auth/         (JWT, revocation)
│   └── metrics/      (Prometheus)
├── frontend/
│   ├── src/
│   │   ├── components/   (Dashboard, versioning UI)
│   │   ├── hooks/        (WebSocket connection)
│   │   └── pages/
│   └── package.json
├── chaos/            (chaos test scripts)
├── deployments/
│   ├── docker-compose.yml  (3-node local cluster)
│   ├── Dockerfile
│   └── k8s/         (if deploying to cloud)
├── docs/
│   ├── ARCHITECTURE.md
│   ├── POST_MORTEM_*.md
│   └── API.md
├── tests/
│   └── integration_test.go
├── .github/workflows/
│   └── ci.yml       (tests, Docker scan, deploy)
└── README.md
```

---

## Recruiting Story (What You'll Tell Interviewers)

> "I built a distributed configuration server from first principles. It uses Raft consensus to replicate config changes across multiple nodes safely, even if nodes crash. The system decouples writes (complex version tracking in PostgreSQL) from reads (ultra-fast in-memory KV store), so configs can be served to millions of clients in under a millisecond.
> 
> The admin dashboard (React) lets you visualize the version tree, modify configs, and hit a 'Panic Button' to roll back globally in real-time. Behind the scenes, I used Raft for leader election, gRPC for inter-node communication, and Prometheus for deep observability.
> 
> To validate production readiness, I ran intentional chaos tests: killed nodes mid-transaction, simulated database failures, and induced network partitions. The system self-healed every time. I documented each failure in a post-mortem to show how I debugged and improved the system.
> 
> This taught me how distributed systems handle consistency, availability, and fault tolerance—and how to design for observability from day one."

That's a **serious** engineering story.

---

## Next Steps

1. **Pick your Go starter**: Decide on `etcd/raft` vs. rolling your own (recommend using the library)
2. **Set up local PostgreSQL + Docker Compose** (do this this week)
3. **Draft API contract** (OpenAPI spec for all endpoints)
4. **Begin Phase 0 immediately**

Ready to start Phase 0?
