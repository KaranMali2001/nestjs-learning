# Session Index

Chronological table of contents for every learning/decision session on this project. One line per session — load this alone to get oriented, then jump into specific sessions for detail.

For convention/format details, see [README.md](README.md).
To start a new session, copy [_template.md](_template.md).

---

## Phase 0 — Bootstrap

| # | Date | Conversation ID | Topic | File |
|---|---|---|---|---|
| 01 | 2026-05-14 | `2c04098f-1ddf-489e-b9f9-1ec8238b6cc1` | Domain model, RPC surface, state machine, repo layout | [phase-0-bootstrap/01-design.md](phase-0-bootstrap/01-design.md) |
| 02 | 2026-05-18 | `fcd6abea-6372-4c94-9d09-9606101e286d` | buf config, codegen scaffold, monorepo layout (server/client own gen + deps) | [phase-0-bootstrap/02-setup.md](phase-0-bootstrap/02-setup.md) |

## Phase 1 — Unary RPC (CreatePipelineAndJobs)

| # | Date | Conversation ID | Topic | File |
|---|---|---|---|---|
| 01 | 2026-05-29 | `98b218a8-a0db-47f6-9cab-8115681a3a3b` | First server implementation: listener, registration, reflection, metadata, status errors | [phase-1-unary/01-server-impl.md](phase-1-unary/01-server-impl.md) |
| 02 | 2026-05-30 | `82595e25-bf5f-4f62-a2d1-0772e6bc5b3d` | Request validation with protovalidate + BeautifyValidationInterceptor — CEL rules, package-name collision, CEL `\n` escape trap, onion-not-pipeline interceptor wrapping order | [phase-1-unary/02-validation-protovalidate.md](phase-1-unary/02-validation-protovalidate.md) |

## Phase 2 — Server Streaming (StreamJobLogs)
*Not started.*

## Phase 3 — Client Streaming (BatchCreateJobs)
*Not started.*

## Phase 4 — Bidi (JobControlSession)
*Not started.*

## Phase 5 — Harden
*Not started.*
