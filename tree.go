package merkle

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
func (t *Tree) AddContent(ctx context.Context, data [][]byte, extraData [][]byte) error {
	if len(data) != len(extraData) {
		return fmt.Errorf("data and extraDate length must match: data: %v extraData: %v", len(data), len(extraData))
	}

	// Now create leafs and add them to tree
	leaves := make([]*Node, len(data))
	for i := 0; i < len(data); i++ {
		leafHash := CreateLeafHash(data[i])
		leaves[i] = &Node{
			Type:      NodeTypeLeaf,
			Hash:      leafHash,
			ExtraData: extraData[i],
		}
	}

	t.Leaves = leaves
	t.Layers = nil
	return nil
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
	proof := Proof{Proofs: []ProofEntry{}}

	var siblingIndex int
	layer := 0

	// Go from the bottom of the tree up
	for i := t.Depth - 1; i > 0; i-- {

		// Only one item?
		levelLen := len(t.Layers[i])
		if (leafIndex == levelLen-1) && (levelLen%2 == 1) {
			leafIndex = int(leafIndex / 2)
			continue
		}

		// Where can I find the sibling?
		if leafIndex%2 == 0 {
			siblingIndex = leafIndex + 1
		} else {
			siblingIndex = leafIndex - 1
		}

		// Add the layer proof to the collection
		proof.Proofs = append(proof.Proofs, ProofEntry{
			Layer:  layer,
			IsLeft: t.Layers[i][siblingIndex].IsLeft,
			Hash:   t.Layers[i][siblingIndex].Hash,
		})
		layer++
		leafIndex = int(leafIndex / 2)
	}

	return &proof
}

// VerifyProof will verify a passed proof
func (t *Tree) VerifyProof(proof *Proof, root, indexHash Hash) bool {
	proofHash := indexHash

	for _, p := range proof.Proofs {
		if p.IsLeft {
			proofHash = CreateNodeHash(p.Hash, proofHash)
		} else {
			proofHash = CreateNodeHash(proofHash, p.Hash)
		}
	}

	return bytes.Equal(root, proofHash)
}

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

// Serialize will get a transport / storage version of the
func (t *Tree) Serialize() ([]byte, error) {
	return json.Marshal(t.Layers)
}

// DeSerialize will take serialized data and return a tree
func DeSerialize(data []byte) (*Tree, error) {
	var layers [][]*Node
	err := json.Unmarshal(data, &layers)
	if err != nil {
		return nil, err
	}

	return &Tree{Layers: layers, Depth: len(layers)}, nil
}
