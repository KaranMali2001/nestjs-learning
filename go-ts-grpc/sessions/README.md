# Sessions

Chronological record of every AI conversation that shaped this project. Each session = one conversation = one markdown file.

## Layout

```
sessions/
  index.md            — TOC of every session across all phases
  _template.md        — copy this to start a new session
  phase-N-<name>/
    README.md         — what this phase is about + exit criteria
    NN-<topic>.md     — one file per conversation
```

Phase folders mirror the 5 phases in `../README.md`. Phase 0 (bootstrap) covers pre-code work.

## File naming

`NN-<topic>.md` where `NN` is a 2-digit number, incrementing within the phase (not globally). Example: `phase-1-unary/03-status-codes-debug.md`.

## Frontmatter — required

Every session file starts with:

```yaml
---
conversation_id: <claude-session-uuid>   # e.g. fcd6abea-6372-4c94-9d09-9606101e286d
date: YYYY-MM-DD
project: go-ts-grpc (CI/CD Pipeline Runner)
phase: phase-N-<name>
status: <one-line current state>
---
```

`conversation_id` is **Claude Code's session UUID** — the durable handle that lets you re-find the original chat transcript on disk.

**To find the current session's UUID:**

```
ls -lt ~/.claude/projects/-Users-karanmali5599-Desktop-Projects-Learnings/*.jsonl | head -1
```

The most recently modified `.jsonl` file in that folder is your active session. The filename (without `.jsonl`) is the conversation ID. Copy that into the frontmatter.

## Body sections — standard four

1. **Decisions Made** — what was chosen, why, where it lives in code
2. **Rough Edges Hit** — symptom, root cause, fix, lesson
3. **Concepts Learned** — one-liners, promote big ones to `docs/concepts.md`
4. **What's Next** + **Open Questions** — so the next session picks up cold

Use [_template.md](_template.md) — it has these sections pre-stubbed.

## Why this folder exists

These aren't tutorials or polished docs. They're a learning trail — the actual back-and-forth where decisions got argued through, mistakes got made, and reasoning got recorded.

Future-me (or anyone picking up this repo cold) reads sessions in order to understand **why** the code looks the way it does, not just **what** it does.

For synthesized stable knowledge (extracted *from* sessions, organized by topic), see [../docs/](../docs/).

## Workflow when starting a new chat

1. Identify the phase you're in.
2. Open the phase folder's README to remind yourself of the goal.
3. Copy `_template.md` to `phase-N-<name>/NN-<topic>.md`.
4. Fill the frontmatter at the start of the chat (status will keep updating).
5. As decisions / rough edges happen, log them inline.
6. At session end: update `status:`, append to `index.md`.
