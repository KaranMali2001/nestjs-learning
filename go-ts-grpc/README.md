# Project — CI/CD Pipeline Runner + Code Execution Engine

## What We're Building

A CI/CD pipeline runner where each job IS a code execution.
Mirrors: GitHub Actions, CircleCI internals.

```
Pipeline  →  Job 1: run tests     (executes shell commands, streams output)
          →  Job 2: build binary  (executes shell commands, streams output)
          →  Job 3: deploy        (executes shell commands, streams output)
```

**Stack:** Go server + TypeScript client (Node.js, @grpc/grpc-js)

---

## Repo Structure

```
docs/                 — stable reference notes (synthesized after sessions)
  concepts.md         — gRPC fundamentals
sessions/             — chronological learning/decision logs
  index.md            — TOC of every session
  README.md           — convention guide
  _template.md        — starter for new sessions
  phase-0-bootstrap/  — design + setup conversations
  phase-1-unary/      — first unary RPC
  phase-2-streaming/  — server streaming
  phase-3-client-streaming/
  phase-4-bidi/
  phase-5-harden/
proto/                — protobuf source
server/               — Go gRPC server (owns its go.mod, gen/)
client/               — TS gRPC client (owns its package.json, gen/)
buf.yaml              — proto module config
buf.gen.yaml          — codegen plugin config
```

- Each session = one AI conversation, captured in markdown with frontmatter (conversation_id, date, status).
- `sessions/` = raw history. `docs/` = synthesized stable knowledge extracted from sessions.
- Start a new session by copying `sessions/_template.md` into the relevant phase folder.

---

## Why This Project

- Deadline + cancellation are CORE features, not bolted on
  - Every job has a timeout — if it runs too long, kill it
  - Cancel button must actually kill the running process
- All 4 RPC types fit naturally, nothing forced
- No simulator needed — shell commands generate real output
- Mirrors a real production system you've used

---

## The 4 RPC Types Mapped

| RPC | Method | What it does |
|---|---|---|
| Unary | `CreatePipeline` | Submit a pipeline definition, get a pipeline ID back |
| Unary | `GetJobStatus` | Poll current status of a job |
| Server Streaming | `StreamJobLogs` | Watch live output as job runs, line by line |
| Client Streaming | `UploadArtifacts` | Stream build artifacts (binaries, test reports) after job |
| Bidi | `JobControlSession` | Client sends cancel/pause signals, server streams live status back |

---

## gRPC Concepts Each Phase Teaches

### Phase 1 — Unary (Week 1)
Build `CreatePipeline` and `GetJobStatus`.

Concepts:
- Proto file syntax — messages, services, rpc definitions
- Code generation with buf
- Basic Go server + TS client setup
- Status codes — `NOT_FOUND` when job ID doesn't exist, `INVALID_ARGUMENT` for bad pipeline config

### Phase 2 — Server Streaming (Week 2)
Build `StreamJobLogs`.

Concepts:
- Server streaming — send output line by line as shell command runs
- Cancellation — client hits stop, server must kill the running process via `ctx.Done()`
- Deadlines — job has a max execution time, `DEADLINE_EXCEEDED` when it runs too long
- Back pressure — what if job outputs 10,000 lines/sec? Watch write calls block

### Phase 3 — Client Streaming (Week 3)
Build `UploadArtifacts`.

Concepts:
- Client streaming — stream file chunks to server
- Flow control — large files, server processes slower than client sends
- `END_STREAM` — server waits for client to finish, then returns upload summary
- Metadata — send filename, content-type as metadata headers

### Phase 4 — Bidirectional Streaming (Week 4)
Build `JobControlSession`.

Concepts:
- True bidi — client sends control signals (cancel/pause/resume), server streams live status
- Interceptors — auth interceptor (only pipeline owner can cancel)
- Metadata — pass auth token, pipeline ID in headers
- Deadline propagation — pipeline deadline flows through to each job's ctx
- Context propagation — ctx passed through every layer

### Phase 5 — Harden (Week 5)
Layer production concerns across everything.

Concepts:
- Health checking — expose health endpoint
- Graceful shutdown — finish running jobs before server stops
- Proper status codes audit — review every handler
- Server interceptors — logging (method, duration, status), recovery (catch panics)
- Client interceptors — attach auth metadata to every call automatically

---

## Proto Design (Rough)

```protobuf
service PipelineRunner {
  // Unary
  rpc CreatePipeline(CreatePipelineRequest) returns (CreatePipelineResponse);
  rpc GetJobStatus(GetJobStatusRequest) returns (JobStatus);

  // Server Streaming
  rpc StreamJobLogs(StreamJobLogsRequest) returns (stream LogLine);

  // Client Streaming
  rpc UploadArtifacts(stream ArtifactChunk) returns (UploadSummary);

  // Bidirectional
  rpc JobControlSession(stream ControlSignal) returns (stream JobStatusUpdate);
}
```

---

## Simulation Plan (No fake data needed)

```
Phase 1: curl-equivalent TS script
  → create a pipeline with 2 jobs
  → poll status until done

Phase 2: watch logs in terminal
  → trigger a job that runs `for i in {1..100}; do echo $i; sleep 0.1; done`
  → watch output stream live
  → cancel mid-run, verify process actually dies

Phase 3: upload a real file
  → stream a local file in 1KB chunks
  → verify server reassembles correctly

Phase 4: open two terminals
  → terminal 1: start a long-running job
  → terminal 2: connect to JobControlSession, send cancel
  → verify terminal 1 job stops

Phase 5: kill the server mid-job
  → verify graceful shutdown waits for job to finish
  → verify health endpoint goes NOT_SERVING before shutdown
```

---

## Key Technical Decisions to Make

- How to actually kill a running process when ctx is cancelled (Go: `exec.CommandContext`)
- Where to store pipeline/job state (in-memory map is fine for learning)
- How to handle multiple concurrent jobs (goroutine per job)
- Artifact storage (write to local disk, return file path)
- Auth strategy (simple JWT in metadata, verified by interceptor)

---

## Reference Links

- Official gRPC guides: https://grpc.io/docs/guides/
- Uber push platform (bidi streaming in production): https://www.uber.com/us/en/blog/ubers-next-gen-push-platform-on-grpc/
- Dropbox Courier (production gRPC patterns): https://dropbox.tech/infrastructure/courier-dropbox-migration-to-grpc
- Datadog gRPC at scale: https://www.datadoghq.com/blog/grpc-at-datadog/
