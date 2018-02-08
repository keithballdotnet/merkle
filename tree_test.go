package merkle

import (
	"context"
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {

	ctx := context.TODO()

	tree := NewTree(&Sha256Hasher{})
	if tree == nil {
		t.Fail()
	}

	// Create somethings
	things := [][]byte{
		[]byte("never"),
		[]byte("be"),
		[]byte("the"),
		[]byte("same"),
	}

	tree.AddContent(ctx, things)

	fmt.Printf("Tree: %s\n", tree.ToString())

	tree.Build()

	fmt.Printf("Tree: %s\n", tree.ToString())

	tree2 := NewTree(&Sha256Hasher{})
	if tree2 == nil {
		t.Fail()
	}

	// Create somethings
	things = [][]byte{
		[]byte("never"),
		[]byte("be"),
		[]byte("the"),
		[]byte("same"),
		[]byte("again"),
	}

	tree2.AddContent(ctx, things)

	fmt.Printf("Tree: %s\n", tree2.ToString())

	tree2.Build()

	fmt.Printf("Tree: %s\n", tree2.ToString())
}
