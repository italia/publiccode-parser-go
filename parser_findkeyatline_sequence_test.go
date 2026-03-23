package publiccode

import (
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// TestFindKeyAtLineNilTokenSequence verifies that findKeyAtLine does not panic
// when a SequenceNode entry has a nil token (GetToken() == nil).
//
// Without a nil check, parser.go:587 dereferences the token directly:
//
//	if entry.GetToken().Position.Line == targetLine {
//
// A nil return from GetToken() causes a nil pointer dereference.
func TestFindKeyAtLineNilTokenSequence(t *testing.T) {
	// SequenceNode with a nil-token entry as the first element.
	nilTokEntry := &ast.NullNode{BaseNode: &ast.BaseNode{}} // GetToken() returns nil

	node := &ast.SequenceNode{
		BaseNode: &ast.BaseNode{},
		Values:   []ast.Node{nilTokEntry},
	}

	// Must not panic regardless of target line.
	result := findKeyAtLine(node, 1, "items")
	if result != "" {
		t.Errorf("expected empty result for nil-token entry, got %q", result)
	}
}

// TestFindKeyAtLineNilTokenSequenceWithRealEntry checks that a nil-token entry
// does not prevent traversal from reaching a subsequent valid entry.
func TestFindKeyAtLineNilTokenSequenceWithRealEntry(t *testing.T) {
	nilTokEntry := &ast.NullNode{BaseNode: &ast.BaseNode{}} // GetToken() returns nil

	realTok := &token.Token{
		Type:  token.StringType,
		Value: "hello",
		Position: &token.Position{
			Line:   7,
			Column: 3,
		},
	}
	realEntry := &ast.StringNode{
		BaseNode: &ast.BaseNode{},
		Token:    realTok,
		Value:    "hello",
	}

	node := &ast.SequenceNode{
		BaseNode: &ast.BaseNode{},
		Values:   []ast.Node{nilTokEntry, realEntry},
	}

	// Must not panic, and must find the real entry at line 7.
	result := findKeyAtLine(node, 7, "items")
	if result != "items[1]" {
		t.Errorf("expected 'items[1]', got %q", result)
	}
}
