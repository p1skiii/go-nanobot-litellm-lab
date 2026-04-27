# Spec 0002: LiteLLM Integration

## Purpose

Call LiteLLM Proxy as downstream gateway.

## Inputs

| Input | Source |
|---|---|
| task prompt | `internal/tasks` |
| model alias | `internal/router` or default config |
| base URL | config |
| API key | environment/config |
| request id | API middleware or task creation |

## Outputs

| Output | Consumer |
|---|---|
| assistant text | task result |
| model name | task metadata |
| usage fields | UsageLogger |
| latency | telemetry and task metadata |
| downstream error | API error mapper |

## Data Structures

M1 should implement a narrow chat completion client shape compatible with LiteLLM's OpenAI-style proxy endpoint.

```json
{
  "model": "code-cheap",
  "messages": [
    {"role": "system", "content": "string"},
    {"role": "user", "content": "string"}
  ],
  "stream": false
}
```

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

## Test Matrix

| Test | Expected |
|---|---|
| build request from review task | valid LiteLLM request |
| timeout mapped to 504 | pass |
| downstream 5xx mapped to 502 | pass |
| capture returned model | pass |

## Non-goals

- Do not implement provider adapters directly.
- Do not bypass LiteLLM in M1.
- Do not implement streaming before non-stream works.
