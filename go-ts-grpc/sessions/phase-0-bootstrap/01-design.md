---
conversation_id: 2c04098f-1ddf-489e-b9f9-1ec8238b6cc1
date: 2026-05-14
project: go-ts-grpc (CI/CD Pipeline Runner)
status: design phase — pre-proto, decisions locked, ready to draft proto
---

# Design Session Log

Companion to [README.md](../../README.md). README.md is the high-level pitch; this file captures design decisions made during planning conversations so future sessions can pick up cold.

---

## Tooling Decisions

| Area | Decision | Why |
|---|---|---|
| Proto compiler | `buf` (not raw `protoc`) | Declarative codegen, lint, breaking-change detection |
| Go codegen | `protoc-gen-go` + `protoc-gen-go-grpc` via buf | Standard |
| TS codegen | `ts-proto` with `outputServices=grpc-js` | Idiomatic TS, full type safety, works with `@grpc/grpc-js` |
| TS runtime | `@grpc/grpc-js` | Pure JS, no native deps |
| Dev environment | Native macOS, no Docker | Docker Desktop overhead not worth parity; Mac+Linux behave identically for this scope |
| Docker | Phase 5 capstone only (Dockerfile as exercise) | Not a dev-environment decision |

## Repo Layout

```
go-ts-grpc/
  proto/
    pipeline/v1/pipeline.proto   # versioned package path (buf convention)
  buf.yaml
  buf.gen.yaml
  gen/
    go/...    (gitignored, regenerated)
    ts/...    (gitignored, regenerated)
  server/     # Go
  client/     # TS
```

Package: `pipeline.v1`. Version in path from day one so breaking-change detection works.

---

## Domain Model

### Pipeline
- Group of jobs (1:N)
- Fields: `id`, `name`, `owner_user_id`, `created_at`, `jobs[]`
- `overall_status` is **derived** from job statuses, not stored
- Sequential execution only (no DAG — deferred)

### Job
- Belongs to a pipeline (composite key: pipeline_id + job_id)
- Has one `Step` via `oneof`: `CommandStep` or `CodeStep`
- Fields: `id`, `pipeline_id`, `name`, `step`, `status`, `timeout_seconds`, `env_vars`, `working_dir`, `started_at`, `ended_at`
- **No logs field, no signal history field** — those are not persisted on Job

### Step (oneof)
- `CommandStep { command: string }` → exec via `sh -c`
- `CodeStep { language: "python" | "node", source: string }` → write source to temp file, exec interpreter
- ScriptStep dropped (folded into CodeStep concept)

### Execution Model
- Source written to temp file (not inline `-c`/`-e`) to avoid shell escaping, get proper line numbers in tracebacks, no ARG_MAX limit
- Each job gets cwd: `/tmp/pipeline-<id>/job-<id>/`
- Process killed via `exec.CommandContext` + `Setpgid: true` + group kill — works on macOS and Linux

### LogLine
- **Not stored on Job.** Lives only in stream and short-term buffer.
- Fields: `text`, `stream` (STDOUT|STDERR enum — merged single stream), `timestamp_ns`
- 5-minute in-memory ring buffer per job for late subscribers / reconnect replay
- Buffer freed after job ends + 5 min idle

---

## State Machine

States: `PENDING`, `RUNNING`, `SUCCEEDED`, `FAILED`, `CANCELLED`, `TIMED_OUT`

Transitions:
- **User-triggered:** `PENDING → CANCELLED`, `RUNNING → CANCELLED`
- **System-triggered:** `PENDING → RUNNING`, `RUNNING → SUCCEEDED`, `RUNNING → FAILED`, `RUNNING → TIMED_OUT`

Terminal: `SUCCEEDED`, `FAILED`, `CANCELLED`, `TIMED_OUT`.

Attempting an illegal transition (e.g., cancel a SUCCEEDED job) → `FAILED_PRECONDITION`.

---

## RPC Surface (7 endpoints)

