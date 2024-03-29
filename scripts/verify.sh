#!/usr/bin/env sh

INSPECT_COVERAGE="${INSPECT_COVERAGE:=false}"

set -e
set -x

go vet ./...
go fmt ./...

if [ "$INSPECT_COVERAGE" == "true" ]; then
  go test ./... -coverprofile=coverage.out "$@"
  go tool cover -html=coverage.out
else
  go test  ./... -cover "$@"
fi

go run ./cmd/build generate pkg/build/example/workflow.yaml --config pkg/build/example/pipeline.yaml