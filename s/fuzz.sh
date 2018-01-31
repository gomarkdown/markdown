#!/bin/bash
set -u -e -o pipefail

# This runs a go-fuzz as described in:
# https://medium.com/@dgryski/go-fuzz-github-com-arolek-ase-3c74d5a3150c
# https://github.com/dvyukov/go-fuzz

go get github.com/dvyukov/go-fuzz/go-fuzz
go get github.com/dvyukov/go-fuzz/go-fuzz-build

# this step is expensive, so re-use previous runs if possible
if [ ! -f ./markdown-fuzz.zip ]; then
    mkdir -p fuzz-workdir/corpus
    cp testdata/*.text fuzz-workdir/corpus
    echo "running go-fuzz-build, might take a while..."
    go-fuzz-build github.com/gomarkdown/markdown
fi

echo "running go-fuzz"
go-fuzz -bin=./markdown-fuzz.zip -workdir=fuzz-workdir
