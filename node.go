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
	Left   *Node
	Right  *Node
	Parent *Node
}

// IsLeaf retuns true if this node is a leaf
func (n *Node) IsLeaf() bool {
	return (n.Left == nil && n.Right == nil)
}
