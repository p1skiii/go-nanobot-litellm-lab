# AGENTS.md

## Working Agreements

- Build maintainable files and simple folder structures.
- Prefer markdown and yaml for state and protocols.
- Do not introduce heavy dependencies unless necessary.
- When creating automation, keep it reviewable and easy to disable.
- For personal assistant systems, separate dynamic state from stable protocols.
- Never overwrite formal protocol files without creating a proposal first.

## Role Model

This repo uses PM, Architect, Engineer, and QA as operating modes, even when one human or agent performs the work.

| Role | Owns | Answers |
|---|---|---|
| PM | `docs/vision.md`, `docs/roadmap.md`, `docs/current-state.md` | Current milestone, weekly focus, completion criteria, non-goals |
| Architect | `docs/gap-analysis.md`, `docs/adr/`, spec review | Architecture boundaries, accepted decisions, decision reversals |
| Engineer | `cmd/`, `internal/`, `configs/`, `deploy/` | Code implementation, tests, CI fixes |
| QA | `docs/test-plans/`, `docs/ci-status.md` | Coverage, verification status, failures, gaps |

## Formal Protocol Files

Formal protocol files are stable collaboration contracts:

- `CLAUDE.md`
- `AGENTS.md`
- `docs/adr/*.md`
- `docs/specs/*.md`
- `docs/test-plans/*.md`
- `docs/vision.md`
- `docs/roadmap.md`
- `docs/gap-analysis.md`

After M0, changing these files requires creating a proposal first. Dynamic state files such as `docs/current-state.md`, `docs/ci-status.md`, and `docs/research/*.md` may be updated directly as work progresses.

## Execution Rule

ADR decides architecture. Specs fill implementation detail. Test plans validate specs. Code should follow accepted ADRs and active specs instead of inventing new architecture during implementation.
