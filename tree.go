package merkle

import (
	"context"
	"encoding/base64"
	"fmt"
)

var hasher Hasher

// Tree ...
type Tree struct {
	RootHash Hash
	Root     *Node
	Leaves   []*Node
	Height   int
}

// NewTree will create a new Merkle tree
func NewTree(h Hasher) *Tree {
	hasher = h
	return &Tree{}
}

// AddContent will add data to the tree
func (t *Tree) AddContent(ctx context.Context, data [][]byte) {
	// Now create leafs and add them to tree
	var leaves []*Node
	for _, d := range data {
		leafHash := CreateLeafHash(d)
		leaves = append(leaves, &Node{
			Type: NodeTypeLeaf,
			Hash: leafHash,
		})
	}
	t.Leaves = leaves
}

// Build the merkle tree once we have added content
func (t *Tree) Build(ctx context.Context) {

	// Anything to do?
	// TODO:  Return an error?
	if len(t.Leaves) == 0 {
		return
	}

	// Build the layers of the tree
	layer := t.Leaves[:]
	height := 0
	for len(layer) != 1 {
		layer = buildLayer(ctx, layer, &height)
	}

	// Get tree root information
	t.Height = height
	t.Root = layer[0]
	t.RootHash = t.Root.Hash
}

// Buld a layer of the tree
func buildLayer(ctx context.Context, layer []*Node, height *int) []*Node {
	var newLayer []*Node
	*height++

	// Separate any odd node off from the collection
	odd := &Node{}
	if len(layer)%2 == 1 {
		odd = layer[len(layer)-1]
		layer = layer[:len(layer)-1]
	}

	// Loop through the layer
	for i := 0; i <= len(layer)-1; i += 2 {
		nodeHash := CreateNodeHash(layer[i].Hash, layer[i+1].Hash)
		newnode := Node{
			Type: NodeTypeInternal,
			Hash: nodeHash,
		}

		// Set up the nodes relationships
		newnode.Left = layer[i]
		newnode.Right = layer[i+1]
		layer[i].Parent = &newnode
		layer[i+1].Parent = &newnode

		// Add to the new layer
		newLayer = append(newLayer, &newnode)
	}

	// The odd nodes will be pushed upwards
	if odd.Hash != nil {
		newLayer = append(newLayer, odd)
	}

	return newLayer
}

// ToString will create a string representation of the tree
func (t *Tree) ToString() string {
	str := fmt.Sprintf("\nroot: %s height: %v\n", base64.StdEncoding.EncodeToString(t.RootHash), t.Height)
	for _, l := range t.Leaves {
		str = str + fmt.Sprintf("leaf: %s\n", base64.StdEncoding.EncodeToString(l.Hash))
		parent := l.Parent
		indent := "  "
		for parent != nil {
			str = str + fmt.Sprintf("%sParent: %s\n", indent, base64.StdEncoding.EncodeToString(parent.Hash))
			parent = parent.Parent
			indent += indent
		}
	}

	return str
}
