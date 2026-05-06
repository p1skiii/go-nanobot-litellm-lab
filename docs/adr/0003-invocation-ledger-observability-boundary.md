# ADR-0003: Invocation Ledger is the observability boundary

## Status

Accepted

## Context

Usage logging, usage query APIs, smoke scripts, failure replay, Router evaluation, and ContextManager evaluation all need the same execution facts.
Keeping separate record formats for each surface causes repeated fields and weak comparisons.

ADR-0001 still stands: LiteLLM remains the downstream gateway while the Go service learns and wraps selected infra behavior.
ADR-0002 still stands: this remains a modular monolith.

## Decision

Use Invocation Ledger as the source of truth for execution facts.

The ledger stores append-only JSONL records at `data/invocations.jsonl` by default.
The path is configurable with `NANOBOT_INVOCATION_LOG_PATH`.

`usage.Record` remains as a compatibility type, but it is a projection/subset inside `invocation.Record`.
Usage APIs must read Invocation Ledger records and project usage views instead of reading a separate usage ledger.

Replay, Debug, Router evaluation, and ContextManager evaluation must use Invocation Ledger records instead of inventing new record formats.

## Consequences

### Pros

- One source of truth for task execution facts.
- Router and ContextManager experiments can compare runs using the same schema.
- Legacy usage APIs can remain while the internal boundary becomes cleaner.
- The append-only JSONL model remains reviewable with shell tools.

### Cons

- Existing M5-M7 docs need to be reclassified as projections over the ledger.
- Local `data/usage.jsonl` lab state is not migrated.
- The ledger schema becomes a stronger contract and should change carefully.
