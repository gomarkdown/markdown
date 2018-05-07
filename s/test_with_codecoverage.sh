#!/usr/bin/env bash

# based on https://github.com/ory/go-acc/blob/master/README.md

set -e

cover_pkgs=$(go list ./... | grep -v /vendor | tr "\n" ",")
echo $cover_pkgs

echo 'mode: atomic' > coverage.txt
go list ./... | grep -v /vendor | grep -v /cmd | xargs -n1 -I{} sh -c 'go test -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg ${cover_pkgs} {} && tail -n +2 coverage.tmp >> coverage.txt || exit 255' && rm coverage.tmp
