package markdown

import (
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

func execRecoverableTestSuite(t *testing.T, tests []string, params TestParams, suite func(candidate *string)) {
	// Catch and report panics. This is useful when running 'go test -v' on
	// the integration server. When developing, though, crash dump is often
	// preferable, so recovery can be easily turned off with doRecover = false.
	var candidate string
	const doRecover = false
	if doRecover {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("\npanic while processing [%#v]: %s\n", candidate, err)
			}
		}()
	}
	suite(&candidate)
}

func runMarkdown(input string, params TestParams) string {
	params.RendererOptions.Flags = params.Flags
	parser := parser.NewWithExtensions(params.extensions)
	parser.ReferenceOverride = params.referenceOverride
	renderer := html.NewRenderer(params.RendererOptions)

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

func doTestsBlock(t *testing.T, tests []string, extensions parser.Extensions) {
	doTestsParam(t, tests, TestParams{
		extensions: extensions,
		Flags:      html.UseXHTML,
	})
}

func doTestsParam(t *testing.T, tests []string, params TestParams) {
	execRecoverableTestSuite(t, tests, params, func(candidate *string) {
		for i := 0; i+1 < len(tests); i += 2 {
			input := tests[i]
			*candidate = input
			expected := tests[i+1]
			actual := runMarkdown(*candidate, params)
			if actual != expected {
				t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
					*candidate, expected, actual)
			}
		}
	})
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

func doTestsReference(t *testing.T, files []string, flag parser.Extensions) {
	params := TestParams{extensions: flag}
	execRecoverableTestSuite(t, files, params, func(candidate *string) {
		for _, basename := range files {
			filename := filepath.Join("testdata", basename+".text")
			inputBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				continue
			}
			input := string(inputBytes)

			filename = filepath.Join("testdata", basename+".html")
			expectedBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Errorf("Couldn't open '%s', error: %v\n", filename, err)
				continue
			}
			expected := string(expectedBytes)

			actual := string(runMarkdown(input, params))
			if actual != expected {
				t.Errorf("\n    [%#v]\nExpected[%#v]\nActual  [%#v]",
					basename+".text", expected, actual)
			}
		}
	})
}
