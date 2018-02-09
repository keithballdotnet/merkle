package merkle

import (
	"context"
	"encoding/base64"
	"fmt"
)

var hasher Hasher

// Tree ...
type Tree struct {
	Leaves []*Node
	// Layers is populated after build
	Layers [][]*Node
	Depth  int
}

// NewTree will create a new Merkle tree
func NewTree(h Hasher) *Tree {
	hasher = h
	return &Tree{}
}

// GetRoot will return the root node
func (t *Tree) GetRoot() *Node {
	if len(t.Layers) == 0 {
		return nil
	}
	return t.Layers[0][0]
}

// GetRootHash will return the root hash
func (t *Tree) GetRootHash() Hash {
	if len(t.Layers) == 0 {
		return nil
	}
	return t.Layers[0][0].Hash
}

// AddContent will add data to the tree (replaces all leaves)
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
	t.Layers = nil
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
	depth := 1
	layers := [][]*Node{}
	layers = append(layers, layer)
	for len(layer) != 1 {
		layer = buildLayer(ctx, layer)
		layers = append(layers, layer)
		depth++
	}

	// Reverse layers so root is at index 0 0
	for i, j := 0, len(layers)-1; i < j; i, j = i+1, j-1 {
		layers[i], layers[j] = layers[j], layers[i]
	}

	// Set tree
	t.Depth = depth
	t.Layers = layers
}

// Buld a layer of the tree
func buildLayer(ctx context.Context, layer []*Node) []*Node {
	var newLayer []*Node

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
		layer[i].IsLeft = true
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

// GetProof will return a collection of hashes that can be used to prove some data is in the tree
func (t *Tree) GetProof(ctx context.Context, leafIndex int) *Proof {

	return nil
}

// // Verify data will indicate if data is present in the tree and also recalulate that the data
// func (t *Tree) VerifyData(expectedRoot Hash, data []byte) bool {
// 	expectedHash := CreateLeafHash(data)
// 	for _, leaf := range t.Leaves {
// 		// Found data in leaves
// 		if bytes.Equal(expectedHash, leaf.Hash) {

// 		}

// 	}
// 	return false
// }

// ToString will create a string representation of the tree
func (t *Tree) ToString(ctx context.Context) string {
	str := fmt.Sprintf("\nroot: %s depth: %v\n", base64.StdEncoding.EncodeToString(t.GetRootHash()), t.Depth)
	for i := 0; i < t.Depth; i++ {
		str += fmt.Sprintf("Depth: %v", i)
		for _, l := range t.Layers[i] {
			str = str + fmt.Sprintf(" - H: %s T: %v L: %v", base64.StdEncoding.EncodeToString(l.Hash), l.Type, l.IsLeft)
		}
		str += "\n"
	}

	return str
}
