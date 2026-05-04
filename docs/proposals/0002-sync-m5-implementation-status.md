# Proposal 0002: Sync M5 Implementation Status

## Status

Accepted for documentation sync

## Context

M5 UsageLogger has been implemented and verified with a real provider-backed end-to-end smoke test.
The formal roadmap, UsageLogger spec, and UsageLogger test plan need to reflect what is implemented and what remains blocked.

## Proposed Changes

- Update `docs/roadmap.md` so M5 is marked in progress instead of next.
- Clarify that append-only JSONL usage logging is implemented.
- Clarify that full Docker Compose runtime verification remains blocked until Docker daemon is available.
- Update UsageLogger spec and test plan with current implementation and verification status.

## Non-goals

- Do not change ADR-0001 or ADR-0002.
- Do not add Router, ContextManager, K8s, OTel, or gateway behavior beyond current milestones.
- Do not replace LiteLLM usage accounting.

## Acceptance

- A future agent can distinguish completed M5 UsageLogger work from pending Docker Compose runtime verification.
- The next step after this sync is either committing M5 or rerunning full Compose verification when Docker Desktop is available.
