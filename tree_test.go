package merkle

import "testing"

func TestTree(t *testing.T) {
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

	// Now create leafs and add them to tree
	var hashes []Hash
	for _, things := range things {
		thingHash := hasher.Hash(things)
		hashes = append(hashes, thingHash)
	}

	tree.AddHashes(hashes)

}
