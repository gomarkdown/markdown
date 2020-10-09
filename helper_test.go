package markdown

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/moorara/markdown/parser"
	"github.com/moorara/markdown/render/html"
)

type TestParams struct {
	extensions        parser.Extensions
	referenceOverride parser.ReferenceOverrideFunc
	html.Flags
	html.RendererOptions
}

func runMarkdown(input string, params TestParams) string {
	params.RendererOptions.Flags = params.Flags
	parser := parser.NewWithExtensions(params.extensions)
	parser.ReferenceOverride = params.referenceOverride
	renderer := html.NewRenderer(params.RendererOptions)

	d := ToHTML([]byte(input), parser, renderer)
	return string(d)
}

// runTests runs full document tests using MarkdownCommon configuration.
func runTests(t *testing.T, tests []string) {
	runParamTests(t, tests, TestParams{
		extensions: parser.CommonExtensions,
		RendererOptions: html.RendererOptions{
			Flags: html.CommonFlags,
		},
	})
}

func runBlockTests(t *testing.T, tests []string, extensions parser.Extensions) {
	runParamTests(t, tests, TestParams{
		extensions: extensions,
		Flags:      html.UseXHTML,
	})
}

func runParamTests(t *testing.T, tests []string, params TestParams) {
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

func runInlineTests(t *testing.T, tests []string) {
	runInlineParamTests(t, tests, TestParams{})
}

func runInlineLinkTests(t *testing.T, tests []string) {
	runInlineTests(t, tests)

	prefix := "http://localhost"
	params := html.RendererOptions{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	runInlineParamTests(t, transformTests, TestParams{
		RendererOptions: params,
	})
	runInlineParamTests(t, transformTests, TestParams{
		Flags:           html.UseXHTML,
		RendererOptions: params,
	})
}

func runInlineSafeTests(t *testing.T, tests []string) {
	runInlineParamTests(t, tests, TestParams{Flags: html.Safelink})

	// All the links in this test should not have the prefix appended, so
	// just rerun it with different parameters and the same expectations.
	prefix := "http://localhost"
	params := html.RendererOptions{AbsolutePrefix: prefix}
	transformTests := transformLinks(tests, prefix)
	runInlineParamTests(t, transformTests, TestParams{
		Flags:           html.Safelink,
		RendererOptions: params,
	})
}

func runInlineParamTests(t *testing.T, tests []string, params TestParams) {
	params.extensions |= parser.Autolink | parser.Strikethrough
	params.Flags |= html.UseXHTML
	runParamTests(t, tests, params)
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

func runReferenceTests(t *testing.T, files []string, flag parser.Extensions) {
	params := TestParams{extensions: flag}
	for _, basename := range files {
		filename := filepath.Join("testdata", basename+".text")
		inputBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
			continue
		}
		inputBytes = normalizeNewlines(inputBytes)
		input := string(inputBytes)

		filename = filepath.Join("testdata", basename+".html")
		expectedBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
			continue
		}
		expectedBytes = normalizeNewlines(expectedBytes)
		expected := string(expectedBytes)

		actual := string(runMarkdown(input, params))
		if actual != expected {
			t.Errorf("\n    [%#v]\nExpected[%#v]\nActual  [%#v]",
				basename+".text", expected, actual)
		}
	}
}

func normalizeNewlines(d []byte) []byte {
	// replace CR LF (windows) with LF (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF (mac) with LF (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}
