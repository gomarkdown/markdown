package ast

import (
	"bytes"
	"fmt"
)

// ListType contains bitwise or'ed flags for list and list item objects.
type ListType int

// These are the possible flag values for the ListItem renderer.
// Multiple flag values may be ORed together.
// These are mostly of interest if you are writing a new output format.
const (
	ListTypeOrdered ListType = 1 << iota
	ListTypeDefinition
	ListTypeTerm

	ListItemContainsBlock
	ListItemBeginningOfList // TODO: figure out if this is of any use now
	ListItemEndOfList
)

// CellAlignFlags holds a type of alignment in a table cell.
type CellAlignFlags int

// These are the possible flag values for the table cell renderer.
// Only a single one of these values will be used; they are not ORed together.
// These are mostly of interest if you are writing a new output format.
const (
	TableAlignmentLeft CellAlignFlags = 1 << iota
	TableAlignmentRight
	TableAlignmentCenter = (TableAlignmentLeft | TableAlignmentRight)
)

// Node defines an ast node
type Node interface {
	AsTreeNode() *TreeNode
	GetParent() Node
	SetParent(newParent Node)
	GetChildren() []Node
	SetChildren(newChildren []Node)
	FirstChild() Node
	LastChild() Node
}

// TreeNode is a common part of all nodes, used to represent tree and contain
// data that all nodes have in common
type TreeNode struct {
	Parent   Node
	Children []Node

	Literal []byte // Text contents of the leaf nodes
	Content []byte // Markdown content of the block nodes
}

// AsTreeNode returns itself as *TreeNode
func (n *TreeNode) AsTreeNode() *TreeNode {
	res := n
	//fmt.Printf("TreeNode.AsTreeNode() called. n: %p, res: %p %v\n", n, res, res)
	return res
}

// GetParent returns parent
func (n *TreeNode) GetParent() Node {
	return n.Parent
}

// SetParent sets the parent
func (n *TreeNode) SetParent(newParent Node) {
	n.Parent = newParent
}

// GetChildren returns children
func (n *TreeNode) GetChildren() []Node {
	return n.Children
}

// SetChildren sets children
func (n *TreeNode) SetChildren(newChildren []Node) {
	n.Children = newChildren
}

// LeafNode is a common part of all nodes, used to represent tree and contain
// data that all nodes have in common
type LeafNode struct {
	Parent Node

	Literal []byte // Text contents of the leaf nodes
	Content []byte // Markdown content of the block nodes
}

// AsTreeNode returns itself as *TreeNode
func (n *LeafNode) AsTreeNode() *TreeNode {
	return nil
}

// GetParent returns parent
func (n *LeafNode) GetParent() Node {
	return n.Parent
}

// SetParent sets the parent
func (n *LeafNode) SetParent(newParent Node) {
	n.Parent = newParent
}

// GetChildren returns children
func (n *LeafNode) GetChildren() []Node {
	return nil
}

// SetChildren sets children
func (n *LeafNode) SetChildren(newChildren []Node) {
	// do nothing, LeafNode has no children
}

// FirstChild returns children
func (n *LeafNode) FirstChild() Node {
	return nil
}

// LastChild returns children
func (n *LeafNode) LastChild() Node {
	return nil
}

// PanicIfTreeNode will panic if node is *TreeNode
func PanicIfTreeNode(node Node) {
	if _, ok := node.(*TreeNode); ok {
		panic(fmt.Sprintf("%v is TreeNode", node))
	}
}

// AddChild adds child node to parent node
func AddChild(parent Node, child Node) {
	PanicIfTreeNode(parent)
	PanicIfTreeNode(child)
	pn := parent.AsTreeNode()
	pn.Parent = parent
	pn.Children = append(pn.Children, child)
}

// Document represents document
type Document struct {
	TreeNode
}

// BlockQuote represents data for block quote node
type BlockQuote struct {
	TreeNode
}

