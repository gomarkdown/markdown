#!/bin/bash
set -u -e -o pipefail -o verbose

go clean -testcache
go test -race -v ./...
