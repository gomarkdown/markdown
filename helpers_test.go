package markdown

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type TestParams struct {
	extensions        parser.Extensions
	referenceOverride parser.ReferenceOverrideFunc
	html.Flags
	html.RendererOptions
}

func runMarkdown(input string, params TestParams) string {
	isSafeURL := newSafeURLOverride([]string{"monero:", "bitcoin:"})

	params.RendererOptions.Flags = params.Flags
	parser := parser.NewWithExtensions(params.extensions)
	parser.IsSafeURLOverride = isSafeURL
	parser.ReferenceOverride = params.referenceOverride
	renderer := html.NewRenderer(params.RendererOptions)
	renderer.IsSafeURLOverride = isSafeURL

	d := ToHTML([]byte(input), parser, renderer)
	return string(d)
}

// doTests runs full document tests using MarkdownCommon configuration.
func doTests(t *testing.T, tests []string) {
	doTestsParam(t, tests, TestParams{
		extensions: parser.CommonExtensions,
		RendererOptions: html.RendererOptions{
			Flags: html.CommonFlags,
		},
	})
}

func doTestsBlock(t *testing.T, path string, extensions parser.Extensions) {
	tests := readTestFile2(t, path)
	doTestsParam(t, tests, TestParams{
		extensions: extensions,
		Flags:      html.UseXHTML,
	})
}

func doTestsParam(t *testing.T, tests []string, params TestParams) {
	for i := 0; i+1 < len(tests); i += 2 {
		input := tests[i]
		expected := tests[i+1]
		got := runMarkdown(input, params)
		if got != expected {
			t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\nInput:\n%s\nExpected:\n%s\nGot:\n%s\n",
				input, expected, got, input, expected, got)
		}
	}
}

func doTestsInline(t *testing.T, tests []string) {
	doTestsInlineParam(t, tests, TestParams{})
}

func doLinkTestsInline(t *testing.T, tests []string) {
	doTestsInline(t, tests)

	prefix := "http://localhost"
	params := html.RendererOptions{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	doTestsInlineParam(t, transformTests, TestParams{
		RendererOptions: params,
	})
	doTestsInlineParam(t, transformTests, TestParams{
		Flags:           html.UseXHTML,
		RendererOptions: params,
	})
}

func doSafeTestsInline(t *testing.T, tests []string) {
	doTestsInlineParam(t, tests, TestParams{Flags: html.Safelink})

	// All the links in this test should not have the prefix appended, so
	// just rerun it with different parameters and the same expectations.
	prefix := "http://localhost"
	params := html.RendererOptions{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	doTestsInlineParam(t, transformTests, TestParams{
		Flags:           html.Safelink,
		RendererOptions: params,
	})
}

func doTestsInlineParam(t *testing.T, tests []string, params TestParams) {
	params.extensions |= parser.Autolink | parser.Strikethrough
	params.Flags |= html.UseXHTML
	doTestsParam(t, tests, params)
}

func transformLinks(tests []string, prefix string) []string {
	newTests := make([]string, len(tests))
	anchorRe := regexp.MustCompile(`<a href="/(.*?)"`)
	imgRe := regexp.MustCompile(`<img src="/(.*?)"`)
	for i, test := range tests {
		if i%2 == 1 {
			test = anchorRe.ReplaceAllString(test, `<a href="`+prefix+`/$1"`)
			test = imgRe.ReplaceAllString(test, `<img src="`+prefix+`/$1"`)
		}
		newTests[i] = test
	}
	return newTests
}

func newSafeURLOverride(uris []string) func(url []byte) bool {
	return func(url []byte) bool {
		if parser.IsSafeURL(url) {
			return true
		}
		for _, prefix := range uris {
			if bytes.HasPrefix(url, []byte(prefix)) {
				return true
			}
		}
		return false
	}
}

func doTestsReference(t *testing.T, files []string, flag parser.Extensions) {
	params := TestParams{extensions: flag}
	for _, basename := range files {
		t.Run(basename, func(t *testing.T) {
			filename := filepath.Join("testdata", basename+".text")
			inputBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				return
			}
			inputBytes = NormalizeNewlines(inputBytes)
			input := string(inputBytes)

			filename = filepath.Join("testdata", basename+".html")
			expectedBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				return
			}
			expectedBytes = NormalizeNewlines(expectedBytes)
			expected := string(expectedBytes)

			actual := string(runMarkdown(input, params))
			if actual != expected {
				t.Errorf("\n    [%#v]\nExpected[%#v]\nActual  [%#v]",
					basename+".text", expected, actual)
			}
		})
	}
}
