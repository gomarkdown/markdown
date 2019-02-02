package markdown

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

type testData struct {
	md   []byte
	html []byte
}

func readTestFile(path string) ([]*testData, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(d, []byte("+++\n"))
	if len(parts)%2 != 0 {
		return nil, fmt.Errorf("odd test tuples in file %s: %d", path, len(parts))
	}
	res := []*testData{}
	n := len(parts) / 2
	for i := 0; i < n; i++ {
		j := i * 2
		td := &testData{
			md:   parts[j],
			html: parts[j+1],
		}
		res = append(res, td)
	}
	return res, nil
}

func TestMmark(t *testing.T) {
	path := filepath.Join("testdata", "mmark.test")
	testData, err := readTestFile(path)
	if err != nil {
		t.Fatalf("readTestFile() failed with %s", err)
	}

	ext := parser.CommonExtensions | parser.Attributes | parser.OrderedListStart | parser.SuperSubscript | parser.Mmark
	for _, td := range testData {
		p := parser.NewWithExtensions(ext)

		got := ToHTML(td.md, p, nil)
		want := td.html

		if bytes.Compare(got, want) != 0 {
			t.Errorf("want (%d bytes) %s, got (%d bytes) %s, for input %q", len(want), want, len(got), got, td.md)
		}
	}
}