// List represents data list node
type List struct {
	TreeNode

	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// ListItem represents data for list item node
type ListItem struct {
	TreeNode

	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// Paragraph represents data for paragraph node
type Paragraph struct {
	TreeNode
}

// Heading contains fields relevant to a Heading node type.
type Heading struct {
	TreeNode

	Level        int    // This holds the heading level number
	HeadingID    string // This might hold heading ID, if present
	IsTitleblock bool   // Specifies whether it's a title block
}

// HorizontalRule represents data for horizontal rule node
type HorizontalRule struct {
	LeafNode
}

// Emph represents data for emp node
type Emph struct {
	TreeNode
}

// Strong represents data for strong node
type Strong struct {
	TreeNode
}

// Del represents data for del node
type Del struct {
	TreeNode
}

// Link represents data for link node
type Link struct {
	TreeNode

	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
	NoteID      int    // NoteID contains a serial number of a footnote, zero if it's not a footnote
	Footnote    Node   // If it's a footnote, this is a direct link to the footnote Node. Otherwise nil.
}

// Image represents data for image node
type Image struct {
	TreeNode

	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
}

// Text represents data for text node
type Text struct {
	LeafNode
}

// HTMLBlock represents data for html node
type HTMLBlock struct {
	LeafNode
}

// CodeBlock contains fields relevant to a CodeBlock node type.
type CodeBlock struct {
	LeafNode

	IsFenced    bool   // Specifies whether it's a fenced code block or an indented one
	Info        []byte // This holds the info string
	FenceChar   byte
	FenceLength int
	FenceOffset int
}

// Softbreak represents data for softbreak node
// Note: not used currently
type Softbreak struct {
	LeafNode
}

// Hardbreak represents data for hard break node
type Hardbreak struct {
	LeafNode
}

// Code represents data for code node
type Code struct {
	LeafNode
}

// HTMLSpan represents data for html span node
type HTMLSpan struct {
	LeafNode
}

// Table represents data for table node
type Table struct {
	TreeNode
}

// TableCell contains fields relevant to a table cell node type.
type TableCell struct {
	TreeNode

	IsHeader bool           // This tells if it's under the header row
	Align    CellAlignFlags // This holds the value for align attribute
}

// TableHead represents data for a table head node
type TableHead struct {
	TreeNode
}

// TableBody represents data for a tablef body node
type TableBody struct {
	TreeNode
}

// TableRow represents data for a table row node
type TableRow struct {
	TreeNode
}

/*
func (n *Node) String() string {
	ellipsis := ""
	snippet := n.Literal
	if len(snippet) > 16 {
		snippet = snippet[:16]
		ellipsis = "..."
	}
	return fmt.Sprintf("%T: '%s%s'", n.Data, snippet, ellipsis)
}
*/

func removeNodeFromArray(a []Node, node Node) []Node {
	n := len(a)
	for i := 0; i < n; i++ {
		if a[i] == node {
			return append(a[:i], a[i+1:]...)
		}
	}
	return nil
}

// RemoveFromTree removes this node from tree
func RemoveFromTree(n Node) {
	if n.GetParent() == nil {
		return
	}
	// important: don't clear n.Children if n has no parent
	// we're called from AppendChild and that might happen on a node
	// that accumulated Children but hasn't been inserted into the tree
	n.SetChildren(nil)
	p := n.GetParent()
	newChildren := removeNodeFromArray(p.GetChildren(), n)
	if newChildren != nil {
		p.SetChildren(newChildren)
	}
}

// AppendChild adds a node 'child' as a child of 'n'.
// It panics if either node is nil.
func AppendChild(n Node, child Node) {
	RemoveFromTree(child)
	child.SetParent(n)
	newChildren := append(n.GetChildren(), child)
	n.SetChildren(newChildren)
}

func isContainer(n Node) bool {
	return n.AsTreeNode() != nil
}

// LastChild returns last child of this node
func (n *TreeNode) LastChild() Node {
	a := n.Children
	if len(a) > 0 {
		return a[len(a)-1]
	}
	return nil
}

// FirstChild returns first child of this node
func (n *TreeNode) FirstChild() Node {
	a := n.Children
	if len(a) > 0 {
		return a[0]
	}
	return nil
}

// NextNode returns next sibling of this node
// We can't make it part of TreeNode or LeafNode because we loose Node identity
func NextNode(n Node) Node {
	parent := n.GetParent()
	if parent == nil {
		return nil
	}
	a := parent.GetChildren()
	len := len(a) - 1
	for i := 0; i < len; i++ {
		if a[i] == n {
			return a[i+1]
		}
	}
	return nil
}

// PrevNode returns sibling node before n
// We can't make it part of TreeNode or LeafNode because we loose Node identity
func PrevNode(n Node) Node {
	parent := n.GetParent()
	if parent == nil {
		return nil
	}
	a := parent.GetChildren()
	len := len(a)
	for i := 1; i < len; i++ {
		if a[i] == n {
			return a[i-1]
		}
	}
	return nil
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
	Visit(node Node, entering bool) WalkStatus
}

// NodeVisitorFunc casts a function to match NodeVisitor interface
type NodeVisitorFunc func(node Node, entering bool) WalkStatus

// Visit calls visitor function
func (f NodeVisitorFunc) Visit(node Node, entering bool) WalkStatus {
	return f(node, entering)
}

// Walk is a convenience method that instantiates a walker and starts a
// traversal of subtree rooted at n.
func Walk(n Node, visitor NodeVisitor) {
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
func WalkFunc(n Node, f NodeVisitorFunc) {
	visitor := NodeVisitorFunc(f)
	Walk(n, visitor)
}

type nodeWalker struct {
	current  Node
	root     Node
	entering bool
}

func newNodeWalker(root Node) *nodeWalker {
	return &nodeWalker{
		current:  root,
		root:     root,
		entering: true,
	}
}

func (nw *nodeWalker) next() {
	isCont := isContainer(nw.current)
	if (!isCont || !nw.entering) && nw.current == nw.root {
		nw.current = nil
		return
	}
	if nw.entering && isCont {
		firstChild := nw.current.FirstChild()
		if firstChild != nil {
			nw.current = firstChild
			nw.entering = true
		} else {
			nw.entering = false
		}
	} else if NextNode(nw.current) == nil {
		nw.current = nw.current.GetParent()
		nw.entering = false
	} else {
		nw.current = NextNode(nw.current)
		nw.entering = true
	}
}

func dump(ast Node) {
	fmt.Println(dumpString(ast))
}

func dumpR(ast Node, depth int) string {
	if ast == nil {
		return ""
	}
	indent := bytes.Repeat([]byte("\t"), depth)
	content := ast.AsTreeNode().Literal
	if content == nil {
		content = ast.AsTreeNode().Content
	}
	result := fmt.Sprintf("%s%T(%q)\n", indent, ast, content)
	for _, n := range ast.GetChildren() {
		result += dumpR(n, depth+1)
	}
	return result
}

func dumpString(ast Node) string {
	return dumpR(ast, 0)
}
