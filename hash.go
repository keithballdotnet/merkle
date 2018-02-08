package merkle

import (
	"bytes"
	"crypto/sha256"
)

// Hash is a hash
type Hash []byte

// HashPrefix is prefix to the hash being done
type HashPrefix []byte

// HashPrefix types
var (
	PrefixLeaf HashPrefix = []byte{0}
	PrefixNode HashPrefix = []byte{1}
)

// Concatenate will join and hash two hashes
func CreateLeafHash(data []byte) Hash {
	// Concat prefix and hashes
	concat := bytes.Join([][]byte{PrefixLeaf, data}, []byte{})

	return hasher.Hash(concat)
}

// CreateNodeHash will join and hash two hashes
func CreateNodeHash(left Hash, right Hash) Hash {
	// Concat prefix and hashes
	concat := bytes.Join([][]byte{PrefixNode, left, right}, []byte{})

	return hasher.Hash(concat)
}

// Hasher interface
type Hasher interface {
	Hash(data []byte) Hash
}

// Sha256Hasher fullfills the hasher interface
type Sha256Hasher struct{}

// Hash will hash data using the sha256 algorithm
func (h *Sha256Hasher) Hash(data []byte) Hash {
	hash := sha256.Sum256(data)
	return Hash(hash[:])
}
