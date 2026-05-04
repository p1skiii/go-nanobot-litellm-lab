# Proposal 0001: Sync M1-M4 Docs and Prepare M5

## Status

Accepted for documentation sync

## Context

M1 through M4 have been implemented and verified, but some formal docs still describe earlier milestone shapes.
The harness rule says formal protocol files should not be changed after M0 without a proposal first.

## Proposed Changes

- Update `README.md` to show the current M4 state and M5 next step.
- Update `CLAUDE.md` to require real provider end-to-end smoke tests after each completed target.
- Update `docs/roadmap.md` with M1-M4 completion status and M5 entry criteria.
- Update specs 0001-0005 so request/response fields match current behavior.
- Add M3, M4, and M5 test plans.
- Add a concise M1-M4 milestone summary under `docs/milestones/`.
- Update dynamic state files with the latest commits and M4 real provider smoke result.

## Non-goals

- Do not change ADR-0001 or ADR-0002.
- Do not implement M5 in this proposal.
- Do not add new architecture beyond the existing modular monolith.

## Acceptance

- A future agent can read the docs and correctly understand current behavior through M4.
- A future M5 implementation starts from a clear UsageLogger scope and test plan.
- The real provider smoke requirement is visible in the harness.
