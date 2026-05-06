#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
LITELLM_BASE_URL="${LITELLM_BASE_URL:-http://127.0.0.1:4000}"
LITELLM_API_KEY="${LITELLM_API_KEY:-sk-local-dev}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq
require_cmd go

tmpdir="$(mktemp -d /tmp/nanobot-m7-failure.XXXXXX)"
pids=()

cleanup() {
  for pid in "${pids[@]}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
  rm -rf "$tmpdir"
}
trap cleanup EXIT

request_json() {
  local method="$1"
  local url="$2"
  local data="${3:-}"
  local body_file="$4"
  local run_id="${5:-}"
  local scenario="${6:-}"
  local status
  local args=(-sS -o "$body_file" -w '%{http_code}' -X "$method" "$url")

  if [[ -n "$run_id" ]]; then
    args+=(-H "X-Run-ID: $run_id")
  fi
  if [[ -n "$scenario" ]]; then
    args+=(-H "X-Scenario: $scenario")
  fi

  if [[ -n "$data" ]]; then
    status="$(curl "${args[@]}" -H 'content-type: application/json' -d "$data")"
  else
    status="$(curl "${args[@]}")"
  fi

  printf '%s' "$status"
}

assert_status() {
  local got="$1"
  local want="$2"
  local label="$3"
  local body_file="$4"

  if [[ "$got" != "$want" ]]; then
    echo "$label: status=$got want=$want" >&2
    cat "$body_file" >&2
    exit 1
  fi
}

wait_health() {
  local url="$1"
  for _ in $(seq 1 40); do
    if curl -fsS "$url/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done
  echo "server did not become healthy: $url" >&2
  return 1
}

start_temp_server() {
  local port="$1"
  local timeout="$2"
  local models="$3"
  local policies="$4"
  local invocation_path="$5"
  local log_path="$tmpdir/server-$port.log"

  NANOBOT_ADDR=":$port" \
    LITELLM_BASE_URL="$LITELLM_BASE_URL" \
    LITELLM_API_KEY="$LITELLM_API_KEY" \
    LITELLM_MODEL="code-cheap" \
    LITELLM_TIMEOUT="$timeout" \
    NANOBOT_MODELS_CONFIG="$models" \
    NANOBOT_POLICIES_CONFIG="$policies" \
    NANOBOT_INVOCATION_LOG_PATH="$invocation_path" \
    go run ./cmd/server >"$log_path" 2>&1 &

  local pid="$!"
  pids+=("$pid")
  wait_health "http://127.0.0.1:$port"
}

valid_payload='{
  "diff": "diff --git a/main.go b/main.go\n+func add(a int, b int) int { return a + b }",
  "repo_summary": "Failure replay smoke test.",
  "prior_plan": "Check failure mapping and recovery.",
  "logs": "failure script",
  "notes": "stale note",
  "budget_hint": "low"
}'

echo "case: empty diff -> 400"
body="$tmpdir/empty-diff.json"
empty_run="run_m8_empty_$(date +%s)"
status="$(request_json POST "$BASE_URL/tasks/review-diff" '{"diff":"   "}' "$body" "$empty_run" "smoke.failure.empty_diff")"
assert_status "$status" "400" "empty diff" "$body"
jq . "$body"
empty_invocations="$tmpdir/empty-invocations.json"
status="$(request_json GET "$BASE_URL/invocations/runs/$empty_run" "" "$empty_invocations")"
assert_status "$status" "200" "empty diff invocation lookup" "$empty_invocations"
jq '{run_id,count,records}' "$empty_invocations"
jq -e '.count >= 1 and .records[-1].task_status == "rejected" and .records[-1].error_kind == "validation"' "$empty_invocations" >/dev/null

echo "case: streaming request -> 400"
body="$tmpdir/streaming.json"
stream_run="run_m8_stream_$(date +%s)"
status="$(request_json POST "$BASE_URL/tasks/review-diff" '{"diff":"diff","stream":true}' "$body" "$stream_run" "smoke.failure.streaming")"
assert_status "$status" "400" "streaming request" "$body"
jq . "$body"
stream_invocations="$tmpdir/stream-invocations.json"
status="$(request_json GET "$BASE_URL/invocations/runs/$stream_run" "" "$stream_invocations")"
assert_status "$status" "200" "streaming invocation lookup" "$stream_invocations"
jq '{run_id,count,records}' "$stream_invocations"
jq -e '.count >= 1 and .records[-1].task_status == "rejected" and .records[-1].error_kind == "validation"' "$stream_invocations" >/dev/null

