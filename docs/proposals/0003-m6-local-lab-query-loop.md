# Proposal 0003: M6 Local Lab Query Loop

## Status

Accepted for M6 implementation

## Context

M1-M5 prove the write path: review task submission, context governance, policy routing, LiteLLM call, and task-level usage logging.
The project still needs a simple local read loop so a human or agent can inspect recent usage without opening JSONL files manually.

## Proposed Changes

- Add read-only usage endpoints backed by the existing append-only JSONL file.
- Add a small local smoke script that demonstrates review task submission, task lookup, and usage lookup.
- Update roadmap, current state, and CI status as M6 progresses.

## Non-goals

- Do not add a database.
- Do not add a dashboard.
- Do not add K8s, OTel, or distributed tracing.
- Do not change ContextManager or PolicyRouter behavior.
- Do not replace LiteLLM accounting.

## Acceptance

- `GET /usage/recent` returns recent usage records.
- `GET /usage/tasks/{id}` returns usage records for one task id.
- `POST /tasks/review-diff` and usage lookup can be verified in one local smoke flow.
- Real provider E2E and Docker Compose runtime E2E pass after implementation.
