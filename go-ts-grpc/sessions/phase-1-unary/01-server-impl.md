---
conversation_id: 98b218a8-a0db-47f6-9cab-8115681a3a3b
date: 2026-05-29
project: go-ts-grpc (CI/CD Pipeline Runner)
phase: phase-1-unary
status: server boots, CreatePipelineAndJobs echoes request, verified end-to-end with grpcui
---

# Phase 1.01 — First Unary Server Implementation

First time wiring up an actual gRPC server in Go. Started from a stub `main.go` (just `fmt.Println("HELLO SERVER")`) and ended with a working server that:
- accepts the `CreatePipelineAndJobs` RPC,
- reads incoming metadata,
- rejects requests missing an `auth` header,
- echoes back the submitted pipeline,
- and is reachable from grpcui via reflection.

The focus was **syntax + concepts**, not business logic — the handler still just echoes input.

---

## Decisions Made

- **Decision:** Use `package main` (not `package server`) for the server entrypoint.
  **Why:** Go requires `package main` for an executable with `func main()`. Started with `package server` which silently did nothing.
  **Where it lives in code:** `server/main.go:1`

- **Decision:** Port `:8080` for the gRPC server (not the conventional `:50051`).
  **Why:** Personal preference; gRPC has no enforced port, `:50051` is convention only.
  **Where it lives in code:** `server/main.go` — `net.Listen("tcp", ":8080")`

- **Decision:** Use `status.Error(codes.X, ...)` instead of `fmt.Errorf` for handler errors.
  **Why:** `fmt.Errorf` returns get coerced to `codes.Unknown` by gRPC — the client can't branch on it. Status codes give clients a structured way to react.
  **Where it lives in code:** `server/main.go` — `CreatePipelineAndJobs` error returns.

- **Decision:** Enable server reflection (`reflection.Register(srv)`) for the dev workflow.
  **Why:** Lets grpcui/grpcurl introspect the server with no proto path config. Trivially small cost.
  **Where it lives in code:** `server/main.go` — line right after service registration.

---

## Rough Edges Hit

- **Symptom:** Method was named `createPipelineAndJobs` (lowercase c) — code compiled, but RPC returned `Unimplemented` to the client.
  **Root cause:** In Go, identifier case controls export. Lowercase methods don't satisfy an interface from another package. The embedded `UnimplementedPipelineServiceServer` silently satisfied the interface instead, so the compiler stayed quiet.
  **Fix:** Renamed to `CreatePipelineAndJobs`.
  **Lesson:** When implementing a generated interface, capitalize. The "silent fallback to Unimplemented" pattern is dangerous — no compile error, just a runtime `Unimplemented` response.

- **Symptom:** Proto `rpc CreatePipelineAndJobs(...) returns (CreatePipelineAndJobsRequest);` — return type was Request, not Response. Compiler complained when I tried to return `*Response`.
  **Root cause:** Typo in `pipeline.proto`. Generated code dutifully reflected it.
  **Fix:** Edited proto to `returns (CreatePipelineAndJobsResponse)`, re-ran `buf generate`.
  **Lesson:** Generated code is downstream of the schema — if Go signatures look wrong, fix the proto first, regenerate, don't hand-edit the generated code.

- **Symptom:** Returning `nil, nil` from the handler would cause a marshaling panic on the response.
  **Root cause:** gRPC requires either a non-nil response or a non-nil error. Both nil = no message to serialize.
  **Fix:** `return &pb.CreatePipelineAndJobsResponse{Pipeline: req.Pipeline}, nil`
  **Lesson:** "No error and no data" isn't a valid gRPC reply. Echo an empty struct at minimum.

- **Symptom:** Dead code — `panic(err)` after `log.Fatalf(...)`.
  **Root cause:** `log.Fatalf` calls `os.Exit(1)`; nothing after it runs.
  **Fix:** Deleted the `panic`.
  **Lesson:** `log.Fatalf` ≠ "log and continue". It terminates the process.

- **Symptom:** `fmt.Println("HELLO SERVER")` placed *after* `srv.Serve(listener)` — never printed.
  **Root cause:** `Serve` blocks until the server stops. Anything after it is effectively unreachable in normal operation.
  **Fix:** Moved the startup log *before* `Serve`. Wrapped `Serve` in an `if err := …; err != nil` to capture shutdown errors.
  **Lesson:** Print before you block.

---

## Concepts Learned

