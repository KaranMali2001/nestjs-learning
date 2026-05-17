# Phase 0 — Bootstrap

Pre-code sessions: design discussions and dev-environment scaffolding.

| # | Conversation | Outcome |
|---|---|---|
| 01 | `2c04098f-1ddf-489e-b9f9-1ec8238b6cc1` | Domain model, RPC surface, state machine, repo layout decided |
| 02 | `fcd6abea-6372-4c94-9d09-9606101e286d` | buf + codegen working; server/client each own gen/ and deps |

**Exit criteria** (when this phase closes):
- `buf generate` produces both Go and TS code without errors ✅
- `go.mod` initialized inside `server/`, `tsconfig.json` inside `client/`
- Empty `server/main.go` and `client/src/main.ts` files committed
- Ready to write the first RPC handler → enters phase-1-unary
