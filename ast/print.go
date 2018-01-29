package ast

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Print is for debugging. It prints a string representation of parsed
// markdown doc (result of parser.Parse()) to dst.
//
// To make output readable, it shortens text output.
func Print(dst io.Writer, doc Node) {
	PrintWithPrefix(dst, doc, "  ")
}

// PrintWithPrefix is like Pring but allows customizing prefix used for
// indentation. By default it's 2 spaces. You can change it to e.g. tab
// by passing "\t"
func PrintWithPrefix(w io.Writer, doc Node, prefix string) {
	printRecur(w, doc, prefix, 0)
}

// ToString is like Dump but returns result as a string
func ToString(doc Node) string {
	var buf bytes.Buffer
	Print(&buf, doc)
	return buf.String()
}

func contentToString(d1 []byte, d2 []byte) string {
	if d1 != nil {
		return string(d1)
	}
	if d2 != nil {
		return string(d2)
	}
	return ""
}

func getContent(node Node) string {
	if c := node.AsContainer(); c != nil {
		return contentToString(c.Literal, c.Content)
	}
	leaf := node.AsLeaf()
	return contentToString(leaf.Literal, leaf.Content)
}

func shortenString(s string, maxLen int) string {
	// for cleaner, one-line ouput, replace some white-space chars
	// with their escaped version
	s = strings.Replace(s, "\n", `\n`, -1)
	s = strings.Replace(s, "\r", `\r`, -1)
	s = strings.Replace(s, "\t", `\t`, -1)
	if maxLen < 0 {
		return s
	}
	if len(s) < maxLen {
		return s
	}
	// add "..." to indicate truncation
	return s[:maxLen-3] + "..."
}

// get a short name of the type of v which excludes package name
// and strips "()" from the end
func getNodeType(node Node) string {
	s := fmt.Sprintf("%T", node)
	s = strings.TrimSuffix(s, "()")
	if idx := strings.Index(s, "."); idx != -1 {
		return s[idx+1:]
	}
	return s
}

func printRecur(w io.Writer, node Node, prefix string, depth int) {
	if node == nil {
		return
	}
	indent := strings.Repeat(prefix, depth)
	io.WriteString(w, indent)

	content := shortenString(getContent(node), 40)
	typeName := getNodeType(node)
	fmt.Fprintf(w, "%s%s '%s'\n", indent, typeName, content)
	for _, child := range node.GetChildren() {
		printRecur(w, child, prefix, depth+1)
	}
}
