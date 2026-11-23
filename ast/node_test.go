package ast

import (
	"reflect"
	"testing"
)

func TestContainerCanHaveChildren(t *testing.T) {
	parent := &Container{}
	child := &Leaf{}

	if len(parent.GetChildren()) != 0 {
		t.Error("Parent did not start out without children")
	}

	newChildren := []Node{child}
	parent.SetChildren(newChildren)
	child.Parent = parent

	if !reflect.DeepEqual(parent.GetChildren(), newChildren) {
		t.Error("Failed to set children")
	}
}

func TestLeafCannotHaveChildren(t *testing.T) {
	parent := &Leaf{}
	child := &Leaf{}

	newChildren := []Node{child}

	// Expect that SetChildren panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("Leaf.SetChildren did not panic but was expected to")
		}
	}()

	parent.SetChildren(newChildren)
	child.Parent = parent
}

func TestLeafCanSetEmptyChildren(t *testing.T) {
	parent := &Leaf{}

	parent.SetChildren(nil)
	parent.SetChildren([]Node{})
}

func TestRemoveLeaveFromTree(t *testing.T) {
	// Create a tree to remove nodes from:
	/*
	      grandparent
	           |
	         parent
	        /      \
	 toBeRemoved  sibling
	*/

	grandparent := &Container{}
	parent := &Container{}
	toBeRemoved := &Leaf{}
	sibling := &Leaf{}

	grandparent.SetChildren([]Node{parent})
	parent.Parent = grandparent

	parent.SetChildren(([]Node{toBeRemoved, sibling}))
	toBeRemoved.Parent = parent
	sibling.Parent = parent

	RemoveFromTree(toBeRemoved)

	if !reflect.DeepEqual(grandparent.GetChildren(), []Node{parent}) {
		t.Error("Unexpectedly modified children of grandparent when removing grandchild")
	}

	if !reflect.DeepEqual(parent.GetChildren(), []Node{sibling}) {
		t.Errorf("Unexpected modification of removed node's siblings: %v", parent.GetChildren())
	}

	// The parent reference of the removed node is left intact
	if toBeRemoved.Parent != parent {
		t.Errorf("Unexpectedly modified parent of removed node to: %v", toBeRemoved.Parent)
	}
}

func TestRemoveContainerKeepsItsChildren(t *testing.T) {
	doc := &Document{}
	para := &Paragraph{}
	text := &Text{Leaf: Leaf{Literal: []byte("hi")}}
	AppendChild(para, text)
	AppendChild(doc, para)

	other := &Document{}
	AppendChild(other, para)

	if got := para.GetChildren(); len(got) != 1 || got[0] != Node(text) {
		t.Fatalf("moved paragraph children = %v, want the text node", got)
	}
	if text.GetParent() != para {
		t.Fatal("text node lost its parent when the paragraph was moved")
	}
	if len(doc.GetChildren()) != 0 {
		t.Fatal("paragraph was still a child of the document it was moved from")
	}
	if len(other.GetChildren()) != 1 || other.GetChildren()[0] != Node(para) {
		t.Fatal("paragraph was not attached to its new parent")
	}
}

func TestSiblingLinks(t *testing.T) {
	parent := &Document{}
	first := &Paragraph{}
	second := &Paragraph{}
	third := &Paragraph{}

	AppendChild(parent, first)
	AppendChild(parent, second)
	AppendChild(parent, third)

	if got := GetPrevNode(second); got != first {
		t.Fatalf("GetPrevNode(second) = %v, want first", got)
	}
	if got := GetNextNode(second); got != third {
		t.Fatalf("GetNextNode(second) = %v, want third", got)
	}

	RemoveFromTree(second)

	if got := GetNextNode(first); got != third {
		t.Fatalf("GetNextNode(first) after remove = %v, want third", got)
	}
	if got := GetPrevNode(third); got != first {
		t.Fatalf("GetPrevNode(third) after remove = %v, want first", got)
	}
	if got := GetPrevNode(second); got != nil {
		t.Fatalf("GetPrevNode(removed) = %v, want nil", got)
	}
	if got := GetNextNode(second); got != nil {
		t.Fatalf("GetNextNode(removed) = %v, want nil", got)
	}
}
