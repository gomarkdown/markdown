#!/bin/bash
set -u -e -o pipefail -o verbose

go test -bench=. -test.benchmem
