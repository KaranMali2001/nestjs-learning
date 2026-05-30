---
conversation_id: 82595e25-bf5f-4f62-a2d1-0772e6bc5b3d
date: 2026-05-30
project: go-ts-grpc (CI/CD Pipeline Runner)
phase: phase-1-unary
status: protovalidate wired as interceptor; empty/invalid CreatePipelineAndJobs requests now rejected with InvalidArgument before reaching the handler
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

## What's Next

1. **Verify end-to-end:** hit the server with grpcui sending empty / oversized / malformed payloads. Confirm `InvalidArgument` with the expected message, *before* reaching the handler.
2. **Rich error details:** wrap validation failures with `status.WithDetails(...)` so the client gets a structured list of field violations, not just a string.
3. **Phase 2:** server streaming (`StreamJobLogs`). Validation interceptor will need its `StreamServerInterceptor` counterpart added to the chain.
4. **AIP-133 follow-up:** consider splitting `Pipeline` into separate request/response shapes once a second mutating RPC is added.

---

## Open Questions

- *Should the validator skip server-set fields when the same message is used in both request and response?* — Today: yes, by omission. Long-term: split shapes per AIP-133. Decision deferred until a second RPC forces the issue.
- *How do `protovalidate` errors interact with gRPC's `status.Details`?* — The interceptor returns `InvalidArgument` with a flat message. There's a richer `Violations` proto inside protovalidate that could be surfaced via `status.WithDetails`. Worth a follow-up session.
- *Where exactly does the validator sit in the chain when streaming is added?* — Same position (between auth and recovery), but using `StreamServerInterceptor` from the middleware. To be confirmed in Phase 2.
