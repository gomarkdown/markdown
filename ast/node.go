package ast

import (
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
	AsContainer() *Container
	AsLeaf() *Leaf
	GetParent() Node
	SetParent(newParent Node)
	GetChildren() []Node
	SetChildren(newChildren []Node)
}

// Container is a type of node that can contain children
type Container struct {
	Parent   Node
	Children []Node

	Literal []byte // Text contents of the leaf nodes
	Content []byte // Markdown content of the block nodes
}

// AsContainer returns itself as *Container
func (c *Container) AsContainer() *Container {
	return c
}

// AsLeaf returns nil
func (c *Container) AsLeaf() *Leaf {
	return nil
}

// GetParent returns parent
func (c *Container) GetParent() Node {
	return c.Parent
}

// SetParent sets the parent
func (c *Container) SetParent(newParent Node) {
	c.Parent = newParent
}

// GetChildren returns children
func (c *Container) GetChildren() []Node {
	return c.Children
}

// SetChildren sets children
func (c *Container) SetChildren(newChildren []Node) {
	c.Children = newChildren
}

// Leaf is a node that cannot have children
type Leaf struct {
	Parent Node

	Literal []byte // Text contents of the leaf nodes
	Content []byte // Markdown content of the block nodes
}

// AsContainer returns itself as *Container
func (l *Leaf) AsContainer() *Container {
	return nil
}

// AsLeaf returns itself as leaf
func (l *Leaf) AsLeaf() *Leaf {
	return l
}

// GetParent returns parent
func (l *Leaf) GetParent() Node {
	return l.Parent
}

// SetParent sets the parent
func (l *Leaf) SetParent(newParent Node) {
	l.Parent = newParent
}

// GetChildren returns children
func (l *Leaf) GetChildren() []Node {
	return nil
}

// SetChildren sets children
func (l *Leaf) SetChildren(newChildren []Node) {
	// do nothing, Leaf has no children
}

// PanicIfContainer will panic if node is *Container
func PanicIfContainer(node Node) {
	if _, ok := node.(*Container); ok {
		panic(fmt.Sprintf("%v is Container", node))
	}
}

// AddChild adds child node to parent node
func AddChild(parent Node, child Node) {
	PanicIfContainer(parent)
	PanicIfContainer(child)
	pn := parent.AsContainer()
	pn.Parent = parent
	pn.Children = append(pn.Children, child)
}

// Document represents document
type Document struct {
	Container
}

// BlockQuote represents data for block quote node
type BlockQuote struct {
	Container
}

// List represents data list node
type List struct {
	Container

	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// ListItem represents data for list item node
type ListItem struct {
	Container

	ListFlags       ListType
	Tight           bool   // Skip <p>s around list item data if true
	BulletChar      byte   // '*', '+' or '-' in bullet lists
	Delimiter       byte   // '.' or ')' after the number in ordered lists
	RefLink         []byte // If not nil, turns this list item into a footnote item and triggers different rendering
	IsFootnotesList bool   // This is a list of footnotes
}

// Paragraph represents data for paragraph node
type Paragraph struct {
	Container
}

// Heading contains fields relevant to a Heading node type.
type Heading struct {
	Container

	Level        int    // This holds the heading level number
	HeadingID    string // This might hold heading ID, if present
	IsTitleblock bool   // Specifies whether it's a title block
}

// HorizontalRule represents data for horizontal rule node
type HorizontalRule struct {
	Leaf
}

// Emph represents data for emp node
type Emph struct {
	Container
}

// Strong represents data for strong node
type Strong struct {
	Container
}

// Del represents data for del node
type Del struct {
	Container
}

// Link represents data for link node
type Link struct {
	Container

	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
	NoteID      int    // NoteID contains a serial number of a footnote, zero if it's not a footnote
	Footnote    Node   // If it's a footnote, this is a direct link to the footnote Node. Otherwise nil.
}

// Image represents data for image node
type Image struct {
	Container

	Destination []byte // Destination is what goes into a href
	Title       []byte // Title is the tooltip thing that goes in a title attribute
}

// Text represents data for text node
type Text struct {
	Leaf
}

// HTMLBlock represents data for html node
type HTMLBlock struct {
	Leaf
}

// CodeBlock contains fields relevant to a CodeBlock node type.
type CodeBlock struct {
	Leaf

	IsFenced    bool   // Specifies whether it's a fenced code block or an indented one
	Info        []byte // This holds the info string
	FenceChar   byte
	FenceLength int
	FenceOffset int
}

// Softbreak represents data for softbreak node
// Note: not used currently
type Softbreak struct {
	Leaf
}

// Hardbreak represents data for hard break node
type Hardbreak struct {
	Leaf
}

// Code represents data for code node
type Code struct {
	Leaf
}

// HTMLSpan represents data for html span node
type HTMLSpan struct {
	Leaf
}

// Table represents data for table node
type Table struct {
	Container
}

// TableCell contains fields relevant to a table cell node type.
type TableCell struct {
	Container

	IsHeader bool           // This tells if it's under the header row
	Align    CellAlignFlags // This holds the value for align attribute
}

// TableHead represents data for a table head node
type TableHead struct {
	Container
}

// TableBody represents data for a tablef body node
type TableBody struct {
	Container
}

// TableRow represents data for a table row node
type TableRow struct {
	Container
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

// IsContainer returns true if n is a container node (i.e. can have children,
// as opposed to leaf node)
func IsContainer(n Node) bool {
	return n.AsContainer() != nil
}

// LastChild returns last child of this node
// It's implemented as stand-alone function to keep Node interface small
func LastChild(n Node) Node {
	a := n.GetChildren()
	if len(a) > 0 {
		return a[len(a)-1]
	}
	return nil
}

// FirstChild returns first child of this node
// It's implemented as stand-alone function to keep Node interface small
func FirstChild(n Node) Node {
	a := n.GetChildren()
	if len(a) > 0 {
		return a[0]
	}
	return nil
}

// NextNode returns next sibling of this node
// We can't make it part of Container or Leaf because we loose Node identity
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
// We can't make it part of Container or Leaf because we loose Node identity
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

// Walk traverses tree recursively
func Walk(n Node, visitor NodeVisitor) WalkStatus {
	isContainer := IsContainer(n)
	status := visitor.Visit(n, true) // entering
	if status == Terminate {
		// even if terminating, close container node
		if isContainer {
			visitor.Visit(n, false)
		}
		return status
	}
	if isContainer && status != SkipChildren {
		children := n.GetChildren()
		for _, n := range children {
			status = Walk(n, visitor)
			if status == Terminate {
				return status
			}
		}
	}
	if isContainer {
		status = visitor.Visit(n, false) // exiting
		if status == Terminate {
			return status
		}
	}
	return GoToNext
}

// Visit calls visitor function
func (f NodeVisitorFunc) Visit(node Node, entering bool) WalkStatus {
	return f(node, entering)
}

// WalkFunc is like Walk but accepts just a callback function
func WalkFunc(n Node, f NodeVisitorFunc) {
	visitor := NodeVisitorFunc(f)
	Walk(n, visitor)
}
