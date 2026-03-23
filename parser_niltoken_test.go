package publiccode

import (
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// TestFindKeyPosNilToken verifies that findKeyPos does not panic when a
// MappingNode key has a nil token (GetToken() == nil).
//
// Before the fix at parser.go:469, the direct dereference
//
//	keyVal := mv.Key.GetToken().Value
//
// would panic with a nil pointer dereference in this case.
func TestFindKeyPosNilToken(t *testing.T) {
	// First entry: NullNode key with Token == nil → GetToken() returns nil.
	nilKeyEntry := &ast.MappingValueNode{
		BaseNode: &ast.BaseNode{},
		Key:      &ast.NullNode{BaseNode: &ast.BaseNode{}},
		Value:    &ast.StringNode{BaseNode: &ast.BaseNode{}, Value: "v"},
	}

	// Second entry: a real StringNode key so the traversal has something to match.
	realTok := &token.Token{Type: token.StringType, Value: "target"}
	realKeyEntry := &ast.MappingValueNode{
		BaseNode: &ast.BaseNode{},
		Key:      &ast.StringNode{BaseNode: &ast.BaseNode{}, Token: realTok, Value: "target"},
		Value:    &ast.StringNode{BaseNode: &ast.BaseNode{}, Value: "v2"},
	}

	node := &ast.MappingNode{
		BaseNode: &ast.BaseNode{},
		Values:   []*ast.MappingValueNode{nilKeyEntry, realKeyEntry},
	}

	// Search for a key that does not exist: exercises the nil-token skip
	// without entering the "found" branch that would need a real Position.
	// Must not panic — that is the primary assertion.
	line, col := findKeyPos(node, []string{"nonexistent"})
	if line != 0 || col != 0 {
		t.Errorf("expected (0,0) for missing key, got (%d,%d)", line, col)
	}
}
