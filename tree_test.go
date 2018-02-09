package merkle

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {

	ctx := context.TODO()

	testHasher := &Sha256Hasher{}

	t.Run("ExpectedHashes", func(t *testing.T) {
		// Create somethings
		one := [][]byte{
			[]byte("one"),
		}
		oneH, err := base64.StdEncoding.DecodeString("0Nc2CrefWKseHj/mStd+LqC8B+NrX0btIiPt2SmN+ek=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		two := [][]byte{
			[]byte("one"),
			[]byte("two"),
		}
		twoH, err := base64.StdEncoding.DecodeString("T1X2GdkhUjV3iyufF9b0kVsWFxIU0VI4EpNml2Teci4=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		// Create somethings
		five := [][]byte{
			[]byte("one"),
			[]byte("two"),
			[]byte("three"),
			[]byte("four"),
			[]byte("five"),
		}
		fiveH, err := base64.StdEncoding.DecodeString("gy5gl3aksFyiCO95a/1vLXz88A3dRq+0l9Sxte8ZqZQ=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		type args struct {
			ctx  context.Context
			data [][]byte
		}
		tests := []struct {
			name string
			args args
			want Hash
		}{
			{"one", args{ctx, one}, Hash(oneH)},
			{"two", args{ctx, two}, Hash(twoH)},
			{"five", args{ctx, five}, Hash(fiveH)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				tree := getTestTree(tt.args.ctx, testHasher, tt.args.data)
				if tree == nil {
					t.Error("Unable to create tree")
				}
				fmt.Printf("tree: %s\n", tree.ToString(tt.args.ctx))
				if !bytes.Equal(tree.RootHash, tt.want) {
					t.Errorf("Incorrect hash = %v, want %v", tree.RootHash, tt.want)
				}

			})
		}

	})

}

func getTestTree(ctx context.Context, hasher Hasher, data [][]byte) *Tree {
	tree := NewTree(hasher)
	if tree == nil {
		return nil
	}

	tree.AddContent(ctx, data)

	tree.Build(ctx)

	return tree
}
