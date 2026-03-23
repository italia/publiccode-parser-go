package publiccode

import (
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// TestFindKeyAtLineNilTokenMappingValue verifies that findKeyAtLine does not
// panic when a MappingValueNode key has a nil token (GetToken() == nil).
//
// Without nil checks, parser.go:566-576 dereferences the token directly:
//
//	keyTok := n.Key.GetToken()
//	fullKey = keyTok.Value          // panic if keyTok is nil
//	keyTok.Position.Line == target  // panic if keyTok is nil
func TestFindKeyAtLineNilTokenMappingValue(t *testing.T) {
	// MappingValueNode whose key has no token — GetToken() returns nil.
	nilTokEntry := &ast.MappingValueNode{
		BaseNode: &ast.BaseNode{},
		Key:      &ast.NullNode{BaseNode: &ast.BaseNode{}},
		Value:    &ast.StringNode{BaseNode: &ast.BaseNode{}, Value: "v"},
	}

	// Wrap in a MappingNode so findKeyAtLine dispatches through the
	// MappingValueNode case.
	node := &ast.MappingNode{
		BaseNode: &ast.BaseNode{},
		Values:   []*ast.MappingValueNode{nilTokEntry},
	}

	// Must not panic regardless of target line.
	result := findKeyAtLine(node, 1, "")
	if result != "" {
		t.Errorf("expected empty result for nil-token key, got %q", result)
	}

	// Also call directly on the MappingValueNode.
	result = findKeyAtLine(nilTokEntry, 1, "prefix")
	if result != "" {
		t.Errorf("expected empty result for nil-token key, got %q", result)
	}
}

// TestFindKeyAtLineNilTokenMappingValueWithRealKey checks that a nil-token
// entry does not prevent traversal from reaching a subsequent valid entry.
func TestFindKeyAtLineNilTokenMappingValueWithRealKey(t *testing.T) {
	nilTokEntry := &ast.MappingValueNode{
		BaseNode: &ast.BaseNode{},
		Key:      &ast.NullNode{BaseNode: &ast.BaseNode{}},
		Value:    &ast.StringNode{BaseNode: &ast.BaseNode{}, Value: "v"},
	}

	realTok := &token.Token{
		Type:  token.StringType,
		Value: "realkey",
		Position: &token.Position{
			Line:   5,
			Column: 1,
		},
	}
	realEntry := &ast.MappingValueNode{
		BaseNode: &ast.BaseNode{},
		Key:      &ast.StringNode{BaseNode: &ast.BaseNode{}, Token: realTok, Value: "realkey"},
		Value:    &ast.StringNode{BaseNode: &ast.BaseNode{}, Value: "v2"},
	}

	node := &ast.MappingNode{
		BaseNode: &ast.BaseNode{},
		Values:   []*ast.MappingValueNode{nilTokEntry, realEntry},
	}

	// Must not panic, and must find the real key at line 5.
	result := findKeyAtLine(node, 5, "")
	if result != "realkey" {
		t.Errorf("expected 'realkey', got %q", result)
	}
}
