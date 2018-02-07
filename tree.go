package merkle

var hasher Hasher

// Tree ...
type Tree struct {
	RootHash Hash
	Leaves   []*Node
}

// NewTree will create a new Merkle tree
func NewTree(h Hasher) *Tree {
	hasher = h
	return &Tree{}
}

// AddHashes will add the content hashes as leave to the tree
func (t *Tree) AddHashes(hashes []Hash) {
	var leaves []*Node
	for _, hash := range hashes {
		leaves = append(leaves, &Node{
			Type: NodeTypeLeaf,
			Hash: hash,
		})
	}
}
