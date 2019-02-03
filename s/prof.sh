#!/bin/bash
set -u -e -o pipefail -o verbose

go test -cpuprofile cpu.prof -bench=BenchmarkReferenceMarkdownSyntax
