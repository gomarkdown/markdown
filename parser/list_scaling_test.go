package parser

import (
	"strings"
	"testing"
	"time"
)

// buildListHeavyDoc builds many top-level lists with nested sublists. It guards
// the complete list parser against accidental super-linear tree walks.
func buildListHeavyDoc(nList int) []byte {
	var b strings.Builder
	for i := 0; i < nList; i++ {
		b.WriteString("- a\n  - a1\n  - a2\n- b\n- c\n\n")
		b.WriteString("paragraph\n\n")
	}
	return []byte(b.String())
}

func TestListHeavyDocumentScaling(t *testing.T) {
	doc := buildListHeavyDoc(800)
	start := time.Now()
	New().Parse(doc)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("parsing list-heavy document took %s; possible O(N²) tree walk", elapsed)
	}
}

func BenchmarkListHeavyDocument(b *testing.B) {
	doc := buildListHeavyDoc(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New().Parse(doc)
	}
}