| RPC | Type | Purpose |
|---|---|---|
| `CreateAndStartPipeline` | Unary | Submit pipeline + jobs, server validates and starts background execution, returns IDs |
| `GetPipeline` | Unary | Full pipeline + jobs snapshot |
| `GetJobStatus` | Unary | Single job snapshot (lighter than full pipeline) |
| `CancelJob` | Unary | Mutate state + kill subprocess |
| `StreamJobLogs` | Server stream | Replay buffered logs, then live tail. Closes on job end or client disconnect |
| `BatchCreateJobs` | Client stream | Stream job definitions into existing pipeline. First message = pipeline envelope, rest = jobs. Response = per-item result list (partial success) |
| `JobControlSession` | Bidi | Long-lived: client sends control signals (cancel/pause/resume etc), server pushes status updates |

All 4 gRPC RPC types covered.

---

## Identifiers

- Server-generated
- Short UUID strings
- Composite: job IDs unique within a pipeline (pipeline_id + job_id)

---

## Auth

- Static API key in `authorization: bearer <key>` metadata
- Two hardcoded keys mapped to user_ids (e.g., `dev-key-alice` → `alice`)
- Server interceptor: reads metadata → puts user_id on ctx via `context.WithValue`
- Owner check interceptor: only pipeline owner can cancel/control
- JWT, mTLS deferred — not worth the distraction

---

## Stream Subscriber Registry

Used for fan-out of logs and control sessions to currently-connected clients.

Rules:
- One user can have **multiple streams** (one per job, multiple stream types)
- Primary key: `(user_id, job_id, stream_type)`
- Cancel stream = client-side ctx cancel → server goroutine exits via `ctx.Done()` — no domain effect
- Cancel job = separate `CancelJob` unary RPC — kills subprocess + mutates state
- Concurrency: `sync.RWMutex` on the registry map, single big lock to start
- Cleanup: `defer registry.Deregister(handle)` first line after register

Buffering policy when subscriber is slow:
- **Drop the line for that subscriber** OR **disconnect the slow subscriber**
- Never block the producer — one slow client must not stall the job

**Don't build registry yet.** Not needed for Phase 1 (unary only). Build for Phase 2 with log fan-out as concrete use case, generalize for control sessions in Phase 4.

---

## Parked / Deferred

| Topic | Status |
|---|---|
| Channel direction, buffer size, lock granularity details | Revisit when building registry in Phase 2 |
| Cleanup (goroutines, processes, resources) | Separate session, before Phase 2 |
| State persistence (sync.Map vs mutex+map vs owner goroutine) | Decide when Phase 2 introduces concurrency |
| JobEvent audit log | Out of scope for now |
| DAG pipelines | v2 idea |
| Docker | Phase 5 capstone exercise |

---

## What's Done

- [x] Stack chosen (Go server, TS client, buf, ts-proto)
- [x] Repo layout planned
- [x] Domain entities defined
- [x] Step model (oneof Command|Code) locked
- [x] Execution strategy (temp file + interpreter)
- [x] State machine drawn (5 states, user vs system transitions)
- [x] All 7 RPCs identified + mapped to RPC types
- [x] Identifier scheme decided
- [x] Auth approach decided
- [x] Log buffering policy decided
- [x] Stream/job cancel semantics separated

## What's Next

1. **Step 3 design exercise** — for each of the 7 RPCs, write down on paper: request fields, response/stream fields, error cases (with status codes), auth required (yes/no). Slow but bulletproof.
2. **OR** jump straight to drafting `pipeline/v1/pipeline.proto` and discover gaps mid-typing.
3. Once proto drafted: `buf.yaml` + `buf.gen.yaml` + run `buf generate`, confirm Go and TS files produced.
4. Phase 1: minimal Go server + TS client doing one unary `CreateAndStartPipeline` round-trip.

## Open Questions

- Path A (paper design) or Path B (write proto and iterate)? — pending user decision next session.
