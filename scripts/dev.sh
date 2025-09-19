#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
./scripts/init-env.sh
set -a
source .env
set +a
docker compose up -d --wait postgres
docker compose up -d keycloak
ready=0
for ((attempt=0; attempt<90; attempt++)); do
  if curl --silent --fail "$OIDC_ISSUER/.well-known/openid-configuration" >/dev/null; then ready=1; break; fi
  sleep 1
done
if [[ "$ready" != 1 ]]; then echo 'Keycloak did not become ready.' >&2; exit 1; fi
npm --prefix frontend/datacat-ui ci
pids=()
cleanup() { for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done; }
trap cleanup EXIT INT TERM
go run ./cmd/datacat -migrate &
pids+=("$!")
npm --prefix frontend/datacat-ui start &
pids+=("$!")
wait "${pids[@]}"
