set -xe

export GO111MODULE="off"

## see our config, for debugging
wd=`pwd`
echo "pwd: ${pwd}"
go env

## Install go-fuzz
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build

## build and send to fuzzit
go build ./...
go-fuzz-build -libfuzzer -o fuzzer.a .
clang -fsanitize=fuzzer fuzzer.a -o fuzzer

wget -q -O fuzzit https://github.com/fuzzitdev/fuzzit/releases/download/v2.0.0/fuzzit_Linux_x86_64
chmod a+x fuzzit
./fuzzit auth ${FUZZIT_API_KEY}
export TARGET_ID=2n6hO2dQzylLxX5GGhRG

ls -lah

./fuzzit create job --type $1 --branch $TRAVIS_BRANCH --revision $TRAVIS_COMMIT --target_id $TARGET_ID ./fuzzer