- **Struct embedding for forward compatibility** — `type Server struct { pb.UnimplementedPipelineServiceServer }`. Embedding the generated "unimplemented" struct ensures that if the proto adds new RPCs later, the server still compiles (new methods get a default `Unimplemented` response instead of a compile error). The generated `RegisterPipelineServiceServer` actually checks for this embedding at registration time.

- **The "register" call wires the routing table** — `pb.RegisterPipelineServiceServer(srv, &Server{})` is what tells the gRPC server "for incoming requests on `/pipeline.v1.PipelineService/*`, dispatch to this Go object". Without it, requests get `Unimplemented`.

- **Server reflection is a service-about-services** — `reflection.Register(srv)` adds the standard `grpc.reflection.v1.ServerReflection` service to the server. Clients (grpcui/grpcurl) call this at runtime to discover what RPCs and messages exist. Dev-time convenience; usually disabled in prod.

- **Metadata = gRPC's headers** — read on the server with `metadata.FromIncomingContext(ctx)`. Keys are always lowercased. Returns `[]string` per key because HTTP allows repeated headers. `authorization` is the canonical key for auth tokens.

- **Status codes are the contract** — `status.Error(codes.Unauthenticated, …)` etc. Use the right code so clients can react meaningfully. Common ones to know: `InvalidArgument`, `Unauthenticated`, `PermissionDenied`, `NotFound`, `AlreadyExists`, `Internal`, `Unavailable`. (Reference: https://grpc.io/docs/guides/status-codes/)

- **Functional options pattern** — `grpc.NewServer(opts...)` where each option is a function value. Didn't actually pass any in this session (the `var grpcOpts []grpc.ServerOption` line is currently empty/noise). Concept is parked for the next session.

- **Tooling: grpcui** — `grpcui -plaintext localhost:8080`. Uses server reflection by default; falls back to `-proto path/to/file.proto` if reflection is off. `-plaintext` because the server isn't doing TLS yet.

---

## Deferred — to revisit later

These came up but were intentionally not explored in this session. Tracking them so future-me doesn't forget.

- **`grpc.ServerOption` deep dive.** What options exist (`Creds`, `MaxRecvMsgSize`, `KeepaliveParams`, `UnaryInterceptor`, …), when each matters, defaults. Currently `var grpcOpts []grpc.ServerOption` is dead noise in `main.go`.

- **Interceptors (middleware).** Auth check is currently inlined in the handler. Should be extracted to a `UnaryServerInterceptor` so every RPC gets it for free. This is the most impactful next refactor — pulls duplication out of handlers.

- **Graceful shutdown.** Right now Ctrl-C kills in-flight RPCs. Need `os/signal` trapping + `srv.GracefulStop()`.

- **Deadlines / `ctx.Done()` handling.** Handler doesn't currently respect the request deadline. For long-running work (job execution), this matters.

- **TLS / credentials.** Currently plaintext (hence `-plaintext` on grpcui). Production needs `grpc.Creds(credentials.NewTLS(...))`. mTLS would be a good learning exercise.

- **Validation (protovalidate).** The `(google.api.field_behavior) = REQUIRED` annotations in the proto are documentation only — gRPC doesn't enforce them. `buf.build/bufbuild/protovalidate` would make those real.

- **Health check service (`grpc.health.v1`).** Standard endpoint that load balancers / k8s probes can hit.

- **Rich error details (`status.WithDetails`).** When validation fails, return a structured field list instead of a free-form string.

- **Observability — `StatsHandler` / OpenTelemetry.** Production tracing/metrics.

---

## What's Next

1. **Phase 2 setup:** add a server-streaming RPC (e.g., `StreamJobLogs`) to the proto, regenerate, implement on the server, hit from grpcui.
2. **Refactor:** extract the auth check into a `UnaryServerInterceptor`. Wire it via `grpc.ChainUnaryInterceptor(...)` as the first real use of a `ServerOption`.
3. **Hardening:** add graceful shutdown + deadline handling.
4. **TS client:** write the matching client call to `CreatePipelineAndJobs` — proves end-to-end across the language boundary, not just via grpcui.

---

## Open Questions

- *How does the embedded `UnimplementedPipelineServiceServer` actually satisfy the interface when my own method has a typo (wrong case)?* — Understood: embedding promotes the embedded methods to the outer struct. Both `Server` and `UnimplementedPipelineServiceServer` end up with method sets containing the exported method; the typoed lowercase method is a separate unrelated method. Resolved during this session — leaving the note as a reminder.

- *What's the actual difference between `grpc.UnaryInterceptor` and `grpc.ChainUnaryInterceptor` when there's only one interceptor?* — Deferred to the interceptor session.
