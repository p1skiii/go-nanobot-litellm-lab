# Test Plan: Failure Replay Lab

## Script Checks

| Check | Expected |
|---|---|
| empty diff | HTTP `400`, rejected invocation record |
| streaming request | HTTP `400`, rejected invocation record |
| timeout case | HTTP `504`, failed invocation record with `error_kind=timeout` |
| missing model case | HTTP `502`, failed invocation record with `error_kind=downstream` |
| recovery success | HTTP `200`, successful invocation record with usage tokens |

## Unit Tests

M7 itself does not require new Go behavior.
M8 covers the ledger write/read behavior that M7 depends on.

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| failure replay script | Docker Compose LiteLLM + Xiaomi MiMo | all failure cases pass, then recovery request succeeds |

## Non-goals

- No load testing.
- No fallback testing beyond existing M2 LiteLLM study.
- No new Go routing behavior.
