package merkle

import (
	"bytes"
	"crypto/sha256"
)

// Hash is a hash
type Hash []byte

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

// Concatenate will join and hash two hashes
func (h Hash) Concatenate(other Hash) Hash {

	concat := bytes.Join([][]byte{[]byte(h), []byte(other)}, []byte{})

	return hasher.Hash(concat)
}
