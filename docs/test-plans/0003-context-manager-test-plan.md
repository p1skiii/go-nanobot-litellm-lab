# Test Plan: ContextManager

## Unit Tests

| Test | Expected |
|---|---|
| current diff kept | `kept_blocks` contains `current_diff` |
| repo summary kept | `kept_blocks` contains `repo_summary` |
| old plan compressed | `compressed_blocks` contains `old_plan` |
| old logs compressed | `compressed_blocks` contains `old_logs` when useful lines exist |
| irrelevant notes dropped | `dropped_blocks` contains `irrelevant_notes` |
| no usable context rejected | returns `ErrNoUsableContext` |

## API Tests

| Test | Expected |
|---|---|
| `POST /tasks/review-diff` returns context report | response includes `context_report` |
| final context reaches reviewer | reviewer receives non-empty `FinalContext` |
| streaming still rejected | `stream=true` returns 400 |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| M3 context path with real provider | Go service + LiteLLM + Xiaomi MiMo | `200`, model output, `context_report` with keep/compress/drop |

## Non-goals

- Do not test model-generated compression in M3.
- Do not require persisted `context_report` from `GET /tasks/{id}` yet.
