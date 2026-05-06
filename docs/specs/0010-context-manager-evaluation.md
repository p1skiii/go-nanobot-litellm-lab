# Spec 0010: ContextManager Evaluation

## Purpose

Evaluate existing ContextManager behavior using real invocation records.
M10 does not rewrite ContextManager rules unless the experiment supports a follow-up change.

## Inputs

| Scenario | Input |
|---|---|
| `evaluation.context.diff_only` | diff only |
| `evaluation.context.repo_summary` | diff plus repo summary |
| `evaluation.context.full_context` | diff plus repo summary, prior plan, logs, and notes |

All attempts should share one `run_id` and keep model strategy stable.

## Outputs

Compare Invocation Ledger records for:

- `context_report`
- `context_chars`
- prompt/completion/total tokens
- latency
- status
- result length

## Test Matrix

| Test | Expected |
|---|---|
| diff-only attempt | records kept `current_diff` |
| repo-summary attempt | records additional kept repo summary |
| full-context attempt | records keep/compress/drop decisions |
| shared run id | all attempts queryable through `/invocations/runs/{run_id}` |

## Non-goals

- No ContextManager rewrite.
- No semantic compression model.
- No storage of full prompts beyond current lab records.
