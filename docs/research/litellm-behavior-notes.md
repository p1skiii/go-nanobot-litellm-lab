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
- M1 does not prove real provider behavior, fallback, streaming, usage, cost, budget, or rate limits.
- M2 must use provider-backed experiments for final conclusions. Local mock observations are useful only as a synthetic baseline.

## Source Documents

- Xiaomi MiMo first API call: https://platform.xiaomimimo.com/static/docs/quick-start/first-api-call.md
- Xiaomi MiMo OpenAI-compatible API: https://platform.xiaomimimo.com/static/docs/api/chat/openai-api.md
- Xiaomi MiMo pricing and rate limits: https://platform.xiaomimimo.com/static/docs/pricing.md
- Xiaomi MiMo error codes: https://platform.xiaomimimo.com/static/docs/quick-start/error-codes.md
- LiteLLM OpenAI-compatible endpoint behavior: https://docs.litellm.ai/docs/providers/openai_compatible

Key documentation facts:

- Xiaomi MiMo exposes an OpenAI-compatible chat endpoint at `https://api.xiaomimimo.com/v1/chat/completions`.
- Xiaomi MiMo accepts either `api-key: $MIMO_API_KEY` or `Authorization: Bearer $MIMO_API_KEY`.
- `mimo-v2-flash` supports text generation and streaming output.
- Xiaomi MiMo documents `usage` fields, including `completion_tokens`, `prompt_tokens`, `total_tokens`, and token detail fields.
- Xiaomi MiMo documents `429` for too-frequent requests or exhausted Token Plan quota.
- LiteLLM can call OpenAI-compatible endpoints with the `openai/` model prefix and a custom `api_base`.

## M2 Synthetic/Local Mock Baseline

Date: 2026-04-28 18:04:54 CST

Status:

- Reclassified on 2026-05-01 as synthetic/local mock observations.
- These results do not satisfy M2 by themselves because most calls used LiteLLM `mock_response` or intentionally broken local endpoints.
- The provider-backed experiments below are the M2 source of truth.

Setup:

- LiteLLM image: `ghcr.io/berriai/litellm:main-latest`
- Observed readiness version: `1.82.6`
- Temporary config path: `/tmp/litellm-m2/config.yaml`
- Proxy command:

```bash
docker run --rm --name litellm-m2-study \
  -p 4100:4000 \
  -v /tmp/litellm-m2/config.yaml:/app/config.yaml:ro \
  ghcr.io/berriai/litellm:main-latest \
  --port 4000 \
  --config /app/config.yaml
```

Temporary model groups:

| Model group | Purpose |
|---|---|
| `code-cheap` | non-stream mock response and usage probe |
| `stream-mock` | streaming mock response probe |
| `fallback-primary` | intentionally broken primary, `api_base=http://127.0.0.1:9` |
| `fallback-ok` | mock fallback response |
| `rate-limited` | mock response with `rpm: 1` |
| `error-bad-downstream` | intentionally broken downstream with no fallback |

Synthetic observations:

- Configured LiteLLM fallback from broken primary to mock target returned `200 OK` and `x-litellm-attempted-fallbacks: 1`.
- Mock non-stream responses returned a synthetic `usage` body and LiteLLM cost/duration headers.
- Mock streaming returned OpenAI-style SSE chunks ending in `data: [DONE]`.
- Local no-DB `rpm: 1` emitted rate-limit headers but did not enforce `429`.
- Virtual key budget creation returned `500` because LiteLLM DB was not connected.
- Broken downstream returned LiteLLM `500`; missing model returned LiteLLM `400`.

## M2 Real Provider-backed Environment

Date: 2026-05-01 00:45:28 CST

Direct provider sanity check:

- Provider: Xiaomi MiMo API Open Platform
- Provider base URL: `https://api.xiaomimimo.com/v1`
- Provider model: `mimo-v2-flash`
- Authentication used in tests: `Authorization: Bearer $MIMO_API_KEY`
- The API key was passed as an environment variable and was not written to repo files.
- Direct provider call returned `200 OK`, content `provider-ok`, and real `usage`:
  - `completion_tokens: 3`
  - `prompt_tokens: 35`
  - `total_tokens: 38`
  - `completion_tokens_details.reasoning_tokens: 0`
  - `prompt_tokens_details.cached_tokens: 13`

LiteLLM setup:

- LiteLLM image: `ghcr.io/berriai/litellm:main-latest`
- Observed readiness version: `1.82.6`
- Temporary config path: `/tmp/litellm-m2-real/config.yaml`
- LiteLLM listen URL: `http://127.0.0.1:4101`
- LiteLLM DB status: not connected
- Proxy command:

