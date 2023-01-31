#!/usr/bin/env sh

set -e
set -x

go run github.com/vektra/mockery/v2@latest \
  --name="CommandRunner" \
  --dir="./pkg/build" \
  --output="pkg/build/mocks"