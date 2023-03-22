#!/bin/bash
set -u -e -o pipefail -o verbose

go clean -testcache
go test -race -v .
go test -race -v ./ast
go test -race -v ./html
go test -race -v ./parser
