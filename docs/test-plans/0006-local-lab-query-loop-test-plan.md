# Test Plan: Local Lab Query Loop

## Unit Tests

| Test | Expected |
|---|---|
| JSONL reader returns recent records | newest N records returned in chronological order |
| JSONL reader filters by task id | only matching records returned |
| missing JSONL file | empty records, no error |
| malformed JSONL file | error returned |
| invalid recent limit | HTTP 400 |

## Integration Tests

| Test | Setup | Expected |
|---|---|---|
| `GET /usage/recent` | temp JSONL file | records returned |
| `GET /usage/tasks/{id}` | temp JSONL file | matching task records returned |
| review then usage lookup | fake reviewer | task id appears in usage lookup |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| review then recent usage | Go + LiteLLM + Xiaomi MiMo | `200`, task id, route reason, usage tokens |
| compose review then usage lookup | Docker Compose + Xiaomi MiMo | task id appears in `/usage/tasks/{id}` |

## Latest Verification

| Check | Status | Notes |
|---|---|---|
| `go test ./...` | pass | all packages passed |
| Compose rebuild with M6 binary | pass | nanobot image rebuilt from local linux/arm64 static binary |
| `scripts/smoke-real-provider.sh` | pass | health, review, task lookup, usage by task, and recent usage passed through Docker Compose and Xiaomi MiMo |

## Non-goals

- No dashboard.
- No database.
- No auth.
