// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"math"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// MerkleTree holds the hashes of a merkle tree
type MerkleTree []*daghash.Hash

// Root returns the root of the merkle tree
func (mt MerkleTree) Root() *daghash.Hash {
	return mt[len(mt)-1]
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two. This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

// HashMerkleBranches takes two hashes, treated as the left and right tree
// nodes, and returns the hash of their concatenation. This is a helper
// function used to aid in the generation of a merkle tree.
func HashMerkleBranches(left *daghash.Hash, right *daghash.Hash) *daghash.Hash {
	// Concatenate the left and right nodes.
	var hash [daghash.HashSize * 2]byte
	copy(hash[:daghash.HashSize], left[:])
	copy(hash[daghash.HashSize:], right[:])

	newHash := daghash.DoubleHashH(hash[:])
	return &newHash
}

// BuildHashMerkleTreeStore creates a merkle tree from a slice of transactions, based
// on their hash. See `buildMerkleTreeStore` for more info.
func BuildHashMerkleTreeStore(transactions []*util.Tx) MerkleTree {
	txHashes := make([]*daghash.Hash, len(transactions))
	for i, tx := range transactions {
		txHashes[i] = tx.Hash()
	}
	return buildMerkleTreeStore(txHashes)
}

// BuildIDMerkleTreeStore creates a merkle tree from a slice of transactions, based
// on their ID. See `buildMerkleTreeStore` for more info.
func BuildIDMerkleTreeStore(transactions []*util.Tx) MerkleTree {
	txIDs := make([]*daghash.Hash, len(transactions))
	for i, tx := range transactions {
		txIDs[i] = (*daghash.Hash)(tx.ID())
	}
	return buildMerkleTreeStore(txIDs)
}

// buildMerkleTreeStore creates a merkle tree from a slice of hashes,
// stores it using a linear array, and returns a slice of the backing array. A
// linear array was chosen as opposed to an actual tree structure since it uses
// about half as much memory. The following describes a merkle tree and how it
// is stored in a linear array.
//
// A merkle tree is a tree in which every non-leaf node is the hash of its
// children nodes. A diagram depicting how this works for kaspa transactions
// where h(x) is a double sha256 follows:
//
//	         root = h1234 = h(h12 + h34)
//	        /                           \
//	  h12 = h(h1 + h2)            h34 = h(h3 + h4)
//	   /            \              /            \
//	h1 = h(tx1)  h2 = h(tx2)    h3 = h(tx3)  h4 = h(tx4)
//
// The above stored as a linear array is as follows:
//
// 	[h1 h2 h3 h4 h12 h34 root]
//
// As the above shows, the merkle root is always the last element in the array.
//
// The number of inputs is not always a power of two which results in a
// balanced tree structure as above. In that case, parent nodes with no
// children are also zero and parent nodes with only a single left node
// are calculated by concatenating the left node with itself before hashing.
// Since this function uses nodes that are pointers to the hashes, empty nodes
// will be nil.
func buildMerkleTreeStore(hashes []*daghash.Hash) MerkleTree {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(hashes))
	arraySize := nextPoT*2 - 1
	merkles := make(MerkleTree, arraySize)

	// Create the base transaction hashes and populate the array with them.
	for i, hash := range hashes {
		merkles[i] = hash
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := HashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

		// The normal case sets the parent node to the double sha256
		// of the concatentation of the left and right children.
		default:
			newHash := HashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}
