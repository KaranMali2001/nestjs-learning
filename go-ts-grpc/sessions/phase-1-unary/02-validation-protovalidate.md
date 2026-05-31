---
conversation_id: 82595e25-bf5f-4f62-a2d1-0772e6bc5b3d
date: 2026-05-30
project: go-ts-grpc (CI/CD Pipeline Runner)
phase: phase-1-unary
status: protovalidate wired + BeautifyValidationInterceptor repackages Violations into google.rpc.BadRequest; per-field server log on failure; CEL `\n` escape fixed
---

# Phase 1.02 — Request Validation with Protovalidate

Started from a working but unprotected handler — sending an empty `CreatePipelineAndJobs` request via grpcui sailed straight through to the handler. The `(google.api.field_behavior) = REQUIRED` annotations turned out to be documentation only. This session investigated what the official gRPC/Buf story for validation actually is in 2026, picked one, and wired it up end-to-end as a server interceptor.

---

## Decisions Made

- **Decision:** Use [protovalidate](https://protovalidate.com/) (buf, `buf.build/go/protovalidate` v1.x) instead of the older `protoc-gen-validate` (PGV).
  **Why:** PGV is archived. The Buf team explicitly recommends migrating to protovalidate, which uses Google's CEL for custom rules, ships first-class Go support, and works cross-language. The contract lives in the `.proto` — so the TS client sees the same rules without re-implementing them.
  **Where it lives in code:** `proto/pipeline/v1/pipeline.proto` (annotations), `server/main.go` (interceptor wiring), `buf.yaml` (BSR dep).

- **Decision:** Keep `(google.api.field_behavior)` annotations alongside `(buf.validate.field)` instead of stripping them.
  **Why:** They serve different purposes. `field_behavior` is AIP-203 documentation (specifically valuable for `OUTPUT_ONLY` markers, which protovalidate doesn't model). `buf.validate.field` is enforcement. They coexist cleanly.
  **Where it lives in code:** every annotated field in `pipeline.proto` has both.

- **Decision:** Don't validate `OUTPUT_ONLY` fields (e.g. `Pipeline.id`, `Job.status`, `started_at`). "Option 1" from the trade-off list.
  **Why:** The same `Pipeline` message is used in both the request and the response — `id` must be empty on input but present on output. Validating empty-on-input via CEL is doable but bloats the proto. Splitting into separate request/response messages (AIP-133) is cleaner long-term but overkill for a learning project. Accepting that the server will ignore client-supplied output-only fields is the pragmatic call.
  **Where it lives in code:** absence of `buf.validate.field` on `OUTPUT_ONLY` fields in `pipeline.proto`.

- **Decision:** Interceptor order — `Logging → Auth → protovalidate → Recovery`.
  **Why:** Auth before validation means unauthenticated callers don't learn anything about validation rules. Validation before the handler means malformed payloads never burn CPU on business logic. Recovery innermost converts panics to `codes.Internal` before outer interceptors see them.
  **Where it lives in code:** `server/main.go` — `grpc.ChainUnaryInterceptor(...)`.

- **Decision:** Alias the gRPC middleware package on import as `protovalidateinterceptor`.
  **Why:** Both `buf.build/go/protovalidate` and `github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate` use the package name `protovalidate`. Without aliasing, the IDE autocomplete merges symbols from both into one fuzzy list (`UnaryServerInterceptor`, `WithExtensionTypeResolver`, `New`…) and you can't tell which package owns which.
  **Where it lives in code:** `server/main.go` imports.

- **Decision:** Forbid newlines in `CommandStep.command` (CEL: `!this.contains('\n')`).
  **Why:** A `CommandStep` is meant to be a single shell line. Multi-line shell scripts belong in a `CodeStep`. Enforcing this in the contract instead of in the handler keeps it visible to all clients.
  **Where it lives in code:** `pipeline.proto` — `CommandStep.command`.

---

## Rough Edges Hit

- **Symptom:** Empty `CreatePipelineAndJobs` request via grpcui hit the handler (which then panicked on `req.Pipeline == nil`). Expected the `REQUIRED` annotation to reject it at the boundary.
  **Root cause:** `(google.api.field_behavior) = REQUIRED` is AIP-203 documentation — Google's own spec says "the vocabulary in AIP-203 is for descriptive purposes only and does not itself add any validation." Proto3 has no wire-level concept of required either; that was removed from proto2 on purpose. No code is generated; nothing enforces it.
  **Fix:** Switch to `(buf.validate.field).required = true` (and friends), which protovalidate's interceptor actually reads at runtime.
  **Lesson:** `field_behavior` annotations document intent for humans and SDK generators. They do not, by design, run any code. If you want enforcement, you need protovalidate (or write the checks yourself).

- **Symptom:** IDE autocomplete on `protovalidate.` showed `UnaryServerInterceptor`, `StreamServerInterceptor`, **and** `WithExtensionTypeResolver` together, but only some of them compiled.
  **Root cause:** Two distinct Go packages are both named `protovalidate`:
    - `buf.build/go/protovalidate` — the core validator (`New`, `Validate`, `WithExtensionTypeResolver`).
    - `github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate` — the gRPC middleware (`UnaryServerInterceptor`, `StreamServerInterceptor`, `WithIgnoreMessages`).
  Without an import alias, calls like `protovalidate.UnaryServerInterceptor(...)` resolve to whichever package was imported under that name — and the autocomplete blends symbols from any imports it's seen.
  **Fix:** `import protovalidateinterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"` and leave `buf.build/go/protovalidate` unaliased.
  **Lesson:** When two third-party packages share a name, alias one on import. Don't fight the autocomplete; rename the symbol so it's obvious which package owns what.

- **Symptom:** `go.mod` flagged `buf.build/go/protovalidate` as `// indirect` even though we imported it directly.
  **Root cause:** `go mod tidy` had not been re-run since the import was added; it had only been pulled in transitively before.
  **Fix:** `go mod tidy` — promotes it to a direct dependency.
  **Lesson:** Always run `go mod tidy` after adding a new import; the `// indirect` marker is auto-managed and the diagnostic catches it.

---

## Concepts Learned

- **`field_behavior` vs `buf.validate.field`** — `field_behavior` is documentation; `buf.validate.field` is enforcement. They are complementary, not substitutes.
- **CEL (Common Expression Language)** — protovalidate's secret weapon. Define custom rules inline in the proto: `expression: "this.trim().size() > 0"`, `expression: "!this.contains('\n')"`. Compiled at validator construction time, evaluated per request. Same CEL runs in every language's protovalidate implementation, so the rule is portable.
- **Protovalidate interceptor returns `InvalidArgument`** — the middleware automatically converts validation failures into `codes.InvalidArgument` with a descriptive message. No handler code change needed; existing handlers just stop seeing bad payloads.
- **`oneof.required` on a `oneof`** — the protovalidate way to say "you must pick one branch of this oneof". On `Job.step`: `option (buf.validate.oneof).required = true;`. Without it, sending neither branch would silently produce a Job with no step.
- **`enum.defined_only` + `not_in: [0]`** — common combo for proto3 enums: reject the implicit zero (`*_UNSPECIFIED`) value plus anything not declared in the enum. Both checks; both matter.
- **AIP-133 (separate request/response shapes)** — if `OUTPUT_ONLY` fields cause request-side validation pain, the canonical Google answer is to split the message: a `CreatePipelineInput` (no OUTPUT_ONLY fields) that the server turns into a fully-populated `Pipeline` in the response. Deferred for now.

---

## Validation rules added (snapshot)

For future-me — what's enforced after this session:

| Field | Rule |
|---|---|
| `Pipeline.name` / `Job.name` | 1–100 chars + not-blank CEL (`this.trim().size() > 0`) |
| `Pipeline.description` / `Job.description` | ≤ 400 chars |
| `Pipeline.jobs` | 1–50 items |
| `Job.step` (oneof) | `oneof.required = true` |
| `Job.timeout_seconds` | `gt: 0, lte: 3600` |
| `CommandStep.command` | 1–4096 chars + no newlines |
| `CodeStep.language` | `defined_only: true, not_in: [0]` |
| `CodeStep.code` | 1–100000 chars |
| `CreatePipelineAndJobsRequest.pipeline` | `required = true` (non-nil) |

---

## Addendum — same conversation, after first round of testing

End-to-end testing surfaced three things that needed fixing on top of the initial wiring. All resolved within this session.

### Additional Decisions Made

- **Decision:** Add a `BeautifyValidationInterceptor` *outside* the protovalidate middleware in the chain rather than replacing the middleware.
  **Why:** The middleware already attaches the structured `Violations` proto as a status detail via `status.New(...).WithDetails(valErr.ToProto())` (confirmed by reading its source). The structured info is not lost — a downstream interceptor can pull it out of `status.Convert(err).Details()` and reshape it. Wrapping (not replacing) preserves the middleware's job, gets future upgrades for free, and keeps single-responsibility per interceptor.
  **Where it lives in code:** `server/validation_beautify_interceptor.go`, and the chain order in `server/main.go` — `Beautify` is placed **before** the middleware in `ChainUnaryInterceptor` so that on the way out it sees the middleware's error.

- **Decision:** Convert protovalidate's `Violations` into the standard `google.rpc.BadRequest` shape via `status.WithDetails`.
  **Why:** `BadRequest` + `FieldViolation` is the well-known gRPC error contract that real clients (TS via `@grpc/grpc-js` + `errdetails`, Go via `errdetails`) decode out of the box. Reshaping once at the boundary means handlers and downstream code never have to know about protovalidate's specific proto.
  **Where it lives in code:** loop inside `BeautifyValidationInterceptor` building `errdetails.BadRequest_FieldViolation` entries.

- **Decision:** Log per-field violations **inside** `BeautifyValidationInterceptor` (`field=... rule=... msg=...`) rather than creating a separate `ErrorLoggingInterceptor`.
  **Why:** Beautify already iterates the structured violations on its way through. A separate downstream interceptor would have to re-parse the error it just packaged — and would only ever see the squashed `"validation failed"` string, having lost the field detail. One interceptor doing transform + log on the same in-hand data is the right granularity; if a generic "log any error" emerges later, that belongs *in* `LoggingInterceptor`, not a third file.
  **Where it lives in code:** `fmt.Printf("[validation] ... — N violation(s)")` plus per-violation lines inside the violations loop.

- **Decision:** Escape `\n` inside CEL expressions as `\\n` in the proto source (or `\\x0A` per official docs).
  **Why:** Proto string literals process escape sequences *before* CEL parses the expression. Writing `'\n'` in proto produces a *literal* newline in the string handed to CEL, which then sees `!this.contains('<actual newline>')` — i.e. a CEL string literal split across lines — and fails compilation with a parser syntax error. `\\n` in proto becomes the two-character sequence `\n`, which CEL correctly parses as a newline escape.
  **Where it lives in code:** `CommandStep.command` CEL rule in `pipeline.proto` — `expression: "!this.contains('\\n')"`.

### Additional Rough Edges Hit

- **Symptom:** Sending a syntactically valid payload returned `code=Internal` (caught by `LoggingInterceptor`) — no `PANIC in ...` line from `RecoveryInterceptor`, so it was not a Go panic. Could not tell *what* was wrong because `LoggingInterceptor` only printed the code, not the message.
  **Root cause:** The CEL expression on `CommandStep.command` was malformed. Protovalidate lazily compiles CEL on first use of each message type; compilation failed with a parser error, which the middleware returns as `Internal` (this is *not* a validation failure, it is the validator itself failing). The actual error was hidden because logging didn't surface `err.Error()`.
  **Fix:** First — surface the error in `LoggingInterceptor` (added `fmt.Println("LOGGING ERROR", err)`). Then read the actual message, which revealed the CEL parser was choking on a literal newline character inside a string literal. Re-escaped as `\\n` in the proto.
  **Lesson:** *Always log the error message, not just the code.* A `code=...` line alone is debugging by guesswork. Print `err.Error()` (or `status.Convert(err).Message()`) whenever `err != nil`. Also: protovalidate failure modes split into two categories — *your data is invalid* → `InvalidArgument` (good), and *your rules are broken* → `Internal` (configuration error, never a request-level issue). `Internal` from the validator never means the user did something wrong.

- **Symptom:** After adding `BeautifyValidationInterceptor`, grpcui showed `"@error": "google.rpc.BadRequest is not recognized; see @value for raw binary message data"` even though the wire payload was correct.
  **Root cause:** grpcui decodes error-detail `Any` values by looking up the type via server reflection. It only finds types belonging to *your* service's transitive imports — `google.rpc.BadRequest` lives in `genproto/googleapis/rpc/errdetails`, which the server imports in Go but doesn't expose through the proto schema, so reflection has no descriptor for it. The payload itself was fine — real clients with `errdetails` imported decode it without issue.
  **Fix:** None applied this session. Discussed two options: (a) define `ValidationFailure` + `FieldViolation` messages in `pipeline.v1` so reflection picks them up; (b) accept grpcui's limitation and rely on real-client decoding. User chose (b) for now — the on-wire contract is what production traffic uses, and the server-side log already provides full dev visibility.
  **Lesson:** grpcui ≠ real clients. It is a reflection-driven dev tool; well-known external types in error details won't render unless you bring their descriptors into your service's reflection surface (typically by defining your own equivalent in your proto package). Decide whether you optimise for grpcui (your own types) or for cross-service consistency (errdetails). For learning projects, server logs cover dev visibility either way.

- **Symptom:** `code=Internal` mystery hung around longer than necessary because `LoggingInterceptor` only printed `code` and `duration`.
  **Root cause:** Same as above — silent logger. Worth calling out separately as a design lesson, not just a debug-driver.
  **Fix:** Log `err.Error()` and (where applicable) `status.Convert(err).Details()` to surface attached protos.
  **Lesson:** A logging interceptor that doesn't log the error body is barely an interceptor. Default to logging everything you have access to on the error path; trim later if it gets noisy. The asymmetry is right: success path = code + duration, error path = code + duration + message + details.

### Additional Concepts Learned

- **Interceptor wrapping order in `ChainUnaryInterceptor`** — interceptors form an onion. Element N+1 in the chain is *called by* element N (it is element N's `handler`). To **wrap** another interceptor's behavior (e.g. observe or reshape its error), your interceptor must come *before* the target in the chain, not after. Beautify *before* protovalidate-middleware = Beautify wraps it.
- **Lazy CEL compilation in protovalidate** — `protovalidate.New()` returns without ever touching CEL. CEL expressions are compiled per message type on the first `Validate(msg)` call. Implications: (1) startup is fast, (2) CEL errors don't surface until runtime, (3) test every CEL rule at least once before deployment or you'll discover the parse error in prod.
- **`status.Convert` vs `status.Code` vs `status.FromError`** — `Code(err)` returns just the code (`OK` for nil, `Unknown` for non-status errors). `Convert(err)` always returns a non-nil `*Status` (synthesises `Unknown` if the error isn't a real gRPC status). `FromError(err)` returns `(*Status, bool)` where the bool tells you whether the error was a real gRPC status. Rule of thumb: `Convert` for logging (safe + complete), `FromError` when the bool changes behavior, `Code` when you only need the code.
- **`google.rpc.BadRequest` / `FieldViolation` are the gRPC-standard error contract** — defined in `google/rpc/error_details.proto`, exposed in Go as `google.golang.org/genproto/googleapis/rpc/errdetails`. Used by Google's own APIs (Cloud, AIP). If you don't have a strong reason to invent your own error-detail proto, use these.
- **Two-source-of-truth pitfall when a `.proto` schema changes:** Go server reads the *compiled* `pipeline.pb.go` for descriptors, including the bytes of CEL expressions inside annotations. Editing `pipeline.proto` without running `buf generate` means the validator still sees the old CEL. After every CEL/annotation edit: regen first, then restart.

### Updated What's Next

1. **Verify the new BadRequest payload in a real client** — once the TS client is wired (Phase 2), confirm `@grpc/grpc-js` + `errdetails` decodes the field violations without manual unmarshalling. This closes the loop on whether the choice to keep `errdetails.BadRequest` (over a self-defined `ValidationFailure`) holds up in practice.
2. **Phase 2 (server streaming):** the same protovalidate middleware ships a `StreamServerInterceptor`. The Beautify wrapper does not — it would need a streaming counterpart that intercepts the final stream-close error.
3. **Optional:** if grpcui dev experience matters, swap to a self-defined `ValidationFailure` message in `pipeline.v1` so reflection covers it. Cost: one new message + Beautify uses it instead of `errdetails.BadRequest`. Benefit: grpcui renders the violations natively.
4. **AIP-133 follow-up:** still on the shelf — revisit when a second mutating RPC lands.

### Open Questions — revised

- *Is keeping `errdetails.BadRequest` worth the grpcui dev-UX hit?* — depends on whether the TS client (Phase 2) decodes cleanly. Tracking.
- *Should `LoggingInterceptor` walk `status.Convert(err).Details()` and dump them too?* — would cover *all* future error-detail types automatically (not just validation). Cheap to add, but currently Beautify covers the validation case explicitly. Decision deferred.
- *What happens to the Beautify wrapper when the request is a stream?* — `StreamServerInterceptor` has a very different shape (works on the stream object, not request/response pairs). To be investigated in Phase 2.

---

## Open Questions

- *Should the validator skip server-set fields when the same message is used in both request and response?* — Today: yes, by omission. Long-term: split shapes per AIP-133. Decision deferred until a second RPC forces the issue.
- *How do `protovalidate` errors interact with gRPC's `status.Details`?* — The interceptor returns `InvalidArgument` with a flat message. There's a richer `Violations` proto inside protovalidate that could be surfaced via `status.WithDetails`. Worth a follow-up session.
- *Where exactly does the validator sit in the chain when streaming is added?* — Same position (between auth and recovery), but using `StreamServerInterceptor` from the middleware. To be confirmed in Phase 2.
