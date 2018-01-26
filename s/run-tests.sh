#!/bin/bash
set -u -e -o pipefail -o verbose

go build ./...

go test -v
