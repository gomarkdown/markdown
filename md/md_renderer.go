package md

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// Renderer renders to markdown. Allows converting to a canonical
// form.
type Renderer struct {
	orderedListCounter map[int]int

	lastOutputLen  int
	listDepth      int
	lastNormalText string

	C *RendererConfig

	linkcache map[string]bool // cache for link definitions to write in the footer, if renderLinksInFooter is set
}

const listIndentSize = 4

type RendererConfig struct {
	Flags Flags
}

type Flags int

const renderLinksInFooter Flags = 1 << iota

type RendererOpt func(c *RendererConfig)

// NewRenderer returns a Markdown renderer.
func NewRenderer(opts ...RendererOpt) *Renderer {
	c := &RendererConfig{}
	for _, opt := range opts {
		opt(c)
	}
	return &Renderer{
		orderedListCounter: map[int]int{},
		C:                  c,
	}
}

func WithRenderInFooter(renderInFooter bool) RendererOpt {
	return func(c *RendererConfig) {
		if renderInFooter {
			c.Flags |= renderLinksInFooter
		} else {
			c.Flags &^= renderLinksInFooter
		}
	}
}

func (r *Renderer) out(w io.Writer, d []byte) {
	r.lastOutputLen = len(d)
	w.Write(d)
}

func (r *Renderer) outs(w io.Writer, s string) {
	r.lastOutputLen = len(s)
	io.WriteString(w, s)
}

func (r *Renderer) list(w io.Writer, node *ast.List, entering bool) {
	if entering {
		r.listDepth++
		flags := node.ListFlags
		if flags&ast.ListTypeOrdered != 0 {
			r.orderedListCounter[r.listDepth] = 1
		}
	} else {
		r.listDepth--
		if _, ok := node.Parent.(*ast.ListItem); !ok {
			fmt.Fprintf(w, "\n")
		}
	}
}

func (r *Renderer) listItem(w io.Writer, node *ast.ListItem, entering bool) {
	if !entering {
		return
	}
	io.WriteString(w, strings.Repeat(" ", (r.listDepth-1)*listIndentSize))
	if node.ListFlags&ast.ListTypeOrdered != 0 {
		fmt.Fprintf(w, "%d. ", r.orderedListCounter[r.listDepth])
		r.orderedListCounter[r.listDepth]++
	} else {
		io.WriteString(w, string(node.BulletChar)+" ")
	}
}

func (r *Renderer) para(w io.Writer, node *ast.Paragraph, entering bool) {
	if !entering && r.lastOutputLen > 0 {
		br := "\n\n"

		// List items don't need the extra line-break.
		if _, ok := node.Parent.(*ast.ListItem); ok {
			br = "\n"
		}

		r.outs(w, br)
	}
}

// escape replaces instances of backslash with escaped backslash in text.
func escape(text []byte) []byte {
	return bytes.Replace(text, []byte(`\`), []byte(`\\`), -1)
}

func isNumber(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func needsEscaping(text []byte, lastNormalText string) bool {
	switch string(text) {
	case `\`,
		"`",
		"*",
		"_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#",
		"+",
		"-":
		return true
	case "!":
		return false
	case ".":
		// Return true if number, because a period after a number must be escaped to not get parsed as an ordered list.
		return isNumber(lastNormalText)
	case "<", ">":
		return true
	default:
		return false
	}
}

// cleanWithoutTrim is like clean, but doesn't trim blanks.
func cleanWithoutTrim(s string) string {
	var b []byte
	var p byte
	for i := 0; i < len(s); i++ {
		q := s[i]
		if q == '\n' || q == '\r' || q == '\t' {
			q = ' '
		}
		if q != ' ' || p != ' ' {
			b = append(b, q)
			p = q
		}
	}
	return string(b)
}

func (r *Renderer) text(w io.Writer, text *ast.Text) {
	lit := text.Literal
	if needsEscaping(lit, r.lastNormalText) {
		lit = append([]byte("\\"), lit...)
	}
	r.lastNormalText = string(text.Literal)
	if r.listDepth > 0 && string(lit) == "\n" {
		// TODO: See if this can be cleaned up... It's needed for lists.
		return
	}
	cleanString := cleanWithoutTrim(string(lit))
	if cleanString == "" {
		return
	}
	r.outs(w, cleanString)
}

func (r *Renderer) htmlSpan(w io.Writer, node *ast.HTMLSpan) {
	r.out(w, node.Literal)
}

func (r *Renderer) htmlBlock(w io.Writer, node *ast.HTMLBlock) {
	r.outs(w, "\n")
	r.out(w, node.Literal)
	r.outs(w, "\n\n")
}

func (r *Renderer) codeBlock(w io.Writer, node *ast.CodeBlock) {
	r.outs(w, "\n")
	text := node.Literal
	language := ""
	for _, elt := range strings.Fields(string(node.Info)) {
		if elt[0] == '.' {
			elt = elt[1:]
		}
		if elt != "" {
			language = elt
			break
		}
	}
	r.outs(w, "```"+language+"\n")
	r.out(w, text)
	if len(text) == 0 || text[len(text)-1] != '\n' {
		r.outs(w, "\n")
	}
	r.outs(w, "```\n\n")
}

func (r *Renderer) code(w io.Writer, node *ast.Code) {
	r.outs(w, "`"+string(node.Literal)+"`")
}

func (r *Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		r.outs(w, strings.Repeat("#", node.Level)+" ")
		r.out(w, node.Literal)
	} else {
		r.outs(w, "\n\n")
	}
}

