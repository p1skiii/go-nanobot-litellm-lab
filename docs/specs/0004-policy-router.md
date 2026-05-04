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

API response field:

```json
{
  "model": "code-smart",
  "route_reason": "task=review_diff context_chars=108 budget=high_quality selected=code-smart score=1.065"
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
| no matching model | return 400 before LiteLLM call |
| stream requested but unsupported | choose supported model or reject |
| context exceeds all limits | reject before LiteLLM call |

## Config

- `configs/models.yaml` defines model profiles.
- `configs/policies.yaml` defines scoring weights and fallback chains.
- `NANOBOT_MODELS_CONFIG` can override the models config path.
- `NANOBOT_POLICIES_CONFIG` can override the policies config path.

Current routing behavior:

- `budget_hint=high_quality` favors quality and selects `code-smart` with the default config.
- `budget_hint=low` favors lower cost and selects `code-cheap` with the default config.
- `context_chars` filters out models whose `context_limit` is too small.
- `stream=true` is still rejected by the API before routing because streaming has no public contract yet.
- The selected alias is passed to LiteLLM as the chat completion `model`.

## Test Matrix

| Test | Expected |
|---|---|
| high quality review selects smart model | pass |
| low budget task selects cheap model | pass |
| stream requirement filters unsupported models | pass |
| context too large rejects before LiteLLM | pass |
| config files load | pass |
| route reason is populated | pass |
| real provider smoke routes high quality to `code-smart` and low budget to `code-cheap` | pass |

## Non-goals

- No ML router.
- No online bandit.
- No provider-specific adapter routing outside LiteLLM.
