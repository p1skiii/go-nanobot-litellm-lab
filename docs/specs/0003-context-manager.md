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

## Data Structures

| Type | Example | Default action |
|---|---|---|
| `current_diff` | git diff | keep |
| `repo_summary` | project summary | keep |
| `old_plan` | previous plan | compress |
| `old_logs` | long logs | compress/drop |
| `irrelevant_notes` | stale discussion | drop |

## Error Cases

| Case | Expected |
|---|---|
| no usable context | return validation error |
| compressed context still too large | drop lower-priority optional blocks |

## Config

M3 may add context budgets and block rules under `configs/policies.yaml`.

## Test Matrix

| Test | Expected |
|---|---|
| current diff kept | pass |
| repo summary kept | pass |
| old plan compressed | pass |
| irrelevant notes dropped | pass |

## Non-goals

- No semantic vector search in M3.
- No model-generated compression until deterministic rules exist.