```bash
docker run --rm --name litellm-m2-real \
  -p 4101:4000 \
  -e MIMO_API_KEY="$MIMO_API_KEY" \
  -v /tmp/litellm-m2-real/config.yaml:/app/config.yaml:ro \
  ghcr.io/berriai/litellm:main-latest \
  --port 4000 \
  --config /app/config.yaml
```

Temporary provider-backed model groups:

| Model group | Purpose |
|---|---|
| `mimo-real-flash` | real Xiaomi MiMo non-stream and streaming calls |
| `mimo-fallback-primary` | intentionally broken primary, `api_base=http://127.0.0.1:9/v1` |
| `mimo-fallback-ok` | real Xiaomi MiMo fallback target |
| `mimo-rate-limited` | real Xiaomi MiMo call with LiteLLM `rpm: 1` |
| `mimo-bad-auth` | real Xiaomi MiMo endpoint with intentionally invalid upstream API key |

Relevant config shape:

```yaml
model_list:
  - model_name: mimo-real-flash
    litellm_params:
      model: openai/mimo-v2-flash
      api_base: https://api.xiaomimimo.com/v1
      api_key: os.environ/MIMO_API_KEY

litellm_settings:
  request_timeout: 20
  num_retries: 0
```

## Model Alias

M1 confirmed that `model_name: code-cheap` in `configs/litellm.yaml` can be called by the Go service as the chat completion `model`.

M2 confirmed that `model_name: mimo-real-flash` can map to a real OpenAI-compatible provider model using `model: openai/mimo-v2-flash` and `api_base: https://api.xiaomimimo.com/v1`.

## Fallback

### Provider-backed experiment: broken primary falls back to real MiMo model

Setup:

- `mimo-fallback-primary` points at unreachable `api_base=http://127.0.0.1:9/v1`.
- `mimo-fallback-ok` points at real Xiaomi MiMo `mimo-v2-flash`.
- `litellm_settings.fallbacks` maps `mimo-fallback-primary` to `mimo-fallback-ok`.

Command:

```bash
curl -sS -i -X POST http://127.0.0.1:4101/v1/chat/completions \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"model":"mimo-fallback-primary","messages":[{"role":"system","content":"You are MiMo, an AI assistant developed by Xiaomi. Answer concisely."},{"role":"user","content":"Return exactly: fallback-real-ok"}],"max_completion_tokens":32,"temperature":0,"stream":false}'
```

Observed response:

- HTTP status: `200 OK`
- Body content: `fallback-real-ok`
- Body `model`: `mimo-v2-flash`
- Body `usage`: `completion_tokens=4`, `prompt_tokens=36`, `total_tokens=40`
- Header `x-litellm-model-group`: `mimo-fallback-ok`
- Header `x-litellm-attempted-fallbacks`: `1`
- Header `x-litellm-model-api-base`: `https://api.xiaomimimo.com/v1`
- Header `x-litellm-response-duration-ms`: `1434.979`

Implication for our Go service:

- Real fallback works when LiteLLM owns the fallback chain.
- The JSON body reports the provider model, while `x-litellm-model-group` reports the selected LiteLLM alias.
- M1 currently stores `model` from the JSON body only, so it loses the route/fallback alias. A future metadata capture change should persist `x-litellm-model-group` and `x-litellm-attempted-fallbacks`.
- Do not implement custom fallback or PolicyRouter in M2.

## Usage

### Provider-backed experiment: non-stream real usage fields

Setup:

- `mimo-real-flash` points at real Xiaomi MiMo `mimo-v2-flash`.
- Request used a short deterministic prompt and `stream=false`.

Command:

```bash
curl -sS -i -X POST http://127.0.0.1:4101/v1/chat/completions \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"model":"mimo-real-flash","messages":[{"role":"system","content":"You are MiMo, an AI assistant developed by Xiaomi. Answer concisely."},{"role":"user","content":"Return exactly: litellm-real-ok"}],"max_completion_tokens":32,"temperature":0,"stream":false}'
```

Observed response:

- HTTP status: `200 OK`
- Body content: `litellm-real-ok`
- Body `model`: `mimo-real-flash`
- Body `usage`:
  - `completion_tokens: 6`
  - `prompt_tokens: 38`
  - `total_tokens: 44`
  - `completion_tokens_details.reasoning_tokens: 0`
  - `prompt_tokens_details.cached_tokens: 26`
