package ast

// BlockAttribute can be attached to block elements. They are specified as
// {#id .classs key="value"} where quotes for value are mandatory, multiple
// key/value pairs are separated by whitespace.
// Any trailing whitespace is trimmed from both elements.
// If multiple attributes are specified, they are cumulative, new classes
// are added, as are new attr. The singleton ID (if given again) is overwritten.
// BlockAttribute must be given on a line by itself.
type BlockAttribute struct {
	Container

	ID      []byte            // single #id
	Classes [][]byte          // zero or more .class
	Attrs   map[string][]byte // zero or more key/value pairs

	Used bool // This BlockAttribute has been used, don't reuse it anymore.
}

//Collapse adds BlockAttr j to b. It overwrites the #id, added classes and collapsing attributes.
func (b *BlockAttribute) Collapse(j *BlockAttribute) *BlockAttribute {
	if j.Used {
		return b
	}

	if len(j.ID) != 0 {
		b.ID = j.ID
	}

	b.Classes = append(b.Classes, j.Classes...)

	for k, a := range j.Attrs {
		b.Attrs[k] = a
	}

	return b
}
