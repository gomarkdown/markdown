package query

import (
	"testing"

	"github.com/moorara/markdown/ast"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name         string
		node         *ast.Node
		query        string
		expectedNode *ast.Node
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			node := Find(tc.node, tc.query)

			// TODO: implement comparison logic
			if node != tc.expectedNode {
				t.Fatalf("Failed!")
			}
		})
	}
}