- LiteLLM headers:
  - `x-litellm-model-group: mimo-real-flash`
  - `x-litellm-attempted-retries: 0`
  - `x-litellm-attempted-fallbacks: 0`
  - `x-litellm-response-duration-ms: 1019.31`
  - `x-litellm-overhead-duration-ms: 29.244`
  - `x-litellm-key-spend: 0.0`
  - `x-litellm-response-cost-original: 0.0`

Implication for our Go service:

- Real provider responses include useful token details beyond the minimal OpenAI token trio.
- LiteLLM does not calculate non-zero cost for this custom OpenAI-compatible MiMo model by default, even though usage tokens are present.
- UsageLogger should not assume LiteLLM cost headers are reliable for custom OpenAI-compatible providers unless model pricing is configured.
- The next code improvement before M5 could be a small metadata capture field set, but not a full UsageLogger.

## Streaming

### Provider-backed experiment: real MiMo streaming through LiteLLM

Setup:

- `mimo-real-flash` points at real Xiaomi MiMo `mimo-v2-flash`.
- Request used `stream=true`.

Command:

```bash
curl -sS -N -i -X POST http://127.0.0.1:4101/v1/chat/completions \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"model":"mimo-real-flash","messages":[{"role":"system","content":"You are MiMo, an AI assistant developed by Xiaomi. Answer concisely."},{"role":"user","content":"Return exactly: stream-real-ok"}],"max_completion_tokens":32,"temperature":0,"stream":true}'
```

Observed response:

- HTTP status: `200 OK`
- `content-type`: `text/event-stream; charset=utf-8`
- LiteLLM forwarded provider headers with `llm_provider-*` prefixes, including:
  - `llm_provider-content-type: text/event-stream`
  - `llm_provider-server: MiFE/3.4.34`
  - `llm_provider-x-mife-upstream-status: 200`
- Body contained three `data: {...}` chunk lines and one `data: [DONE]`.
- Chunks:
  - first chunk included `delta.role=assistant` and `delta.content=stream`
  - second chunk included `delta.content=-real-ok`
  - third chunk had `finish_reason=stop` and an empty delta
- No `usage` field appeared in observed stream chunks.

Implication for our Go service:

- Real provider streaming is OpenAI-compatible SSE after LiteLLM.
- The current M1 JSON task response should not be stretched to handle streaming. Streaming needs a separate response contract or endpoint.
- If streaming is added later, usage capture cannot rely on final JSON body usage. We need either stream usage support, provider-specific final chunks if available, or a follow-up accounting path.
- M2 should record this behavior only; do not implement streaming yet.

## Rate Limit / Budget

### Provider-backed experiment: LiteLLM `rpm: 1` on real MiMo calls

Setup:

- `mimo-rate-limited` points at real Xiaomi MiMo `mimo-v2-flash`.
- LiteLLM model config sets `rpm: 1`.
- LiteLLM ran without a database.

Command:

```bash
for i in 1 2 3; do
  curl -sS -i -X POST http://127.0.0.1:4101/v1/chat/completions \
    -H 'Authorization: Bearer sk-local-dev' \
    -H 'Content-Type: application/json' \
    --data "{\"model\":\"mimo-rate-limited\",\"messages\":[{\"role\":\"system\",\"content\":\"You are MiMo. Answer with only the requested marker.\"},{\"role\":\"user\",\"content\":\"Return exactly: rate-real-$i\"}],\"max_completion_tokens\":16,\"temperature\":0,\"stream\":false}"
done
```

Observed response:

- All three requests returned `200 OK`.
- Request 1 headers: `x-ratelimit-limit-requests: 1`, `x-ratelimit-remaining-requests: 1`
- Request 2 headers: `x-ratelimit-limit-requests: 1`, `x-ratelimit-remaining-requests: 0`
- Request 3 headers: `x-ratelimit-limit-requests: 1`, `x-ratelimit-remaining-requests: -1`
- No `429` was enforced by this local no-DB LiteLLM setup.
- All three requests reached the real provider and returned real `usage` fields.

Budget command:

```bash
curl -sS -i -X POST http://127.0.0.1:4101/key/generate \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"models":["mimo-real-flash"],"max_budget":0.000001,"duration":"1h","key_alias":"m2-real-low-budget"}'
```

Budget observed response:

- HTTP status: `500 Internal Server Error`
- Body error: `DB not connected. See https://docs.litellm.ai/docs/proxy/virtual_keys`

Implication for our Go service:

