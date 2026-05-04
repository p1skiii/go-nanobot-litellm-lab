# Test Plan: PolicyRouter

## Unit Tests

| Test | Expected |
|---|---|
| high quality review selects smart model | `model_alias=code-smart` |
| low budget task selects cheap model | `model_alias=code-cheap` |
| stream requirement filters unsupported models | unsupported model is skipped |
| context exceeds all model limits | returns no matching model error |
| model and policy config files load | router can route from temp YAML files |
| route reason is populated | response has non-empty `route_reason` |

## API Tests

| Test | Expected |
|---|---|
| `budget_hint=high_quality` reaches reviewer as smart alias | reviewer receives `ModelAlias=code-smart` |
| router failure returns before LiteLLM | API returns 400 |
| successful response includes route reason | response includes `route_reason` |

## Real Provider Smoke

| Test | Setup | Expected |
|---|---|---|
| high quality route | Go service + LiteLLM aliases + Xiaomi MiMo | `200`, `model=code-smart`, `route_reason` selects `code-smart` |
| low budget route | Go service + LiteLLM aliases + Xiaomi MiMo | `200`, `model=code-cheap`, `route_reason` selects `code-cheap` |

## Non-goals

- No ML router.
- No online learning.
- No provider adapter bypass outside LiteLLM.
