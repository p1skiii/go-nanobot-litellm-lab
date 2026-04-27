# LiteLLM Behavior Notes

## M1 Observation: Non-stream Mock Path

Date: 2026-04-28 02:26:41 CST

Setup:

- LiteLLM Proxy ran from `deploy/docker-compose.yml`.
- `configs/litellm.yaml` exposed model aliases `code-cheap` and `code-smart`.
- `code-cheap` used LiteLLM `mock_response` so M1 did not require a provider API key.
- Go service called `POST /v1/chat/completions` through LiteLLM using `model=code-cheap` and `stream=false`.

Result:

- LiteLLM Proxy initialized with `code-cheap` and `code-smart`.
- Go `POST /tasks/review-diff` returned `status=success`, a generated `task_id`, propagated `request_id`, `model=code-cheap`, and result `M1 LiteLLM mock review response`.
- `GET /tasks/{id}` returned the stored in-memory task result.

Current interpretation:

- M1 proves the HTTP client -> Go service -> LiteLLM Proxy -> mock model -> Go response chain.
- This does not yet prove real provider behavior, fallback, streaming, usage, cost, budget, or rate limits.
- M2 should replace this single happy-path observation with focused LiteLLM behavior experiments.

## Model Alias

M1 confirmed that `model_name: code-cheap` in `configs/litellm.yaml` can be called by the Go service as the chat completion `model`.

## Fallback

Pending M2 experiment.

## Usage

Pending M2 experiment.

## Streaming

Pending M2 experiment.

## Rate Limit / Budget

Pending M2 experiment.

## Error Mapping

M1 Go unit tests cover timeout to 504 and downstream/malformed LiteLLM response to 502. Real LiteLLM error shapes still need M2 verification.

## What Should Our Go Service Own?

- API request validation
- task id and request id
- in-memory task state for M1
- timeout/error mapping at the service boundary
- latency capture for the LiteLLM call

## What Should Remain Delegated To LiteLLM?

- provider adapters
- model alias resolution
- real fallback behavior
- real usage/cost/rate-limit behavior until observed in M2
