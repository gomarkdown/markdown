#!/bin/bash
set -u -e -o pipefail

# Fuzz with the built-in Go fuzzer.
# FuzzParse covers the historical parser and CommonMark.
# FuzzCommonMark adds the CommonMark spec examples as seeds.
# Without a duration each target runs until a crash or until you stop it.
# Example: ./s/fuzz.sh 30s

if [ "${1:-}" != "" ]; then
	go test -fuzz=FuzzParse -fuzztime="$1"
	go test -fuzz=FuzzCommonMark -fuzztime="$1"
else
	go test -fuzz=FuzzParse
	go test -fuzz=FuzzCommonMark
fi
