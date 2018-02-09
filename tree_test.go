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
		one := getTestData(1)
		oneH, err := base64.StdEncoding.DecodeString("lqKW0iTyhcZ77pPDD4owkVfw2qNdxbh+QQt4YwoJz8c=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		two := getTestData(2)
		twoH, err := base64.StdEncoding.DecodeString("ogv5p8wtyKCPX0FacbGfasQnurVNJO7IaLXTEDRJlTo=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		five := getTestData(5)
		fiveH, err := base64.StdEncoding.DecodeString("uFW0LWww9bCH4FJmeD+9bjlPe5JgE8yqZ3AKiwxaWW8=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		eight := getTestData(8)
		eightH, err := base64.StdEncoding.DecodeString("739JtiD2x+qbljohTaNLUCHG3tjtV3NDgKMRq3JqqQc=")
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		// twoHundred := getTestData(200)
		// twoHundredH, err := base64.StdEncoding.DecodeString("4vB6nPZQJrooLLKG1nO1ZHh6ZHhVda0aqnRGjaAtYWE=")
		// if err != nil {
		// 	t.Errorf("Error: %v", err)
		// }

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
			{"eight", args{ctx, eight}, Hash(eightH)},
			//{"twoHundred", args{ctx, twoHundred}, Hash(twoHundredH)},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				tree := getTestTree(tt.args.ctx, testHasher, tt.args.data)
				if tree == nil {
					t.Error("Unable to create tree")
				}
				fmt.Printf("tree: %s\n", tree.ToString(tt.args.ctx))
				if !bytes.Equal(tree.GetRootHash(), tt.want) {
					t.Errorf("Incorrect hash = %v, want %v", tree.GetRootHash(), tt.want)
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

func getTestData(len int) [][]byte {
	testData := make([][]byte, len)
	for i := 0; i < len; i++ {
		testData[i] = []byte{byte(i)}
	}
	return testData
}
