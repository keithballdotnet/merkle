package merkle

import (
	"context"
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {

	ctx := context.TODO()

	testHasher := &Sha256Hasher{}

	// Create somethings
	things := [][]byte{
		[]byte("never"),
		[]byte("be"),
		[]byte("the"),
		[]byte("same"),
	}

	tree1 := getTestTree(ctx, testHasher, things)

	// Create somethings
	morethings := [][]byte{
		[]byte("never"),
		[]byte("be"),
		[]byte("the"),
		[]byte("same"),
		[]byte("again"),
	}

	tree2 := getTestTree(ctx, testHasher, morethings)

	fmt.Printf("Tree: %s\n", tree1.ToString())
	fmt.Printf("Tree: %s\n", tree2.ToString())

}

func getTestTree(ctx context.Context, hasher Hasher, data [][]byte) *Tree {
	tree := NewTree(hasher)
	if tree == nil {
		return nil
	}

	tree.AddContent(ctx, data)

	tree.Build(ctx)

	return tree
}
