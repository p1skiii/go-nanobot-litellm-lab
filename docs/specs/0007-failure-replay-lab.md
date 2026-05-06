# Spec 0007: Failure Replay Lab

## Purpose

Provide repeatable local failure experiments for the Go -> LiteLLM path.

## Inputs

| Input | Source |
|---|---|
| base service URL | `BASE_URL` |
| LiteLLM URL | `LITELLM_BASE_URL` |
| LiteLLM API key | `LITELLM_API_KEY` |
| temporary usage directory | script-created temp dir |

## Outputs

The script prints compact JSON summaries for:

- request validation failures
- timeout mapping
- downstream error mapping
- usage records for failure cases
- success recovery after failures

## Data Structures

Failure usage records reuse Spec 0005 `usage.Record`.

## Error Cases

| Case | Expected |
|---|---|
| empty diff | `400` |
| streaming requested | `400` |
| tiny LiteLLM timeout | `504`, failed usage record |
| missing LiteLLM model alias | `502`, failed usage record |
| normal request after failures | `200`, successful usage record |

## Config

- No persistent config changes.
- Temporary model and policy YAML may be generated under `/tmp`.
- Temporary Go servers must be stopped on script exit.

## Test Matrix

| Test | Expected |
|---|---|
| failure script runs against running Compose stack | pass |
| timeout case records failed usage | pass |
| missing model case records failed usage | pass |
| recovery case succeeds | pass |

## Non-goals

- No provider fallback implementation.
- No database.
- No streaming support.
- No dashboard.
