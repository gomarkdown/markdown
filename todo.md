# Things to do

[x] `go test` is very slow because `doTestsReference()` tests processing of every substring of original markdown. Remove that and add a separate test program for that to be run only in CI

[ ] instead of using NodeType, use interface{} for Node data, like xml parser

[ ] speed: Node is probably too fat

[ ] speed: Node could be allocated from a slice and pointers could be replaces with int32 integers. Node would have to keep track of the allocator or we can change the api to pass both Parser (which is allocator of nodes) and Node (which would be a typedef for int)

[ ] fuzzing!

[ ] NodeData could distinguish between data that requires pointers (like ItemData) and data that can be value (empty stuff like DocumentData). Would make it explicit in name i.e. \*Ptr has to be always accessed as (*ItemDataPtr) and non-pointer as e.g. .(DocumentData)

[ ] simplify oliPrefix()

[ ] SoftbreakData is not used
