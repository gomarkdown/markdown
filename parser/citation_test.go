package parser

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestCitation(t *testing.T) {
	data := []byte(`[@!RFC1035]`)

	p := New()
	p.extensions |= Mmark

	_, node := citation(p, data, 0)
	dest := string(node.(*ast.Citation).Destination[0])
	if dest != "RFC1035" {
		t.Errorf("failed to find citation, want %s, got %s", "RFC1035", dest)
	}
	tp := node.(*ast.Citation).Type[0]
	if tp != ast.CitationTypeNormative {
		t.Errorf("failed to find citation type, want %d, got %d", ast.CitationTypeNormative, tp)
	}
}

func TestCitationSuffix(t *testing.T) {
	data := []byte(`[@!RFC1035, p. 144]`)

	p := New()
	p.extensions |= Mmark

	_, node := citation(p, data, 0)
	if dest := string(node.(*ast.Citation).Destination[0]); dest != "RFC1035" {
		t.Errorf("failed to find citation, want %s, got %s", "RFC1035", dest)
	}
	tp := node.(*ast.Citation).Type[0]
	if tp != ast.CitationTypeNormative {
		t.Errorf("failed to find citation type, want %d, got %d", ast.CitationTypeNormative, tp)
	}
	suff := string(node.(*ast.Citation).Suffix[0])
	if suff != "p. 144" {
		t.Errorf("failed to find citation suffix, want %s, got %s", "p. 144", suff)
	}
}

func TestCitationSuffixMultiple(t *testing.T) {
	data := []byte(`[@?RFC1034; @!RFC1035, p. 144, more]`)

	p := New()
	p.extensions |= Mmark

	_, node := citation(p, data, 0)
	if dest := string(node.(*ast.Citation).Destination[0]); dest != "RFC1034" {
		t.Errorf("failed to find citation, want %s, got %s", "RFC1034", dest)
	}
	tp := node.(*ast.Citation).Type[0]
	if tp != ast.CitationTypeInformative {
		t.Errorf("failed to find citation type, want %d, got %d", ast.CitationTypeInformative, tp)
	}
	if dest := string(node.(*ast.Citation).Destination[1]); dest != "RFC1035" {
		t.Errorf("failed to find citation, want %s, got %s", "RFC1035", dest)
	}
	suff := string(node.(*ast.Citation).Suffix[1])
	if suff != "p. 144, more" {
		t.Errorf("failed to find citation suffix, want %s, got %s", "p. 144, more", suff)
	}
}
