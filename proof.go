package merkle

type Proof struct {
	Proofs []ProofEntry
}

type ProofEntry struct {
	Layer  int
	IsLeft bool
	Hash   Hash
}
