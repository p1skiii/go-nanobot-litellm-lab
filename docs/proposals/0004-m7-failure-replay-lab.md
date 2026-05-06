# Proposal 0004: M7 Failure Replay Lab

## Status

Accepted for M7 implementation

## Context

M6 proves the happy-path local query loop.
The project now needs repeatable failure checks so timeout, downstream error, validation failure, and recovery behavior are observable without relying on chat history.

## Proposed Changes

- Add a failure smoke script that exercises expected failure cases through HTTP.
- Use temporary local Go service instances where a failure needs different timeout or routing config.
- Record the observed implications in the LiteLLM behavior notes.
- Update roadmap, current state, and CI status with M7 verification results.

## Non-goals

- Do not implement fallback in Go.
- Do not change ContextManager, PolicyRouter, UsageLogger storage, K8s, OTel, or database behavior.
- Do not add a test framework dependency for shell smoke tests.

## Acceptance

- Empty diff returns `400`.
- Streaming request returns `400`.
- Tiny upstream timeout returns `504` and writes a failed usage record.
- Missing LiteLLM model returns `502` and writes a failed usage record.
- A normal real-provider request still succeeds after failure cases.
