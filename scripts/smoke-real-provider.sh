#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
RUN_ID="${RUN_ID:-run_$(date +%s)}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq

echo "health:"
curl -fsS "$BASE_URL/health" | jq .

payload='{
  "diff": "diff --git a/main.go b/main.go\n+func add(a int, b int) int { return a + b }",
  "repo_summary": "Go nanobot service smoke test.",
  "prior_plan": "Verify review task, task lookup, and usage lookup.",
  "logs": "smoke script submits one review task",
  "notes": "stale note should be dropped",
  "budget_hint": "low"
}'

echo "review:"
review_response="$(curl -sS --fail-with-body -X POST "$BASE_URL/tasks/review-diff" \
  -H 'content-type: application/json' \
  -H "X-Run-ID: $RUN_ID" \
  -H 'X-Scenario: smoke.real_provider' \
  -d "$payload")"
printf '%s\n' "$review_response" | jq '{run_id,attempt_id,task_id,status,model,route_reason,latency_ms,context_report}'

task_id="$(printf '%s\n' "$review_response" | jq -r '.task_id')"
if [[ -z "$task_id" || "$task_id" == "null" ]]; then
  echo "missing task_id in review response" >&2
  exit 1
fi

echo "task:"
curl -fsS "$BASE_URL/tasks/$task_id" | jq '{run_id,attempt_id,task_id,status,model,route_reason,latency_ms}'

echo "invocations by task:"
curl -fsS "$BASE_URL/invocations/tasks/$task_id" | jq '{task_id,count,records}'

echo "invocations by run:"
curl -fsS "$BASE_URL/invocations/runs/$RUN_ID" | jq '{run_id,count,records}'

echo "recent invocations:"
curl -fsS "$BASE_URL/invocations/recent?limit=3" | jq '{count,records}'

echo "legacy usage projection by task:"
curl -fsS "$BASE_URL/usage/tasks/$task_id" | jq '{task_id,count,records}'
