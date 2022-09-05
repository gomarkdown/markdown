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
