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
func (t *Tree) Build() {

	// Anything to do?
	// TODO:  Return an error?
	if len(t.Leaves) == 0 {
		return
	}

	// Build the layers of the tree
	layer := t.Leaves[:]
	for len(layer) != 1 {
		layer = buildLayer(layer)
	}

	// Get tree root information
	t.Root = layer[0]
	t.RootHash = t.Root.Hash
}

// Buld a layer of the tree
func buildLayer(layer []*Node) (newLayer []*Node) {
	// Any odd node?
	odd := &Node{}

	// Lets see
	if len(layer)%2 == 1 {
		odd = layer[len(layer)-1]
		layer = layer[:len(layer)-1]
	}

	for i := 0; i <= len(layer)-1; i += 2 {
		nodeHash := CreateNodeHash(layer[i].Hash, layer[i+1].Hash)
		newnode := Node{
			Type: NodeTypeInternal,
			Hash: nodeHash,
		}
		newnode.Left, newnode.Right = layer[i], layer[i+1]
		//layer[i].Left, layer[i+1].Left = true, false
		layer[i].Parent, layer[i+1].Parent = &newnode, &newnode
		newLayer = append(newLayer, &newnode)
	}

	//  Push node up
	if odd.Hash != nil {
		newLayer = append(newLayer, odd)
	}
	return
}

// ToString will create a string representation of the tree
func (t *Tree) ToString() string {
	str := fmt.Sprintf("root: %s\n", base64.StdEncoding.EncodeToString(t.RootHash))
	for _, l := range t.Leaves {
		str = str + fmt.Sprintf("leaf: %s\n", base64.StdEncoding.EncodeToString(l.Hash))
	}
	return str
}
