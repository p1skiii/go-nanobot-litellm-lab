# Test Plan: Failure Replay Lab

## Script Checks

| Check | Expected |
|---|---|
| empty diff | HTTP `400` |
| streaming request | HTTP `400` |
| timeout case | HTTP `504`, failed usage record |
| missing model case | HTTP `502`, failed usage record |
| recovery success | HTTP `200`, usage record with tokens |

## Unit Tests

M7 does not require new Go unit tests unless code changes are introduced.

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| failure replay script | Docker Compose LiteLLM + Xiaomi MiMo | all failure cases pass, then recovery request succeeds |

## Latest Verification

| Check | Status | Notes |
|---|---|---|
| `go test ./...` | pass | all packages passed |
| `scripts/smoke-failure-cases.sh` | pass | empty diff `400`, streaming `400`, timeout `504`, missing model `502`, recovery `200` |
| timeout usage record | pass | failed usage record contained task id, route reason, latency, and error |
| missing model usage record | pass | failed usage record contained `model_alias=missing-model` and LiteLLM `400` body |
| recovery usage record | pass | success record contained returned model and token usage |

## Non-goals

- No load testing.
- No fallback testing beyond existing M2 LiteLLM study.
- No new Go routing behavior.
