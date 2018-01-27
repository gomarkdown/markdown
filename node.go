package markdown

import (
	"bytes"
	"fmt"
)

// NodeData represents data field of Node
type NodeData interface{}

// DocumentData represents top-level document node
type DocumentData struct {
}

// BlockQuoteData represents data for block quote node
type BlockQuoteData struct {
}

// ListData represents data list node
type ListData struct {
	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// ListItemData represents data for list item node
type ListItemData struct {
	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// ParagraphData represents data for paragraph node
type ParagraphData struct {
}

// HeadingData contains fields relevant to a Heading node type.
type HeadingData struct {
	Level        int    // This holds the heading level number
	HeadingID    string // This might hold heading ID, if present
	IsTitleblock bool   // Specifies whether it's a title block
}

// HorizontalRuleData represents data for horizontal rule node
type HorizontalRuleData struct {
}

// EmphData represents data for emp node
type EmphData struct {
}

// StrongData represents data for strong node
type StrongData struct {
}

// DelData represents data for del node
type DelData struct {
}

// LinkData represents data for link node
type LinkData struct {
	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
	NoteID      int    // NoteID contains a serial number of a footnote, zero if it's not a footnote
	Footnote    *Node  // If it's a footnote, this is a direct link to the footnote Node. Otherwise nil.
}

// ImageData represents data for image node
type ImageData struct {
	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
}

// TextData represents data for text node
type TextData struct {
}

// HTMLBlockData represents data for html node
type HTMLBlockData struct {
}

// CodeBlockData contains fields relevant to a CodeBlock node type.
type CodeBlockData struct {
	IsFenced    bool   // Specifies whether it's a fenced code block or an indented one
	Info        []byte // This holds the info string
	FenceChar   byte
	FenceLength int
	FenceOffset int
}

// SoftbreakData represents data for softbreak node
// Note: not used currently
type SoftbreakData struct {
}

// HardbreakData represents data for hard break node
type HardbreakData struct {
}

// CodeData represents data for code node
type CodeData struct {
}

// HTMLSpanData represents data for html span node
type HTMLSpanData struct {
}

// TableData represents data for table node
type TableData struct {
}

// TableCellData contains fields relevant to a table cell node type.
type TableCellData struct {
	IsHeader bool           // This tells if it's under the header row
	Align    CellAlignFlags // This holds the value for align attribute
}

// TableHeadData represents data for a table head node
type TableHeadData struct {
}

// TableBodyData represents data for a tablef body node
type TableBodyData struct {
}

// TableRowData represents data for a table row node
type TableRowData struct {
}

// Node is a single element in the abstract syntax tree of the parsed document.
// It holds connections to the structurally neighboring nodes and, for certain
// types of nodes, additional information that might be needed when rendering.
type Node struct {
	Parent   *Node // Points to the parent
	Children []*Node

	Literal []byte // Text contents of the leaf nodes

	Data NodeData

	content []byte // Markdown content of the block nodes
	open    bool   // Specifies an open block node that has not been finished to process yet
}

// NewNode allocates a node of a specified type.
func NewNode(d NodeData) *Node {
	return &Node{
		Data: d,
		open: true,
	}
}

func (n *Node) String() string {
	ellipsis := ""
	snippet := n.Literal
	if len(snippet) > 16 {
		snippet = snippet[:16]
		ellipsis = "..."
	}
	return fmt.Sprintf("%T: '%s%s'", n.Data, snippet, ellipsis)
}

func removeNodeFromArray(a []*Node, node *Node) []*Node {
	n := len(a)
	for i := 0; i < n; i++ {
		if a[i] == node {
			return append(a[:i], a[i+1:]...)
		}
	}
	return a
}

func removeNodeFromTree(n *Node) {
	if n.Parent == nil {
		return
	}
	// important: don't clear n.Children if n has no parent
	// we're called from AppendChild and that might happen on a node
	// that accumulated Children but hasn't been inserted into the tree
	n.Parent.Children = removeNodeFromArray(n.Parent.Children, n)
	n.Parent = nil
	n.Children = nil
}

// AppendChild adds a node 'child' as a child of 'n'.
// It panics if either node is nil.
func (n *Node) AppendChild(child *Node) {
	removeNodeFromTree(child)
	child.Parent = n
	n.Children = append(n.Children, child)
}

// LastChild returns last child of this node
func (n *Node) LastChild() *Node {
	a := n.Children
	if len(a) > 0 {
		return a[len(a)-1]
	}
	return nil
}

// FirstChild returns first child of this node
func (n *Node) FirstChild() *Node {
	a := n.Children
	if len(a) > 0 {
		return a[0]
	}
	return nil
}

// Next returns next sibling of this node
func (n *Node) Next() *Node {
	if n.Parent == nil {
		return nil
	}
	a := n.Parent.Children
	len := len(a) - 1
	for i := 0; i < len; i++ {
		if a[i] == n {
			return a[i+1]
		}
	}
	return nil
}

// Prev returns previous sibling of this node
func (n *Node) Prev() *Node {
	if n.Parent == nil {
		return nil
	}
	a := n.Parent.Children
	len := len(a)
	for i := 1; i < len; i++ {
		if a[i] == n {
			return a[i-1]
		}
	}
	return nil
}

func (n *Node) isContainer() bool {
	// list of non-containers is smaller so we check against that for speed
	switch n.Data.(type) {
	case *HorizontalRuleData, *TextData, *HTMLBlockData, *CodeBlockData, *SoftbreakData, *HardbreakData, *CodeData, *HTMLSpanData:
		return false
	default:
		return true
	}
}

func isListData(d NodeData) bool {
	_, ok := d.(*ListData)
	return ok
}

func isListTight(d NodeData) bool {
	if listData, ok := d.(*ListData); ok {
		return listData.Tight
	}
	return false
}

func isListItemData(d NodeData) bool {
	_, ok := d.(*ListItemData)
	return ok
}

func isListItemTerm(node *Node) bool {
	data, ok := node.Data.(*ListItemData)
	return ok && data.ListFlags&ListTypeTerm != 0
}

func isLinkData(d NodeData) bool {
	_, ok := d.(*LinkData)
	return ok
}

func isTableRowData(d NodeData) bool {
	_, ok := d.(*TableRowData)
	return ok
}

func isTableCellData(d NodeData) bool {
	_, ok := d.(*TableCellData)
	return ok
}

func isBlockQuoteData(d NodeData) bool {
	_, ok := d.(*BlockQuoteData)
	return ok
}

func isDocumentData(d NodeData) bool {
	_, ok := d.(*DocumentData)
	return ok
}

func (n *Node) canContain(v NodeData) bool {
	switch n.Data.(type) {
	case *ListData:
		return isListItemData(v)
	case *DocumentData, *BlockQuoteData, *ListItemData:
		return !isListItemData(v)
	case *TableData:
		switch v.(type) {
		case *TableHeadData, *TableBodyData:
			return true
		default:
			return false
		}
	case *TableHeadData, *TableBodyData:
		return isTableRowData(v)
	case *TableRowData:
		return isTableCellData(v)
	}
	return false
}

// WalkStatus allows NodeVisitor to have some control over the tree traversal.
// It is returned from NodeVisitor and different values allow Node.Walk to
// decide which node to go to next.
type WalkStatus int

const (
	// GoToNext is the default traversal of every node.
	GoToNext WalkStatus = iota
	// SkipChildren tells walker to skip all children of current node.
	SkipChildren
	// Terminate tells walker to terminate the traversal.
	Terminate
)

// NodeVisitor is a callback to be called when traversing the syntax tree.
// Called twice for every node: once with entering=true when the branch is
// first visited, then with entering=false after all the children are done.
type NodeVisitor interface {
	Visit(node *Node, entering bool) WalkStatus
}

// NodeVisitorFunc casts a function to match NodeVisitor interface
type NodeVisitorFunc func(node *Node, entering bool) WalkStatus

// Visit calls visitor function
func (f NodeVisitorFunc) Visit(node *Node, entering bool) WalkStatus {
	return f(node, entering)
}

// Walk is a convenience method that instantiates a walker and starts a
// traversal of subtree rooted at n.
func (n *Node) Walk(visitor NodeVisitor) {
	w := newNodeWalker(n)
	for w.current != nil {
		status := visitor.Visit(w.current, w.entering)
		switch status {
		case GoToNext:
			w.next()
		case SkipChildren:
			w.entering = false
			w.next()
		case Terminate:
			return
		}
	}
}

// WalkFunc is like Walk but accepts just a callback function
func (n *Node) WalkFunc(f NodeVisitorFunc) {
	visitor := NodeVisitorFunc(f)
	n.Walk(visitor)
}

type nodeWalker struct {
	current  *Node
	root     *Node
	entering bool
}

func newNodeWalker(root *Node) *nodeWalker {
	return &nodeWalker{
		current:  root,
		root:     root,
		entering: true,
	}
}

func (nw *nodeWalker) next() {
	if (!nw.current.isContainer() || !nw.entering) && nw.current == nw.root {
		nw.current = nil
		return
	}
	if nw.entering && nw.current.isContainer() {
		if nw.current.FirstChild() != nil {
			nw.current = nw.current.FirstChild()
			nw.entering = true
		} else {
			nw.entering = false
		}
	} else if nw.current.Next() == nil {
		nw.current = nw.current.Parent
		nw.entering = false
	} else {
		nw.current = nw.current.Next()
		nw.entering = true
	}
}

func dump(ast *Node) {
	fmt.Println(dumpString(ast))
}

func dumpR(ast *Node, depth int) string {
	if ast == nil {
		return ""
	}
	indent := bytes.Repeat([]byte("\t"), depth)
	content := ast.Literal
	if content == nil {
		content = ast.content
	}
	result := fmt.Sprintf("%s%T(%q)\n", indent, ast.Data, content)
	for _, n := range ast.Children {
		result += dumpR(n, depth+1)
	}
	return result
}

func dumpString(ast *Node) string {
	return dumpR(ast, 0)
}
