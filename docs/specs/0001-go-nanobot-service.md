# Spec 0001: Go Nanobot Service

## Purpose

Expose a small HTTP service that accepts a code review task and calls LiteLLM.

## Endpoints

| Method | Path | Purpose | Milestone |
|---|---|---|---|
| GET | `/health` | health check | M0 |
| POST | `/tasks/review-diff` | submit diff review task | M1 |
| GET | `/tasks/{id}` | get task result | M1 |

## Request Shape

```json
{
  "diff": "string",
  "repo_summary": "string optional",
  "prior_plan": "string optional",
  "logs": "string optional",
  "notes": "string optional",
  "budget_hint": "low|high_quality optional",
  "stream": false
}
```

## Response Shape

```json
{
  "task_id": "string",
  "request_id": "string",
  "status": "success|failed",
  "result": "string",
  "model": "string",
  "route_reason": "string",
  "latency_ms": 1234,
  "context_report": {
    "kept_blocks": ["current_diff"],
    "compressed_blocks": ["old_plan"],
    "dropped_blocks": ["irrelevant_notes"]
  }
}
```

Notes:

- `context_report` is returned by `POST /tasks/review-diff`.
- `GET /tasks/{id}` returns persisted task fields. It currently includes `route_reason` but not the full `context_report`.
- `model` is the selected LiteLLM alias after M4, not necessarily the provider model returned in the LiteLLM body.

## M0 Health Response

```json
{
  "status": "ok",
  "service": "go-nanobot-litellm-lab"
}
```

## Error Cases

| Case | Expected |
|---|---|
| unsupported method on `/health` | 405 |
| empty diff | 400 |
| LiteLLM timeout | 504 |
| LiteLLM error | 502 |
| no usable context | 400 |
| no routable model | 400 |

## Config

| Name | Default | Purpose |
|---|---|---|
| `NANOBOT_ADDR` | `:8080` | HTTP listen address |
| `NANOBOT_MODELS_CONFIG` | `configs/models.yaml` | model profile config |
| `NANOBOT_POLICIES_CONFIG` | `configs/policies.yaml` | routing and context policy config |

## Test Matrix

| Test | Milestone |
|---|---|
| `/health` returns 200 and JSON body | M0 |
| unsupported method on `/health` returns 405 | M0 |
| empty diff rejected | M1 |
| task can be fetched by id | M1 |
| context report returned | M3 |
| route reason returned | M4 |

## Non-goals

- Do not call LiteLLM in M0.
- Do not add persistence in M0.
- Do not add streaming before a separate streaming API contract exists.
