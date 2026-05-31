# gRPC Concepts — Learning Notes

## What is gRPC

- Built on HTTP/2 — persistent connection, one TCP handshake, all RPCs reuse it
- Uses Protocol Buffers — binary format, smaller than JSON, fields identified by numbers not names
- Primarily server-to-server communication (browser needs grpc-web + proxy)
- Handles back pressure via flow control

---

## HTTP/2 Stream vs gRPC Stream

Two different things sharing the same word.

**HTTP/2 stream** — transport level. One TCP connection, multiple logical lanes running concurrently. Each lane = one stream with a unique ID. One gRPC call = one HTTP/2 stream.

**gRPC stream** — application level. Defined in `.proto` file. Means "this RPC sends/receives multiple messages" instead of one.

```
File upload client streaming in gRPC
→ one HTTP/2 stream (the pipe)
→ many gRPC messages (the water flowing through it)
```

---

## The 4 RPC Types

| Type | Direction | Use case |
|---|---|---|
| Unary | 1 req → 1 res | Classic request/response |
| Server Streaming | 1 req → N res | Live feed, logs, notifications |
| Client Streaming | N req → 1 res | File upload, batch data |
| Bidirectional | N req → N res | Chat, real-time control |

---

## Back Pressure / Flow Control

Two layers running simultaneously:

```
gRPC message
    ↓
HTTP/2 flow control window  ← per-stream + per-connection (65KB default)
    ↓
TCP sliding window          ← OS level, per-connection
    ↓
Network
```

- TCP prevents OS buffer overflow at transport level
- HTTP/2 window gives per-stream control — one slow stream doesn't starve others
- When receiver is slow → sender's write call BLOCKS (no data loss, no dropping)
- Automatic by default. Manual control via `isReady()`, `setOnReadyHandler()`

**Simulating back pressure:** Add `time.Sleep(100ms)` per message on server, stream 1000 messages from client. Watch client writes start blocking.

---

## Head-of-Line Blocking

HTTP/2 runs multiple streams on ONE TCP connection. If a single packet drops, TCP freezes ALL streams while retransmitting. This is why HTTP/3 (QUIC/UDP) exists — packet loss on one stream doesn't affect others.

---

## Metadata

gRPC's version of HTTP headers. Key-value pairs that travel alongside RPCs. Under the hood they ARE HTTP/2 headers.

**Headers** — sent before messages (auth tokens, trace IDs)
**Trailers** — sent AFTER all messages (final status, error details). HTTP can't do this — gRPC can via HTTP/2 `END_STREAM` frame.

```
HTTP:   Headers → Body
gRPC:   Headers → Messages → Trailers
```

Rules:
- Keys: ASCII, case-insensitive, cannot start with `grpc-`
- Binary values: key must end with `-bin`
- Default server limit: 8KB for request headers

---

## Interceptors

Middleware for gRPC. 4 types:

```
                CLIENT SIDE        SERVER SIDE
Unary RPC       Unary Client       Unary Server
Streaming RPC   Stream Client      Stream Server
```

Execution model:
```
Request →  [interceptor 1]  →  [interceptor 2]  →  Handler
Response ← [interceptor 1]  ← [interceptor 2]  ←  Handler
```

Order matters — auth interceptor before logging = only authenticated requests get logged.

Common uses: auth, logging, tracing, metrics, panic recovery, rate limiting.

### Wrapping order — interceptors compose as an onion, not a pipeline

The word "chain" is misleading. `ChainUnaryInterceptor(A, B, C)` doesn't run A → B → C in sequence; it wraps them: A *calls* B, which *calls* C, which *calls* the handler. Element N+1 in the chain is element N's `handler` parameter.

```
ChainUnaryInterceptor(A, B, C):

A(ctx, req, info, handler=B-bound) {
   resp, err := handler(ctx, req)   // ← this *is* B
   ...
}

B(ctx, req, info, handler=C-bound) {
   resp, err := handler(ctx, req)   // ← this *is* C
   ...
}

C(ctx, req, info, handler=real-handler) {
   resp, err := handler(ctx, req)   // ← real RPC handler
   ...
}
```

**Consequence — to *wrap and reshape* another interceptor's error, your interceptor must come *before* it in the chain, not after.**

If B short-circuits (returns an error without calling its handler), then C never runs. So if you want to repackage B's error, you need to be A (so B is *your* handler call). Putting your wrapper *after* B in the chain means it sits inside B and never runs when B short-circuits.

This is the same pattern as Express/Koa middleware, Connect handlers, Python WSGI middleware — anything described with the word "chain" is actually nested wrapping. The mental shortcut "decorator" is more accurate than "pipeline."

Concrete example from this project: `BeautifyValidationInterceptor` is placed *before* the protovalidate middleware in `ChainUnaryInterceptor` because it needs to wrap the middleware's call — so when the middleware returns an `InvalidArgument` error, Beautify sees it on the way out and can repackage the `Violations` detail.

---

## Context — The Most Important Concept

In HTTP everything is in `req`. In gRPC it's split:

```go
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) {}
//                        ↑ envelope                ↑ letter
```

- `req` — the protobuf payload (your data)
- `ctx` — metadata + deadline + cancellation signal

**Never create a fresh context mid-chain.** Always pass `ctx` downstream or you lose:
- Auth tokens
- Trace IDs
- **The deadline** ← most dangerous to lose

---

## Deadlines

**Timeout** = a duration ("5 seconds")
**Deadline** = absolute point in time ("2:30:05 PM")

gRPC uses deadlines internally. Timeout → converted to `now + N` → propagated as absolute deadline.

```
Client sets deadline → now + 5s = 2:30:05 PM
Service A receives   → 3s remaining
Service B receives   → 2s remaining  (same absolute deadline, gRPC calculates remaining)
Service C receives   → 1s remaining
```