func (r *Renderer) image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		r.outs(w, "![")
	} else {
		link := node.Destination
		title := node.Title
		r.outs(w, "](")
		r.out(w, escape(link))
		if len(title) != 0 {
			r.outs(w, ` "`)
			r.out(w, title)
			r.outs(w, `"`)
		}
		r.outs(w, ")")
	}
}

func linkPlainText(node *ast.Link) []byte {
	var b []byte
	for _, c := range node.GetChildren() {
		t, ok := c.(*ast.Text)
		if !ok {
			return nil
		}
		b = append(b, t.Literal...)
	}
	return b
}

func (r *Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if entering {
		r.outs(w, "[")
	} else if len(node.DeferredID) > 0 && (r.C == nil || r.C.Flags&renderLinksInFooter == 0) {
		if bytes.EqualFold(linkPlainText(node), node.DeferredID) {
			r.outs(w, "]")
		} else {
			r.outs(w, "][")
			r.out(w, escape(node.DeferredID))
			r.outs(w, "]")
		}
	} else {
		link := string(escape(node.Destination))
		title := string(node.Title)
		if r.C == nil || r.C.Flags&renderLinksInFooter == 0 {
			r.outs(w, "](")
			r.outs(w, link)
			if len(title) != 0 {
				r.outs(w, ` "`)
				r.outs(w, title)
				r.outs(w, `"`)
			}
			r.outs(w, ")")
			return
		}

		r.outs(w, "]")
		child, _ := ast.GetFirstChild(node).(*ast.Text)
		linkdefn := fmt.Sprintf("[%s]: %s", string(escape(child.Leaf.Literal)), link)
		if len(title) != 0 {
			linkdefn += fmt.Sprintf(" \"%s\"", title)
		}
		if r.linkcache == nil {
			r.linkcache = make(map[string]bool)
		}
		r.linkcache[linkdefn] = true

	}
}

// RenderNode renders markdown node
func (r *Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Text:
		r.text(w, node)
	case *ast.Emph:
		r.outs(w, "*")
	case *ast.Strong:
		r.outs(w, "**")
	case *ast.Del:
		r.outs(w, "~~")
	case *ast.Link:
		r.link(w, node, entering)
	case *ast.Image:
		r.image(w, node, entering)
	case *ast.Code:
		r.code(w, node)
	case *ast.CodeBlock:
		r.codeBlock(w, node)
	case *ast.Document:
		// do nothing
	case *ast.Paragraph:
		r.para(w, node, entering)
	case *ast.HTMLSpan:
		r.htmlSpan(w, node)
	case *ast.HTMLBlock:
		r.htmlBlock(w, node)
	case *ast.Heading:
		r.heading(w, node, entering)
	case *ast.List:
		r.list(w, node, entering)
	case *ast.ListItem:
		r.listItem(w, node, entering)
	case *ast.Footnotes:
		// nothing by default; just output the list.
	case *ast.ReferenceDefinition:
		if entering {
			r.outs(w, "[")
			r.out(w, node.Label)
			r.outs(w, "]: ")
			r.out(w, node.Destination)
			if len(node.Title) > 0 {
				r.outs(w, " \"")
				r.out(w, node.Title)
				r.outs(w, "\"")
			}
			r.outs(w, "\n")
		}
	default:
		panic(fmt.Sprintf("node %T NYI", node))
	}
	return ast.GoToNext
}

// RenderHeader renders header
func (r *Renderer) RenderHeader(_ io.Writer, _ ast.Node) {
	// do nothing
}

// RenderFooter renders footer
func (r *Renderer) RenderFooter(w io.Writer, _ ast.Node) {
	if r.C == nil || r.C.Flags&renderLinksInFooter == 0 || len(r.linkcache) == 0 {
		return
	}
	links := make([]string, 0, len(r.linkcache))
	for link := range r.linkcache {
		links = append(links, link)
	}
	sort.Strings(links)
	for _, link := range links {
		r.outs(w, "\n")
		r.outs(w, link)
	}
	r.outs(w, "\n")
}
