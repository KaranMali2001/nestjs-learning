---
conversation_id: <claude-session-uuid>   # e.g. fcd6abea-6372-4c94-9d09-9606101e286d
                                         # find: ls -lt ~/.claude/projects/-Users-karanmali5599-Desktop-Projects-Learnings/*.jsonl
                                         # the most recently modified file = current session
date: YYYY-MM-DD
project: go-ts-grpc (CI/CD Pipeline Runner)
phase: phase-N-<name>                # e.g. phase-1-unary
status: <one-line current state>
---

# <Session Title>

One paragraph: what this session was about and what triggered it (a blocker? a design question? a code review?).

---

## Decisions Made

- **Decision:** _what was chosen_
  **Why:** _rationale, alternatives considered, what tipped the balance_
  **Where it lives in code:** _file paths / commits / proto fields_

(Repeat per decision. Skip if none.)

---

## Rough Edges Hit

- **Symptom:** _what broke or confused me_
  **Root cause:** _what was actually wrong (not just the surface fix)_
  **Fix:** _exact change_
  **Lesson:** _what to remember next time_

(Repeat. This section is the most valuable for future-me.)

---

## Concepts Learned

- _Concept name_ — one-line summary. If big enough, promote to `docs/concepts.md` and link.

---

## What's Next

1. _next concrete action_
2. _then..._
3. _eventual goal_

---

## Open Questions

- _question_ — _why deferred / what would unblock it_
