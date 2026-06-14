package parser

import (
	"strings"
	"testing"
	"time"
)

// buildListHeavyDoc builds a markdown document made of many top-level lists
// (each with a nested sublist) separated by paragraphs. This stresses
// finalizeList(): the buggy version iterated over the list's *sibling* blocks
// (list.Parent.GetChildren()) instead of the list's own items, and for each
// sibling that is itself a list it descended via endsWithBlankLine(). That made
// parsing scale super-linearly with the number of top-level lists (tens of
// seconds on large real-world documents). The fix iterates the list's own
// items, making parsing linear.
func buildListHeavyDoc(nList int) []byte {
	var b strings.Builder
	for i := 0; i < nList; i++ {
		b.WriteString("- a\n  - a1\n  - a2\n- b\n- c\n\n")
		b.WriteString("paragraph\n\n")
	}
	return []byte(b.String())
}

// TestFinalizeListScaling guards against finalizeList() regressing to the old
// super-linear behavior. With the buggy code, parsing this document took
// several seconds; the fixed code parses it in a few milliseconds. The bound is
// generous so it only fires if the quadratic blowup is reintroduced.
func TestFinalizeListScaling(t *testing.T) {
	doc := buildListHeavyDoc(800)
	start := time.Now()
	New().Parse(doc)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("parsing list-heavy doc took %s; finalizeList may have regressed to O(N^2)", elapsed)
	}
}

func BenchmarkFinalizeListHeavyDoc(b *testing.B) {
	doc := buildListHeavyDoc(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New().Parse(doc)
	}
}
