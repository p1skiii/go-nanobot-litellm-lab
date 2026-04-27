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
  "stream": false
}
```

## Response Shape

```json
{
  "task_id": "string",
  "status": "success|failed",
  "result": "string",
  "model": "string",
  "latency_ms": 1234
}
```

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

## Config

| Name | Default | Purpose |
|---|---|---|
| `NANOBOT_ADDR` | `:8080` | HTTP listen address |

## Test Matrix

| Test | Milestone |
|---|---|
| `/health` returns 200 and JSON body | M0 |
| unsupported method on `/health` returns 405 | M0 |
| empty diff rejected | M1 |
| task can be fetched by id | M1 |

## Non-goals

- Do not call LiteLLM in M0.
- Do not add persistence in M0.
- Do not add streaming in M0.
