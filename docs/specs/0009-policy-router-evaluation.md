# Spec 0009: PolicyRouter Evaluation

## Purpose

Evaluate existing PolicyRouter decisions using real invocation records.
M9 does not build an advanced Router.

## Inputs

| Input | Source |
|---|---|
| shared diff | smoke/evaluation script |
| run id | shared across comparison attempts |
| scenario | `evaluation.router.code-cheap` and `evaluation.router.code-smart` |
| budget hint | `low` and `high_quality` |
| invocation records | Invocation Ledger |

## Outputs

Write comparison notes to research docs with:

- selected model alias
- returned model
- route reason
- status
- latency
- prompt/completion/total tokens
- result length

## Test Matrix

| Test | Expected |
|---|---|
| low budget run | routes to `code-cheap` |
| high quality run | routes to `code-smart` |
| shared run id | both attempts queryable through `/invocations/runs/{run_id}` |
| comparison output | uses ledger fields, not custom script-only fields |

## Non-goals

- No new Router algorithm.
- No Router weight changes without evaluation evidence.
- No ML router or online bandit.
