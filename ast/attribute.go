package ast

import (
	"fmt"
	"sort"
)

// BlockAttribute can be attached to block elements. They are specified as
// {#id .classs key="value"} where quotes for value are mandatory, multiple
// key/value pairs are separated by whitespace.
// Any trailing whitespace is trimmed from both elements.
// If multiple attributes are specified, they are cumulative, new classes
// are added, as are new attr. The singleton ID (if given again) is overwritten.
// BlockAttribute must be given on a line by itself.
type BlockAttribute struct {
	Container

	ID      string            // single #id
	Classes []string          // zero or more .class
	Attrs   map[string]string // zero or more key/value pairs

	Used bool // This BlockAttribute has been used, don't reuse it anymore.
}

//Collapse adds BlockAttr j to b. It overwrites the #id, added classes and collapsing attributes.
func (b *BlockAttribute) Collapse(j *BlockAttribute) *BlockAttribute {
	if j.Used {
		return b
	}

	if j.ID != "" {
		b.ID = j.ID
	}

	b.Classes = append(b.Classes, j.Classes...)

	for k, a := range j.Attrs {
		b.Attrs[k] = a
	}

	return b
}

// String returns the string representation of b. TODO(miek): should be renderer specific!
func (b *BlockAttribute) String() string {
	if b.Used {
		return ""
	}

	var s string
	if b.ID != "" {
		s = "#" + b.ID
	}

	sort.Strings(b.Classes)
	for _, c := range b.Classes {
		s += " " + c
	}

	var keys []string
	for k := range b.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for k, v := range keys {
		s += fmt.Sprintf("%s=\"%s\" ", k, v)
	}

	return s
}
