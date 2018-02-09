package merkle

import (
	"context"
	"encoding/base64"
	"fmt"
)

// ProofEntry ...
type ProofEntry struct {
	Layer  int
	IsLeft bool
	Hash   Hash
}

// Proof is a collecion of ProofEntrys
type Proof struct {
	Proofs []ProofEntry
}

// ToString gives a string representation of the proofs
func (p *Proof) ToString(ctx context.Context) string {
	str := ""
	for _, pe := range p.Proofs {
		str += fmt.Sprintf("Layer: %v IsLeft: %v Hash: %s\n", pe.Layer, pe.IsLeft, base64.StdEncoding.EncodeToString(pe.Hash))
	}
	return str
}
