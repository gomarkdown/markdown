package markdown

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

type testData struct {
	md   []byte
	html []byte
}

func testDataToStrArray(tests []*testData) []string {
	res := []string{}
	for _, td := range tests {
		res = append(res, string(td.md))
		res = append(res, string(td.html))
	}
	return res
}

func readTestFile2(t *testing.T, fileName string) []string {
	tests := readTestFile(t, fileName)
	return testDataToStrArray(tests)
}

func readTestFile(t *testing.T, fileName string) []*testData {
	path := filepath.Join("testdata", fileName)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("ioutil.ReadFile('%s') failed with %s", path, err)
	}
	parts := bytes.Split(d, []byte("+++\n"))
	if len(parts)%2 != 0 {
		t.Fatalf("odd test tuples in file %s: %d", path, len(parts))
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
	return res
}

func TestMmark(t *testing.T) {
	testData := readTestFile(t, "mmark.test")
	ext := parser.CommonExtensions | parser.Attributes | parser.OrderedListStart | parser.SuperSubscript | parser.Mmark
	for _, td := range testData {
		p := parser.NewWithExtensions(ext)

		got := ToHTML(td.md, p, nil)
		want := td.html

		if !bytes.Equal(got, want) {
			t.Errorf("want (%d bytes) %s, got (%d bytes) %s, for input %q", len(want), want, len(got), got, td.md)
		}
	}
}
