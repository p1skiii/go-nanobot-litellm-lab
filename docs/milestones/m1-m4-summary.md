# M1-M4 Phase Summary

## Current Chain

```text
HTTP client
  -> Go Nanobot Backend
  -> ContextManager
  -> PolicyRouter
  -> LiteLLM Proxy
  -> real provider or mock provider
```

## M1: Go Calls LiteLLM

Goal:

- Prove the Go service can receive a review task, call LiteLLM non-stream, and return a task result.

Implemented:

- `POST /tasks/review-diff`
- `GET /tasks/{id}`
- `task_id` and `request_id`
- in-memory task store
- LiteLLM non-stream chat completion client
- timeout mapped to `504`
- downstream LiteLLM errors mapped to `502`

Verification:

- Unit tests for API, task store, config, and LiteLLM client.
- Mock LiteLLM end-to-end path.

## M2: LiteLLM Behavior Study

Goal:

- Learn real LiteLLM gateway behavior before replacing selected parts.

Implemented as research:

- Reclassified the first mock-only notes as synthetic baseline.
- Ran real Xiaomi MiMo provider-backed LiteLLM experiments.
- Documented model aliasing, fallback, usage, streaming, rate/budget limits, and error mapping.

Key findings:

- LiteLLM can route a real OpenAI-compatible provider via `openai/<model>` and custom `api_base`.
- Fallback behavior is visible through LiteLLM headers such as `x-litellm-model-group` and `x-litellm-attempted-fallbacks`.
- Real non-stream responses include token usage.
- Streaming uses SSE and needs a separate API shape later.
- Local no-DB LiteLLM rate/budget behavior is not enough to rely on hard enforcement.

## M3: Custom ContextManager

Goal:

- Govern task context before LiteLLM by deciding what to keep, compress, or drop.

Implemented:

- `internal/contextmgr` deterministic rules.
- Request inputs: `diff`, `repo_summary`, `prior_plan`, `logs`, `notes`.
- `final_context` is passed to LiteLLM instead of raw concatenation.
- `context_report` is returned on `POST /tasks/review-diff`.

Current rules:

- `current_diff`: keep
- `repo_summary`: keep or cap
- `old_plan`: compress
- `old_logs`: compress or drop
- `irrelevant_notes`: drop

Verification:

- Unit tests for context decisions.
- API tests prove `FinalContext` reaches the reviewer.
- Real provider M3 smoke passed through Go -> LiteLLM -> Xiaomi MiMo.

## M4: Custom PolicyRouter

Goal:

- Choose LiteLLM model aliases from local profiles and policy weights before calling LiteLLM.

Implemented:

- `internal/router` score-based router.
- Reads `configs/models.yaml`.
- Reads `configs/policies.yaml`.
- Request input: optional `budget_hint`.
- Response output: `route_reason`.
- LiteLLM client uses request-level `ModelAlias`.

Current behavior:

- `budget_hint=high_quality` selects `code-smart`.
- `budget_hint=low` selects `code-cheap`.
- Context size and stream support filter model candidates.

Verification:

- Unit tests for scoring, stream filtering, context-limit rejection, and config loading.
- Real provider M4 smoke passed for both `code-smart` and `code-cheap` aliases.

## Current Boundary

Keep:

- LiteLLM remains the downstream gateway.
- Go service owns ContextManager and PolicyRouter.
- Modular monolith remains the architecture.

Do not do yet:

- Do not build a full custom gateway.
- Do not add microservices.
- Do not add K8s.
- Do not add OTel.
- Do not add streaming until a separate API contract exists.

## Next Phase: M5

M5 should add task-level UsageLogger and a reliable local Compose loop.

M5 should not introduce a billing product, dashboard, database, or custom gateway.
