package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

// Mmark tests

func TestMmark(t *testing.T) {
	t.Parallel()
	files := []string{
		"mmark/test1",
	}
	doTestsReference(t, files, parser.CommonExtensions|parser.Mmark)
}
