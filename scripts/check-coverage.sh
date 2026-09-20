#!/bin/sh
set -eu

threshold="${COVERAGE_THRESHOLD:-80}"
coverage_file="${COVERAGE_FILE:-coverage.out}"

go test ./... -coverprofile="$coverage_file"

coverage="$(go tool cover -func="$coverage_file" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')"

awk -v coverage="$coverage" -v threshold="$threshold" 'BEGIN {
  if (coverage + 0 < threshold + 0) {
    printf "coverage %.1f%% is below required %.1f%%\n", coverage, threshold
    exit 1
  }
  printf "coverage %.1f%% meets required %.1f%%\n", coverage, threshold
}'
