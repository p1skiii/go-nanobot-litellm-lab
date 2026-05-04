# Test Plan: Go Nanobot Service

## Unit Tests

| Test | Expected |
|---|---|
| `/health` returns status 200 | pass |
| `/health` returns JSON health body | `status=ok`, `service=go-nanobot-litellm-lab` |
| unsupported method on `/health` | 405 |
| config defaults to `:8080` | pass |
| `NANOBOT_ADDR` overrides default | pass |

## Integration Tests

| Test | Setup | Expected |
|---|---|---|
| server starts | `go run ./cmd/server` | process listens |
| health endpoint reachable | server running | `GET /health` returns health JSON |

## M1 Additions

| Test | Expected |
|---|---|
| empty review diff rejected | 400 |
| review task receives task id | pass |
| task result can be fetched | pass |

## M3-M4 Additions

| Test | Expected |
|---|---|
| context report returned | pass |
| route reason returned | pass |
| selected model alias returned | pass |

## Current Gaps

- No durable persistence yet.
- No streaming endpoint yet.
- No usage log yet.
