# Roadmap

## M0 Harness Ready

Exit:

- repo skeleton exists
- `CLAUDE.md` exists
- vision/gap/roadmap exists
- ADR-0001 accepted
- ADR-0002 accepted
- empty Go server can start

## M1 Go Calls LiteLLM

Exit:

- Go service exposes `/tasks/review-diff`
- Go service calls LiteLLM non-stream
- `task_id`, `request_id`, latency, and model are logged
- minimal unit tests exist

## M2 LiteLLM Behavior Study

Exit:

- fallback tested
- usage tested
- streaming tested
- rate limit/budget tested
- notes written in `docs/research/litellm-behavior-notes.md`

## M3 Custom ContextManager

Exit:

- context blocks exist
- keep/compress/drop rules exist
- review-diff task uses ContextManager before LiteLLM

## M4 Custom PolicyRouter

Exit:

- model profiles exist
- score-based router exists
- `route_reason` logged

## M5 UsageLogger + Docker Compose

Exit:

- task-level usage stored
- docker-compose starts Go service + LiteLLM
