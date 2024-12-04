$ErrorActionPreference = 'Stop'
go clean -testcache
#go test -race -v .
#go test -race -v ./ast
#go test -race -v ./html
#go test -race -v ./parser

go test -v .
go test -v ./ast
go test -v ./html
go test -v ./parser
