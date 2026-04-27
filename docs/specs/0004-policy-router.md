# Spec 0004: PolicyRouter

## Purpose

Choose a LiteLLM model alias based on task features, budget hints, and model profiles.

## Inputs

| Input | Source |
|---|---|
| task type | API/task layer |
| context size | ContextManager |
| stream requested | API request |
| budget hint | config or task metadata |
| model profiles | `configs/models.yaml` |
| policies | `configs/policies.yaml` |

## Outputs

```json
{
  "model_alias": "code-smart",
  "fallback_chain": ["code-cheap"],
  "route_reason": "high quality required for code review"
}
```

## Data Structures

```json
{
  "alias": "code-smart",
  "quality": 0.9,
  "cost": 0.7,
  "latency": 0.5,
  "context_limit": 128000,
  "supports_stream": true
}
```

## Error Cases

| Case | Expected |
|---|---|
| no matching model | deterministic config error |
| stream requested but unsupported | choose supported fallback or reject |
| context exceeds all limits | reject before LiteLLM call |

## Config

- `configs/models.yaml` defines model profiles.
- `configs/policies.yaml` defines scoring weights and fallback chains.

## Test Matrix

| Test | Expected |
|---|---|
| high quality review selects smart model | pass |
| low budget task selects cheap model | pass |
| stream requirement filters unsupported models | pass |
| route reason is populated | pass |

## Non-goals

- No ML router.
- No online bandit.
- No provider-specific adapter routing outside LiteLLM.
