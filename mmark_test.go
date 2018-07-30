package markdown

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestMmark(t *testing.T) {
	testfile := filepath.Join("testdata", "mmark.test")

	data, err := ioutil.ReadFile(testfile)
	if err != nil {
		t.Fatalf("failed to open file %q: %s", testfile, err)
	}

	testdata := bytes.Split(data, []byte("---\n"))
	if len(testdata)%2 != 0 {
		t.Fatalf("odd test tuples: %d", len(testdata))
	}
	for i := 0; i < len(testdata); i += 2 {
		ext := parser.CommonExtensions | parser.Attributes | parser.OrderedListStart | parser.MmarkSpecialHeading
		parser := parser.NewWithExtensions(ext)

		input := testdata[i]
		want := testdata[i+1]

		got := ToHTML([]byte(input), parser, nil)

		if bytes.Compare(got, want) != 0 {
			t.Errorf("want %s, got %s, for input %q", want, got, input)
		}
	}
}
