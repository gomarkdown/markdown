#!/bin/bash
set -u -e -o pipefail -o verbose

go test

go run cmd/crashtest/main.go
