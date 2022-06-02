/// B+ tree implementation on KVStore

package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	store "github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tree is an augmented B+ tree implementation.
// Branches have m sized key index slice. Each key index represents
// the starting index of the child node's index(inclusive), and the
// ending index of the previous node of the child node's index(exclusive).
// TODO: We should abstract out the leaves of this tree to allow more data aside from
// the accumulation value to go there.
type Tree struct {
	store store.KVStore
	m     uint8
}

func NewTree(store store.KVStore, m uint8) Tree {
	tree := Tree{store, m}
	tree.Set(nil, sdk.ZeroInt())
	return tree
}

func (t Tree) Set(key []byte, acc sdk.Int) {
	node := t.nodeGet(0, key)
	node.setLeaf(acc)

	node.parent().push(child{key, acc})
}

func (t Tree) Remove(key []byte) {
	node := t.nodeGet(0, key)
	if !node.exists() {
		return
	}
	parent := node.parent()
	node.delete()
	parent.pull(key)
}

// node is pointer to a specific node inside the tree
type node struct {
	tree  Tree
	level uint16
	key   []byte
	// XXX: cache stored value?
}

// nodeIterator iterates over nodes in a given level. It only iterates directly over the pointers
// to the nodes, not the actual nodes themselves, to save loading additional data into memory.
type nodeIterator struct {
	tree  Tree
	level uint16
	store.Iterator
}

func (iter nodeIterator) node() *node {
	if !iter.Valid() {
		return nil
	}
	res := node{
		tree:  iter.tree,
		level: iter.level,
		key:   iter.Key()[7:],
	}
	return &res
}

type child struct {
	Index []byte
	Acc   sdk.Int
}

type children []child // max length M slice of key bytes, sorted by index

// nodeKey takes in a nodes layer, and its key, and constructs the
// its key in the underlying datastore.
func (t Tree) nodeKey(level uint16, key []byte) []byte {
	bz := make([]byte, 2)
	binary.BigEndian.PutUint16(bz, level)
	return append(append([]byte("node/"), bz...), key...)
}

// leafKey constructs a key for a node pointer representing a leaf node.
func (t Tree) leafKey(key []byte) []byte {
	return t.nodeKey(0, key)
}

func (t Tree) root() *node {
	iter := stypes.KVStoreReversePrefixIterator(t.store, []byte("node/"))
	if !iter.Valid() {
		return nil
	}
	key := iter.Key()[5:]
	return &node{
		tree:  t,
		level: binary.BigEndian.Uint16(key[:2]),
		key:   key[2:],
	}
}

// Get returns the (sdk.Int) accumulation value at a given leaf.
func (t Tree) Get(key []byte) (res sdk.Int) {
	keybz := t.leafKey(key)
	if !t.store.Has(keybz) {
		return sdk.ZeroInt()
	}
	err := json.Unmarshal(t.store.Get(keybz), &res)
	if err != nil {
		panic(err)
	}
	return
}

func (node *node) create(children children) {
	keybz := node.tree.nodeKey(node.level, node.key)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	node.tree.store.Set(keybz, bz)
}

func (t Tree) nodeGet(level uint16, key []byte) *node {
	return &node{
		tree:  t,
		level: level,
		key:   key,
	}
}

// XXX: store.Iterator -> custom node iterator
func (t Tree) nodeIterator(level uint16, begin, end []byte) nodeIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodeKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodeKey(level, nil))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.Iterator(t.nodeKey(level, begin), endBytes),
	}
}

func (t Tree) nodeReverseIterator(level uint16, begin, end []byte) nodeIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodeKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodeKey(level, nil))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.ReverseIterator(t.nodeKey(level, begin), endBytes),
	}
}

func (t Tree) Iterator(begin, end []byte) store.Iterator {
	return t.nodeIterator(0, begin, end)
}

func (t Tree) ReverseIterator(begin, end []byte) store.Iterator {
	return t.nodeReverseIterator(0, begin, end)
}

// accumulationSplit returns the accumulated value for all of the following:
// left: all leaves under nodePointer with key < provided key
// exact: leaf with key = provided key
// right: all leaves under nodePointer with key > provided key
// Note that the equalities here are _exclusive_.
func (node *node) accumulationSplit(key []byte) (left sdk.Int, exact sdk.Int, right sdk.Int) {
	left, exact, right = sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	if node.isLeaf() {
		accumulatedValue := sdk.ZeroInt()
		bz := node.tree.store.Get(node.tree.leafKey(node.key))
		err := json.Unmarshal(bz, &accumulatedValue)
		if err != nil {
			panic(err)
		}
		// Check if the leaf key is to the left of the input key,
		// if so this value is on the left. Similar for the other cases.
		// Recall that all of the output arguments default to 0, if unset internally.
		switch bytes.Compare(node.key, key) {
		case -1:
			left = accumulatedValue
		case 0:
			exact = accumulatedValue
		case 1:
			right = accumulatedValue
		}
		return
	}

	children := node.children()
	idx, match := children.find(key)
	if !match {
		idx--
	}
	left, exact, right = node.tree.nodeGet(node.level-1, children[idx].Index).accumulationSplit(key)
	left = left.Add(children[:idx].accumulate())
	right = right.Add(children[idx+1:].accumulate())
	return left, exact, right
}

// TotalAccumulatedValue returns the sum of the weights for all leaves
func (t Tree) TotalAccumulatedValue() sdk.Int {
	return t.SubsetAccumulation(nil, nil)
}

// Prefix sum returns the total weight of all leaves with keys <= to the provided key.
func (t Tree) PrefixSum(key []byte) sdk.Int {
	return t.SubsetAccumulation(nil, key)
}

// SubsetAccumulation returns the total value of all leaves with keys
// between start and end (inclusive of both ends)
// if start is nil, it is the beginning of the tree.
// if end is nil, it is the end of the tree.
func (t Tree) SubsetAccumulation(start []byte, end []byte) sdk.Int {
	if start == nil {
		left, exact, _ := t.root().accumulationSplit(end)
		return left.Add(exact)
	}
	if end == nil {
		_, exact, right := t.root().accumulationSplit(start)
		return exact.Add(right)
	}
	_, leftexact, leftrest := t.root().accumulationSplit(start)
	_, _, rightest := t.root().accumulationSplit(end)
	return leftexact.Add(leftrest).Sub(rightest)
}

func (t Tree) SplitAcc(key []byte) (sdk.Int, sdk.Int, sdk.Int) {
	return t.root().accumulationSplit(key)
}

func (node *node) visualize(depth int, acc sdk.Int) {
	if !node.exists() {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("- ")
	fmt.Printf("{%d %+v %d}\n", node.level, node.key, acc)
	for i, child := range node.children() {
		childnode := node.child(uint16(i))
		childnode.visualize(depth+1, child.Acc)
	}
}

// DebugVisualize prints the entire tree to stdout
func (t Tree) DebugVisualize() {
	t.root().visualize(0, sdk.ZeroInt())
}
