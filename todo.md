# Things to do

[x] `go test` is very slow because `doTestsReference()` tests processing of every substring of original markdown. Remove that and add a separate test program for that to be run only in CI

[x] instead of using NodeType, use interface{} for Node data, like xml parser

[x] simplify oliPrefix()

[ ] add examples like https://godoc.org/github.com/dgrijalva/jwt-go (put in foo_example_test.go). Or see https://github.com/garyburd/redigo/blob/master/redis/zpop_example_test.go#L5 / https://godoc.org/github.com/garyburd/redigo/redis or https://godoc.org/github.com/go-redis/redis

[ ] figure out expandTabs and parser.TabSizeEight. Are those used?

[ ] add Options { Extensions: } for NewParser to allow for more customization in the future?

[ ] speed: Node is probably too fat

[ ] speed: Node could be allocated from a slice and pointers could be replaces with int32 integers. Node would have to keep track of the allocator or we can change the api to pass both Parser (which is allocator of nodes) and Node (which would be a typedef for int)

[ ] fuzzing!

[ ] SoftbreakData is not used

Non-containers:
type HorizontalRuleData struct {
type TextData struct {
type HTMLBlockData struct {
type CodeBlockData struct {
type SoftbreakData struct {
type HardbreakData struct {
type CodeData struct {
type HTMLSpanData struct {

Containers:
type DocumentData struct {
type BlockQuoteData struct {
type ListData struct {
type ListItemData struct {
type ParagraphData struct {
type HeadingData struct {
type EmphData struct {
type StrongData struct {
type DelData struct {
type LinkData struct {
type ImageData struct {
type TableData struct {
type TableHeadData struct {
type TableBodyData struct {
type TableRowData struct {
type TableCellData struct {

type Node interface {
    Parent() *Node
    Children() []*Node
    IsContainer() bool // ???
}

type NodeCommon {
    Parent *Node
    Children []*Node // only needed for container nodes
	Literal []byte // Text contents of the leaf nodes TODO: move to only leaf nodes
    content []byte // Markdown content of the block nodes
	open    bool   // Specifies an open block node that has not been finished to process yet
}

type DocumentData struct {
    NodeCommon
}

type BlockQuoteData struct {
    NodeCommon
}
