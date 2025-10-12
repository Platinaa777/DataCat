#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
command -v kubectl >/dev/null || { echo 'kubectl is required.' >&2; exit 1; }
if [[ "${1:-}" != '--apply' ]]; then
  kubectl kustomize .ops/k8s
  echo 'Review the manifest, configure images/secrets/hosts, then run with --apply.' >&2
  exit 0
fi
kubectl apply -k .ops/k8s
kubectl -n datacat rollout status deployment/datacat-server --timeout=120s
kubectl -n datacat rollout status deployment/datacat-ui --timeout=120s
