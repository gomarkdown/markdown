# Things to do

[x] `go test` is very slow because `doTestsReference()` tests processing of every substring of original markdown. Remove that and add a separate test program for that to be run only in CI

[x] instead of using NodeType, use interface{} for Node data, like xml parser

[x] simplify oliPrefix()

[x] fuzzing!

[ ] docs: add examples like https://godoc.org/github.com/dgrijalva/jwt-go (put in foo_example_test.go). Or see https://github.com/garyburd/redigo/blob/master/redis/zpop_example_test.go#L5 / https://godoc.org/github.com/garyburd/redigo/redis or https://godoc.org/github.com/go-redis/redis

[ ] figure out expandTabs and parser.TabSizeEight. Are those used?

[ ] SoftbreakData is not used
