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

## Current Gaps

- No `/tasks/review-diff` implementation in M0.
- No persistence in M0.
- No LiteLLM integration in M0.
