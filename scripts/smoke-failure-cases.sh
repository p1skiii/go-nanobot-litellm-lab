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
  local status

  if [[ -n "$data" ]]; then
    status="$(curl -sS -o "$body_file" -w '%{http_code}' -X "$method" "$url" \
      -H 'content-type: application/json' \
      -d "$data")"
  else
    status="$(curl -sS -o "$body_file" -w '%{http_code}' -X "$method" "$url")"
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
  local usage_path="$5"
  local log_path="$tmpdir/server-$port.log"

  NANOBOT_ADDR=":$port" \
    LITELLM_BASE_URL="$LITELLM_BASE_URL" \
    LITELLM_API_KEY="$LITELLM_API_KEY" \
    LITELLM_MODEL="code-cheap" \
    LITELLM_TIMEOUT="$timeout" \
    NANOBOT_MODELS_CONFIG="$models" \
    NANOBOT_POLICIES_CONFIG="$policies" \
    NANOBOT_USAGE_LOG_PATH="$usage_path" \
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
status="$(request_json POST "$BASE_URL/tasks/review-diff" '{"diff":"   "}' "$body")"
assert_status "$status" "400" "empty diff" "$body"
jq . "$body"

echo "case: streaming request -> 400"
body="$tmpdir/streaming.json"
status="$(request_json POST "$BASE_URL/tasks/review-diff" '{"diff":"diff","stream":true}' "$body")"
assert_status "$status" "400" "streaming request" "$body"
jq . "$body"

echo "case: tiny LiteLLM timeout -> 504 + failed usage"
timeout_usage="$tmpdir/timeout-usage.jsonl"
start_temp_server 18091 1ms "configs/models.yaml" "configs/policies.yaml" "$timeout_usage"
body="$tmpdir/timeout.json"
status="$(request_json POST "http://127.0.0.1:18091/tasks/review-diff" "$valid_payload" "$body")"
assert_status "$status" "504" "timeout case" "$body"
jq '{task_id,status,model,route_reason,latency_ms,error}' "$body"
timeout_task_id="$(jq -r '.task_id' "$body")"
timeout_usage_body="$tmpdir/timeout-usage-response.json"
status="$(request_json GET "http://127.0.0.1:18091/usage/tasks/$timeout_task_id" "" "$timeout_usage_body")"
assert_status "$status" "200" "timeout usage lookup" "$timeout_usage_body"
jq '{task_id,count,records}' "$timeout_usage_body"
jq -e '.count == 1 and .records[0].status == "failed" and (.records[0].error | length > 0)' "$timeout_usage_body" >/dev/null

echo "case: missing LiteLLM model -> 502 + failed usage"
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
missing_usage="$tmpdir/missing-usage.jsonl"
start_temp_server 18092 20s "$missing_models" "$missing_policies" "$missing_usage"
body="$tmpdir/missing-model.json"
status="$(request_json POST "http://127.0.0.1:18092/tasks/review-diff" "$valid_payload" "$body")"
assert_status "$status" "502" "missing model case" "$body"
jq '{task_id,status,model,route_reason,latency_ms,error}' "$body"
missing_task_id="$(jq -r '.task_id' "$body")"
missing_usage_body="$tmpdir/missing-usage-response.json"
status="$(request_json GET "http://127.0.0.1:18092/usage/tasks/$missing_task_id" "" "$missing_usage_body")"
assert_status "$status" "200" "missing model usage lookup" "$missing_usage_body"
jq '{task_id,count,records}' "$missing_usage_body"
jq -e '.count == 1 and .records[0].status == "failed" and .records[0].model_alias == "missing-model"' "$missing_usage_body" >/dev/null

echo "case: recovery success -> 200 + usage"
body="$tmpdir/recovery.json"
status="$(request_json POST "$BASE_URL/tasks/review-diff" "$valid_payload" "$body")"
assert_status "$status" "200" "recovery success" "$body"
jq '{task_id,status,model,route_reason,latency_ms}' "$body"
recovery_task_id="$(jq -r '.task_id' "$body")"
recovery_usage_body="$tmpdir/recovery-usage.json"
status="$(request_json GET "$BASE_URL/usage/tasks/$recovery_task_id" "" "$recovery_usage_body")"
assert_status "$status" "200" "recovery usage lookup" "$recovery_usage_body"
jq '{task_id,count,records}' "$recovery_usage_body"
jq -e '.count >= 1 and .records[-1].status == "success" and .records[-1].total_tokens > 0' "$recovery_usage_body" >/dev/null

echo "failure replay passed"
