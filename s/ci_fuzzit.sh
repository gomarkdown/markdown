set -xe

export GO111MODULE="off"

## install go-fuzz
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build

## build and send to fuzzit
go-fuzz-build -libfuzzer -o fuzzer.a .
clang -fsanitize=fuzzer fuzzer.a -o fuzzer

wget -q -O fuzzit https://github.com/fuzzitdev/fuzzit/releases/download/v2.4.27/fuzzit_Linux_x86_64
chmod a+x fuzzit

# TODO: create a target and re-enable
export TARGET=markdown
# ./fuzzit create job --type fuzzing  $TARGET ./fuzzer
