# Spec 0003: ContextManager

## Purpose

Convert raw task input into context blocks and decide keep/compress/drop.

## Inputs

| Input | Example |
|---|---|
| current diff | git diff text |
| repo summary | short project summary |
| prior plan | old implementation plan |
| logs | previous command output |
| notes | stale discussion or unrelated context |

## Outputs

```json
{
  "final_context": "string",
  "kept_blocks": [],
  "compressed_blocks": [],
  "dropped_blocks": []
}
```

API response field:

```json
{
  "context_report": {
    "kept_blocks": ["current_diff", "repo_summary"],
    "compressed_blocks": ["old_plan", "old_logs"],
    "dropped_blocks": ["irrelevant_notes"]
  }
}
```

## Data Structures

| Type | Example | Default action |
|---|---|---|
| `current_diff` | git diff | keep |
| `repo_summary` | project summary | keep |
| `old_plan` | previous plan | compress |
| `old_logs` | long logs | compress/drop |
| `irrelevant_notes` | stale discussion | drop |

Current deterministic rules:

- `current_diff` is kept.
- `repo_summary` is kept, capped if it exceeds the summary budget.
- `old_plan` is compressed into a one-line capped summary.
- `old_logs` keeps important error lines first, otherwise keeps recent tail lines.
- `irrelevant_notes` is dropped.

## Error Cases

| Case | Expected |
|---|---|
| no usable context | return validation error |
| compressed context still too large | drop lower-priority optional blocks |

## Config

M3 uses code defaults for budgets. Context default actions remain documented in `configs/policies.yaml`.

## Test Matrix

| Test | Expected |
|---|---|
| current diff kept | pass |
| repo summary kept | pass |
| old plan compressed | pass |
| old logs compressed | pass |
| irrelevant notes dropped | pass |
| no usable context rejected | pass |
| API passes final context to LiteLLM client | pass |

## Non-goals

- No semantic vector search in M3.
- No model-generated compression until deterministic rules exist.
- No persistence of full context reports in task store yet.
