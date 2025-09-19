#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
command -v protoc >/dev/null || { echo 'Install protoc 25.x first.' >&2; exit 1; }
tools_dir=$(mktemp -d)
trap 'rm -rf "$tools_dir"' EXIT
GOBIN="$tools_dir" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
GOBIN="$tools_dir" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
PATH="$tools_dir:$PATH" protoc -I api/proto \
  --go_out=. --go_opt=module=github.com/Platinaa777/DataCat \
  --go-grpc_out=. --go-grpc_opt=module=github.com/Platinaa777/DataCat \
  api/proto/metrics/v1/metrics.proto
