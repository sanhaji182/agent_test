#!/usr/bin/env bash
# GoTest Agent — End-to-End Smoke Test (ADR-006 post-verification + product E2E gate)
#
# Menjalankan SATU siklus produk lengkap terhadap web app nyata:
#   boot server (in-memory store) → test-connection provider → create run
#   → poll sampai selesai → tarik events/report → verifikasi artefak.
#
# Prasyarat:
#   export ANTHROPIC_API_KEY=sk-ant-...   (atau LLM_PROVIDER+LLM_API_KEY)
#
# Pakai: ./scripts/smoke-e2e.sh [target-url]
#   target-url default: https://example.com (halaman statis — cocok untuk smoke)
set -euo pipefail

TARGET_URL="${1:-https://example.com}"
PORT="${APP_PORT:-8080}"
BASE="http://localhost:${PORT}/api/v1"
LOG="$(mktemp -t gotest-smoke)"
PASS=0; FAIL=0

say()  { printf '\033[1;34m[smoke]\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m  ✓ %s\033[0m\n' "$*"; PASS=$((PASS+1)); }
bad()  { printf '\033[1;31m  ✗ %s\033[0m\n' "$*"; FAIL=$((FAIL+1)); }

if [ -z "${ANTHROPIC_API_KEY:-}${LLM_API_KEY:-}" ]; then
  echo "ERROR: set ANTHROPIC_API_KEY (or LLM_PROVIDER + LLM_API_KEY) first." >&2
  exit 1
fi

say "1/7 build server"
go build -o /tmp/gotest-server ./cmd/server

say "2/7 boot server (in-memory store, dev mode) — log: $LOG"
APP_ENV=development DATABASE_URL="" /tmp/gotest-server >"$LOG" 2>&1 &
SERVER_PID=$!
trap 'kill $SERVER_PID 2>/dev/null || true' EXIT

for i in $(seq 1 30); do
  curl -sf "http://localhost:${PORT}/health" >/dev/null 2>&1 && break
  sleep 0.5
  [ "$i" = 30 ] && { bad "server never became healthy"; cat "$LOG"; exit 1; }
done
ok "server healthy"

say "3/7 provider test-connection (ADR-006 manual check)"
PROVIDER="${LLM_PROVIDER:-anthropic}"
KEY="${LLM_API_KEY:-$ANTHROPIC_API_KEY}"
MODEL="${LLM_MODEL:-claude-sonnet-4-5}"
TC=$(curl -sf -X POST "$BASE/ai/test-provider" -H 'Content-Type: application/json' \
  -d "{\"provider\":\"$PROVIDER\",\"model\":\"$MODEL\",\"api_key\":\"$KEY\"}")
echo "$TC" | grep -q '"success":true' && ok "test-provider success ($PROVIDER/$MODEL)" \
  || bad "test-provider failed: $TC"

say "4/7 create run against $TARGET_URL"
RUN=$(curl -sf -X POST "$BASE/runs" -H 'Content-Type: application/json' -d "{
  \"project_path\": \"$TARGET_URL\",
  \"requirements\": \"Open the page, verify it loads, check the main heading is visible, and confirm at least one link is present.\",
  \"mode\": \"simple\"
}")
RUN_ID=$(echo "$RUN" | sed -n 's/.*"run_id":"\([^"]*\)".*/\1/p')
[ -n "$RUN_ID" ] && ok "run created: $RUN_ID" || { bad "create run failed: $RUN"; exit 1; }

say "5/7 poll run state (timeout 10m — includes one-time Playwright browser download)"
STATE="unknown"
for i in $(seq 1 300); do
  STATE=$(curl -sf "$BASE/runs/$RUN_ID" | sed -n 's/.*"state":"\([^"]*\)".*/\1/p' | head -1)
  case "$STATE" in
    done|failed|simulated) break ;;
  esac
  sleep 2
done
echo "  final state: $STATE"
case "$STATE" in
  done)      ok "run completed: done" ;;
  simulated) ok "run completed: simulated (no real execution path hit — investigate)" ;;
  failed)    bad "run FAILED — inspect events below" ;;
  *)         bad "run stuck in state '$STATE' after timeout" ;;
esac

say "6/7 verify artifacts"
EVENTS=$(curl -sf "$BASE/runs/$RUN_ID/events")
EVCOUNT=$(echo "$EVENTS" | grep -o '"type"' | wc -l | tr -d ' ')
[ "$EVCOUNT" -gt 2 ] && ok "events emitted: $EVCOUNT" || bad "too few events ($EVCOUNT)"
echo "$EVENTS" | grep -q 'test_plan\|plan' && ok "plan-phase events present" || bad "no plan events"

FULL=$(curl -sf "$BASE/runs/$RUN_ID")
echo "$FULL" | grep -q '"test_files":\[' && ok "test files generated" || bad "no test_files on run"
echo "$FULL" | grep -q '"run_result":{' && ok "run_result present" || bad "no run_result"

REPORT=$(curl -sf "$BASE/runs/$RUN_ID/report" || true)
echo "$REPORT" | grep -q '<html' && ok "HTML report renders" || bad "report endpoint failed"

say "7/7 summary"
echo ""
echo "  PASS: $PASS   FAIL: $FAIL"
echo "  Run:    $BASE/runs/$RUN_ID"
echo "  Log:    $LOG"
echo ""
if [ "$FAIL" -gt 0 ]; then
  echo "--- last 40 lines of server log ---"
  tail -40 "$LOG"
  echo "--- last events ---"
  echo "$EVENTS" | tail -c 2000
  exit 1
fi
say "ALL GREEN — end-to-end product loop verified"