- In local no-DB mode, LiteLLM `rpm` headers are visible but not sufficient evidence of hard enforcement.
- LiteLLM virtual key budgets require a database-backed proxy setup.
- Xiaomi MiMo's documented provider limit is `RPM: 100`, `TPM: 10M` for `mimo-v2-flash`; M2 did not intentionally exhaust that provider quota.
- The Go service should preserve rate-limit headers and handle `429`, but we should not design around local no-DB `rpm` as a reliable limiter.

## Error Mapping

M1 Go unit tests cover timeout to 504 and downstream/malformed LiteLLM response to 502.

### Provider-backed experiment: real upstream auth failure through LiteLLM

Setup:

- `mimo-bad-auth` points at real Xiaomi MiMo `mimo-v2-flash`.
- The upstream provider API key is intentionally invalid.
- No fallback is configured for `mimo-bad-auth`.

Command:

```bash
curl -sS -i -X POST http://127.0.0.1:4101/v1/chat/completions \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"model":"mimo-bad-auth","messages":[{"role":"user","content":"auth error probe"}],"max_completion_tokens":16,"temperature":0,"stream":false}'
```

Observed response:

- HTTP status: `401 Unauthorized`
- Body `error.code`: `401`
- Body message included `AuthenticationError`, `OpenAIException - Invalid API Key`, and `No fallback model group found`.
- Headers included:
  - `x-litellm-call-id`
  - `x-litellm-response-cost: 0`
  - `x-litellm-key-spend: 0.0`
  - `x-litellm-timeout: 20`

Implication for our Go service:

- Real provider authentication failures preserve a `401` status through LiteLLM.
- The current Go service maps any non-2xx LiteLLM response to upstream client `502`, which is acceptable for M1 because LiteLLM/provider details are internal.
- Future API policy can choose whether to expose some LiteLLM statuses directly, but that is not needed for M2.

### Provider-backed experiment: invalid LiteLLM model group

Command:

```bash
curl -sS -i -X POST http://127.0.0.1:4101/v1/chat/completions \
  -H 'Authorization: Bearer sk-local-dev' \
  -H 'Content-Type: application/json' \
  --data '{"model":"mimo-missing-model","messages":[{"role":"user","content":"missing model probe"}],"max_completion_tokens":16,"temperature":0,"stream":false}'
```

Observed response:

- HTTP status: `400 Bad Request`
- Body error: `Invalid model name passed in model=mimo-missing-model. Call /v1/models to view available models for your key.`

Implication for our Go service:

- Bad model alias is a LiteLLM configuration/request problem, not a provider generation problem.
- If the Go service chooses models from config instead of accepting arbitrary caller models, this should normally remain an internal `502` path.

### Provider-backed experiment: Go service with real LiteLLM chain

Setup:

- Go service ran with:
  - `LITELLM_BASE_URL=http://127.0.0.1:4101`
  - `LITELLM_MODEL=mimo-real-flash`
  - `LITELLM_TIMEOUT=30s`
- LiteLLM forwarded to real Xiaomi MiMo `mimo-v2-flash`.

Command:

```bash
NANOBOT_ADDR=:18082 \
LITELLM_BASE_URL=http://127.0.0.1:4101 \
LITELLM_API_KEY=sk-local-dev \
LITELLM_MODEL=mimo-real-flash \
LITELLM_TIMEOUT=30s \
go run ./cmd/server

curl -sS -i -X POST http://127.0.0.1:18082/tasks/review-diff \
  -H 'Content-Type: application/json' \
  --data '{"diff":"diff --git a/main.go b/main.go\n+fmt.Println(\"hello\")","repo_summary":"tiny Go CLI","stream":false}'
```

Observed response:

- HTTP status: `200 OK`
- Task status: `success`
- `model`: `mimo-real-flash`
- `latency_ms`: `1680`
- Result was a real code-review style response from MiMo.

Implication for our Go service:

- The M1 chain is now verified against a real provider, not only LiteLLM mock mode.
- M1 code still does not capture LiteLLM usage headers or response body usage. That is a known future improvement before UsageLogger.

## What Should Our Go Service Own?

- API request validation
- task id and request id
- in-memory task state for M1
- timeout/error mapping at the service boundary
- latency capture for the LiteLLM call
- extraction of LiteLLM headers for model group, fallback count, cost, duration, and rate-limit signals
- response body usage extraction for non-stream calls
- an explicit future streaming API contract, if streaming becomes part of the product

## What Should Remain Delegated To LiteLLM?

- provider adapters
- model alias resolution
- fallback behavior through at least M2
- usage/cost calculation until UsageLogger is introduced
- budget and virtual key enforcement until we run LiteLLM with a database

