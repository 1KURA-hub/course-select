#!/usr/bin/env bash
set -euo pipefail

DEPLOY_PATH="${1:-/go-course/course-select}"
DEPLOY_BRANCH="${2:-main}"

cd "$DEPLOY_PATH"

echo "[deploy] path: $DEPLOY_PATH"
echo "[deploy] branch: $DEPLOY_BRANCH"

git fetch origin "$DEPLOY_BRANCH"
git checkout "$DEPLOY_BRANCH"
git pull --ff-only origin "$DEPLOY_BRANCH"

docker compose up -d --build app

docker compose ps

curl -fsS http://127.0.0.1:8080/healthz >/dev/null
echo "[deploy] health check passed"

