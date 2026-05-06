# Spec 0007: Failure Replay Lab

## Purpose

Provide repeatable local failure experiments for the Go -> LiteLLM path.

After ADR-0003, failure replay writes and queries Invocation Ledger records.
It should not introduce a separate failure result schema.

## Inputs

| Input | Source |
|---|---|
| base service URL | `BASE_URL` |
| LiteLLM URL | `LITELLM_BASE_URL` |
| LiteLLM API key | `LITELLM_API_KEY` |
| temporary invocation path | script-created temp dir |
| scenario and run ids | script-generated headers |

## Outputs

The script prints compact JSON summaries for:

- request validation failures
- timeout mapping
- downstream error mapping
- invocation records for failure cases
- success recovery after failures

## Data Structures

Failure records use Spec 0008 `invocation.Record`.
Legacy usage output, when needed, is a projection from the same record.

## Error Cases

| Case | Expected |
|---|---|
| empty diff | `400`, rejected invocation record |
| streaming requested | `400`, rejected invocation record |
| tiny LiteLLM timeout | `504`, failed invocation record with `error_kind=timeout` |
| missing LiteLLM model alias | `502`, failed invocation record with `error_kind=downstream` |
| normal request after failures | `200`, successful invocation record |

## Config

- No persistent config changes.
- Temporary model and policy YAML may be generated under `/tmp`.
- Temporary Go servers must be stopped on script exit.
- Temporary servers use `NANOBOT_INVOCATION_LOG_PATH`.

## Test Matrix

| Test | Expected |
|---|---|
| failure script runs against running Compose stack | pass |
| validation cases write rejected invocation records | pass |
| timeout case records failed invocation | pass |
| missing model case records failed invocation | pass |
| recovery case succeeds and records invocation | pass |

## Non-goals

- No provider fallback implementation.
- No database.
- No streaming support.
- No dashboard.