## M7 Failure Replay Observations

Date: 2026-05-06 10:52:26 CST

Setup:

- Docker Compose stack was already running:
  - Go service at `http://127.0.0.1:8080`
  - LiteLLM at `http://127.0.0.1:4000`
  - LiteLLM configured with Xiaomi MiMo-backed `code-cheap` and `code-smart`
- Failure replay script: `scripts/smoke-failure-cases.sh`
- Temporary local Go services were started for cases that required different timeout or routing config.
- Temporary usage JSONL files were used for failure-specific assertions.

Observed responses:

| Case | Observed | Usage result |
|---|---|---|
| empty diff | Go returned `400` with `diff is required` | no task created |
| streaming requested | Go returned `400` with `streaming is not supported in M1` | no task created |
| tiny LiteLLM timeout | Go returned `504` after `1ms` timeout | failed usage record written |
| missing LiteLLM model | Go returned `502` after LiteLLM returned `400` invalid model | failed usage record written |
| recovery request | Go returned `200` through LiteLLM -> Xiaomi MiMo | success usage record with tokens written |

Specific M7 values from the smoke run:

- Timeout case:
  - task id: `task_56acb23d1ecc38e2`
  - model alias: `code-cheap`
  - latency: `1ms`
  - error: `context deadline exceeded`
- Missing model case:
  - task id: `task_312b987153c47bc0`
  - model alias: `missing-model`
  - LiteLLM body included `Invalid model name passed in model=missing-model`
- Recovery case:
  - task id: `task_837c70a0078a2aa6`
  - model alias: `code-cheap`
  - latency: `3003ms`
  - usage: `prompt_tokens=86`, `completion_tokens=188`, `total_tokens=274`

Implication for our Go service:

- Validation failures should stay before task creation and usage logging.
- Timeout and downstream LiteLLM errors should create failed task records and failed usage records.
- LiteLLM `400` for invalid model alias remains an internal upstream problem for the Go service and maps to client `502`.
- A recovery request after failure cases proves failed attempts do not poison the in-memory task store, usage logger, or LiteLLM client.
- M7 does not require Go-owned fallback; LiteLLM behavior remains the reference until a later milestone explicitly changes that boundary.

## M8 Invocation Ledger Observations

Date: 2026-05-06 11:33:54 CST

Setup:

- Existing Docker Compose LiteLLM container stayed running with real Xiaomi MiMo-backed `code-cheap` and `code-smart`.
- Nanobot was rebuilt from source with the multi-stage Dockerfile.
- Source of truth moved to `data/invocations.jsonl`.
- Smoke script: `scripts/smoke-real-provider.sh`
- Failure replay script: `scripts/smoke-failure-cases.sh`

Observed response:

| Case | Observed |
|---|---|
| real provider smoke | `200`, `run_id`, `attempt_id`, `task_id`, `route_reason`, `context_report` |
| invocation by task | one record with `task_status=success`, `error_kind=none`, nested usage tokens |
| invocation by run | same record queryable by shared `run_id` |
| legacy usage projection | `/usage/tasks/{id}` returned projected `usage.Record` |
| empty diff | `400`, rejected invocation record, no task id |
| streaming requested | `400`, rejected invocation record, no task id |
| tiny timeout | `504`, failed invocation record with `error_kind=timeout` |
| missing model | `502`, failed invocation record with `error_kind=downstream` |
| recovery request | `200`, successful invocation record with token usage |

Specific M8 values from smoke runs:

- Real provider smoke:
  - task id: `task_4fbeadd6f901ac90`
  - run id: `run_1778038406`
  - attempt id: `attempt_2fa382f585c49cbc`
  - model alias: `code-cheap`
  - latency: `3474ms`
  - usage: `prompt_tokens=98`, `completion_tokens=222`, `total_tokens=320`
- Failure replay:
  - timeout task id: `task_cff719f54b8f2791`, `error_kind=timeout`
  - missing model task id: `task_30c5d678825e3dbc`, `error_kind=downstream`
  - recovery task id: `task_3b358219789e248d`, `total_tokens=295`

Implication for our Go service:

- Router, ContextManager, Usage, Replay, Debug, and future Evaluation should read Invocation Ledger records.
- Usage is now a compatibility projection, not its own source of truth.
- Validation failures can be compared with task attempts because they share `run_id`, `attempt_id`, scenario, HTTP status, and error kind.
- M9 should compare `code-cheap` and `code-smart` through shared-run ledger records before changing Router weights.
- M10 should compare ContextManager input variants through shared-run ledger records before changing context rules.
