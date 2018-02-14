package merkle

// NodeType ...
type NodeType int

// Some NodeType constants
const (
	NodeTypeInternal NodeType = 0
	NodeTypeLeaf     NodeType = 1
)

// Node is an element in the merkle tree
type Node struct {
	Type   NodeType
	Hash   Hash
	IsLeft bool
	Parent *Node
}
