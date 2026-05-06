# Spec 0002: LiteLLM Integration

## Purpose

Call LiteLLM Proxy as downstream gateway.

## Inputs

| Input | Source |
|---|---|
| task prompt | `internal/tasks` |
| final context | `internal/contextmgr` |
| model alias | `internal/router` |
| base URL | config |
| API key | environment/config |
| request id | API middleware or task creation |

## Outputs

| Output | Consumer |
|---|---|
| assistant text | task result |
| returned model | task metadata and Invocation Ledger |
| usage fields | Invocation Ledger usage projection |
| latency | telemetry and task metadata |
| downstream error | API error mapper |

## Data Structures

M1 should implement a narrow chat completion client shape compatible with LiteLLM's OpenAI-style proxy endpoint.
M4 can override the model alias per request.

```json
{
  "model": "code-smart",
  "messages": [
    {"role": "system", "content": "string"},
    {"role": "user", "content": "final_context"}
  ],
  "stream": false
}
```

Current behavior:

- `ReviewRequest.FinalContext` is used as the user message when present.
- `ReviewRequest.ModelAlias` overrides the default client model when present.
- Non-2xx LiteLLM responses are mapped to service `502`, except local timeout maps to `504`.
- Usage fields are extracted for non-stream responses and persisted through Invocation Ledger.

## Error Cases

| Case | Expected |
|---|---|
| timeout | map to 504 |
| downstream 5xx | map to 502 |
| downstream 4xx | map to 502 until more specific policy exists |
| malformed downstream response | map to 502 |

## Config

| Name | Purpose |
|---|---|
| `LITELLM_BASE_URL` | LiteLLM proxy base URL |
| `LITELLM_API_KEY` | proxy auth key |
| `LITELLM_MODEL` | fallback default alias when router does not provide one |
| `LITELLM_TIMEOUT` | downstream timeout |

## Test Matrix

| Test | Expected |
|---|---|
| build request from review task | valid LiteLLM request |
| timeout mapped to 504 | pass |
| downstream 5xx mapped to 502 | pass |
| capture returned model | pass |
| use request-level model alias | pass |
| use final context when provided | pass |

## Non-goals

- Do not implement provider adapters directly.
- Do not bypass LiteLLM in M1.
- Do not implement streaming before non-stream works.
- Do not make the LiteLLM client own storage.
