# Proposal 0005: Unify M5-M8 under Invocation Ledger

## Status

Accepted

## Context

M5 introduced usage writes, M6 added usage read APIs, and M7 added failure replay output.
Those were useful milestones, but they split one observability concept across separate surfaces.

The repeated fields are the same execution facts:

- command or scenario
- status
- task id
- request id
- latency
- selected model
- returned model
- token usage
- failure kind

## Proposal

Create M8 as the consolidation milestone:

- Add ADR-0003 for Invocation Ledger as the observability boundary.
- Add Spec 0008 for a unified `invocation.Record`.
- Keep `usage.Record`, but wrap it as the `usage` projection inside `invocation.Record`.
- Move source of truth to `data/invocations.jsonl`.
- Keep `/usage/*` only as compatibility projections over `/invocations/*`.
- Update M5-M7 specs so they are no longer independent architecture boundaries.

## Consequences

- Router, ContextManager, Usage, Replay, Debug, and Evaluation share one execution record model.
- Future milestones compare behavior using ledger records instead of adding another partial record shape.
- M9 and M10 become evaluation milestones, not new advanced behavior milestones.

## Non-goals

- No database.
- No streaming.
- No fallback implementation.
- No K8s.
- No OTel.
- No advanced Router or ContextManager changes in M8.
