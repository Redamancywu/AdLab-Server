#!/usr/bin/env sh
set -eu

BASE_URL="${1:-http://127.0.0.1:8090}"
APP_ID="${2:-app_ios_001}"
PLATFORM="${3:-ios}"

echo ">>> smoke: health"
curl -fsS "${BASE_URL}/health" >/dev/null

echo ">>> smoke: ready"
curl -fsS "${BASE_URL}/ready" >/dev/null

echo ">>> smoke: sdk init"
curl -fsS -X POST "${BASE_URL}/api/v1/sdk/init" \
  -H "Content-Type: application/json" \
  -d "{\"app_id\":\"${APP_ID}\",\"platform\":\"${PLATFORM}\"}" >/dev/null

echo ">>> sdkapi smoke passed"