**What happens on expiry:**
- Client gets `DEADLINE_EXCEEDED`
- Server context gets cancelled
- BUT — your code must check `ctx.Done()` or spawned goroutines keep running

**Node.js is manual** — you must pass deadline explicitly:
```typescript
// Option 1: manual
childClient.doWork({ task }, { deadline: parentCall.getDeadline() }, cb)

// Option 2: propagate flags
childClient.doWork({ task }, { parent: parentCall, propagate_flags: grpc.propagate.DEFAULTS }, cb)
```

Go and Java propagate automatically when you pass `ctx`.

---

## Status Codes (17 total)

### Framework generates these — never return from your code
| Code | When |
|---|---|
| `OK` | Success |
| `CANCELLED` | Client cancelled |
| `DEADLINE_EXCEEDED` | Deadline passed |
| `UNKNOWN` | Error with no info |
| `INTERNAL` | gRPC internal error |
| `UNAVAILABLE` | Server down, safe to retry |
| `UNAUTHENTICATED` | No valid credentials |

### You return these from application code
| Code | When |
|---|---|
| `NOT_FOUND` | Resource doesn't exist |
| `ALREADY_EXISTS` | Creating something that exists |
| `PERMISSION_DENIED` | Authenticated but not allowed |
| `INVALID_ARGUMENT` | Bad input |
| `RESOURCE_EXHAUSTED` | Rate limit, quota exceeded |
| `FAILED_PRECONDITION` | System state wrong for operation |
| `ABORTED` | Concurrency conflict |
| `OUT_OF_RANGE` | Valid type, value out of bounds |
| `UNIMPLEMENTED` | Method not implemented |
| `DATA_LOSS` | Unrecoverable corruption |

### The confusing three
```
age = -5             → INVALID_ARGUMENT   (bad input regardless of state)
delete non-empty dir → FAILED_PRECONDITION (valid input, wrong system state)
page 10 of 3 pages   → OUT_OF_RANGE       (valid type, exceeds bounds)

no token sent        → UNAUTHENTICATED    (who are you?)
token valid, wrong role → PERMISSION_DENIED (I know you, but no)
```

### Retry rules
```
Safe to retry:   UNAVAILABLE, RESOURCE_EXHAUSTED
Never retry:     INVALID_ARGUMENT, NOT_FOUND, PERMISSION_DENIED, ALREADY_EXISTS
```

---

## Validation

gRPC ships **no built-in validator**. proto3 has no wire-level `required`, and `(google.api.field_behavior) = REQUIRED` (AIP-203) is documentation only — Google's spec explicitly says it adds no validation. Two ecosystem options:

| | `protoc-gen-validate` (PGV) | **`protovalidate`** ✅ |
|---|---|---|
| Status | Archived | Current (Buf, 2026) |
| Custom rules | Limited | **CEL** expressions in proto |
| Languages | Go, others | Go, Java, Python, C++, JS/TS |

**Two Go packages with the same name** — alias one on import:
- `buf.build/go/protovalidate` — core validator: `protovalidate.New()`, `Validate()`
- `github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate` — gRPC middleware: `UnaryServerInterceptor`, `StreamServerInterceptor`

Wiring:

```go
validator, _ := protovalidate.New()
grpc.ChainUnaryInterceptor(
    LoggingInterceptor,
    AuthInterceptor,
    protovalidateinterceptor.UnaryServerInterceptor(validator), // auth → validate → handler
    RecoveryInterceptor,
)
```

Failures return `codes.InvalidArgument` with the violation message — handler never sees a bad payload.

**Annotations cheat-sheet:**
```proto
string name = 2 [(buf.validate.field).string = {min_len: 1, max_len: 100}];
repeated Job jobs = 5 [(buf.validate.field).repeated = {min_items: 1, max_items: 50}];
int32 timeout_seconds = 8 [(buf.validate.field).int32 = {gt: 0, lte: 3600}];
CodeLanguage language = 1 [(buf.validate.field).enum = {defined_only: true, not_in: [0]}];
Pipeline pipeline = 1 [(buf.validate.field).required = true];

oneof step {
    option (buf.validate.oneof).required = true;
    CommandStep command = 5;
    CodeStep code = 6;
}

// CEL custom rule
string name = 2 [(buf.validate.field).cel = {
    id: "pipeline.name.not_blank",
    message: "name must not be blank",
    expression: "this.trim().size() > 0"
}];
```

**`field_behavior` vs `buf.validate.field`** — complementary, not substitutes. Keep `field_behavior` for documentation (especially `OUTPUT_ONLY`); add `buf.validate.field` for enforcement.

**OUTPUT_ONLY trap** — if a single message is reused for both request and response, validating "empty on input" gets awkward. AIP-133 says: split into `CreateXInput` and `X`. For learning projects, skip validation on output-only fields.

---

## Real Company Lessons

**Uber** — migrated push notifications to gRPC bidi. 45% drop in p95 latency. Key lesson: had to manually handle flow control during message bursts.

**Dropbox** — built Courier framework ON TOP of gRPC. Added mandatory deadlines, circuit breaking, per-method ACLs, mTLS. Raw gRPC isn't enough at scale.

**Datadog** — default gRPC load balancing sends all traffic to ONE pod in Kubernetes. Silent killer. Broken TCP connections take 15 minutes to detect without keepalive tuning.

---

## Things Not Yet Explored

- Health checking
- Keepalive
- Retry policies + hedging
- Load balancing
- TLS / mTLS
- Reflection + grpcdebug
- Channelz
- OpenTelemetry metrics
- xDS / service mesh
- Graceful shutdown
- Compression