echo "case: tiny LiteLLM timeout -> 504 + failed invocation"
timeout_invocation="$tmpdir/timeout-invocations.jsonl"
start_temp_server 18091 1ms "configs/models.yaml" "configs/policies.yaml" "$timeout_invocation"
body="$tmpdir/timeout.json"
timeout_run="run_m8_timeout_$(date +%s)"
status="$(request_json POST "http://127.0.0.1:18091/tasks/review-diff" "$valid_payload" "$body" "$timeout_run" "smoke.failure.timeout")"
assert_status "$status" "504" "timeout case" "$body"
jq '{run_id,attempt_id,task_id,status,model,route_reason,latency_ms,error}' "$body"
timeout_task_id="$(jq -r '.task_id' "$body")"
timeout_invocation_body="$tmpdir/timeout-invocation-response.json"
status="$(request_json GET "http://127.0.0.1:18091/invocations/tasks/$timeout_task_id" "" "$timeout_invocation_body")"
assert_status "$status" "200" "timeout invocation lookup" "$timeout_invocation_body"
jq '{task_id,count,records}' "$timeout_invocation_body"
jq -e '.count == 1 and .records[0].task_status == "failed" and .records[0].error_kind == "timeout" and (.records[0].usage.error | length > 0)' "$timeout_invocation_body" >/dev/null

echo "case: missing LiteLLM model -> 502 + failed invocation"
missing_models="$tmpdir/missing-models.yaml"
missing_policies="$tmpdir/missing-policies.yaml"
cat >"$missing_models" <<'YAML'
models:
  - alias: missing-model
    quality: 0.5
    cost: 0.5
    latency: 0.5
    context_limit: 128000
    supports_stream: true
YAML
cat >"$missing_policies" <<'YAML'
defaults:
  task_type: review_diff
  model_alias: missing-model
  fallback_chain: []
router:
  scoring:
    quality_weight: 0.5
    cost_weight: 0.3
    latency_weight: 0.2
YAML
missing_invocation="$tmpdir/missing-invocations.jsonl"
start_temp_server 18092 20s "$missing_models" "$missing_policies" "$missing_invocation"
body="$tmpdir/missing-model.json"
missing_run="run_m8_missing_$(date +%s)"
status="$(request_json POST "http://127.0.0.1:18092/tasks/review-diff" "$valid_payload" "$body" "$missing_run" "smoke.failure.missing_model")"
assert_status "$status" "502" "missing model case" "$body"
jq '{run_id,attempt_id,task_id,status,model,route_reason,latency_ms,error}' "$body"
missing_task_id="$(jq -r '.task_id' "$body")"
missing_invocation_body="$tmpdir/missing-invocation-response.json"
status="$(request_json GET "http://127.0.0.1:18092/invocations/tasks/$missing_task_id" "" "$missing_invocation_body")"
assert_status "$status" "200" "missing model invocation lookup" "$missing_invocation_body"
jq '{task_id,count,records}' "$missing_invocation_body"
jq -e '.count == 1 and .records[0].task_status == "failed" and .records[0].error_kind == "downstream" and .records[0].model_alias == "missing-model"' "$missing_invocation_body" >/dev/null

echo "case: recovery success -> 200 + invocation"
body="$tmpdir/recovery.json"
recovery_run="run_m8_recovery_$(date +%s)"
status="$(request_json POST "$BASE_URL/tasks/review-diff" "$valid_payload" "$body" "$recovery_run" "smoke.failure.recovery")"
assert_status "$status" "200" "recovery success" "$body"
jq '{run_id,attempt_id,task_id,status,model,route_reason,latency_ms}' "$body"
recovery_task_id="$(jq -r '.task_id' "$body")"
recovery_invocation_body="$tmpdir/recovery-invocation.json"
status="$(request_json GET "$BASE_URL/invocations/tasks/$recovery_task_id" "" "$recovery_invocation_body")"
assert_status "$status" "200" "recovery invocation lookup" "$recovery_invocation_body"
jq '{task_id,count,records}' "$recovery_invocation_body"
jq -e '.count >= 1 and .records[-1].task_status == "success" and .records[-1].usage.total_tokens > 0' "$recovery_invocation_body" >/dev/null

echo "failure replay passed"
