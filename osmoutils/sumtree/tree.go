/// B+ tree implementation on KVStore

package sumtree

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gogo/protobuf/proto"

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
	if tree.IsEmpty() {
		tree.Set(nil, sdk.ZeroInt())
	}
	return tree
}

func (t Tree) IsEmpty() bool {
	return !t.store.Has(t.leafKey(nil))
}

func (t Tree) Set(key []byte, acc sdk.Int) {
	ptr := t.ptrGet(0, key)
	leaf := NewLeaf(key, acc)
	ptr.setLeaf(leaf)

	ptr.parent().push(leaf.Leaf)
}

func (t Tree) Remove(key []byte) {
	node := t.ptrGet(0, key)
	if !node.exists() {
		return
	}
	parent := node.parent()
	node.delete()
	parent.pull(key)
}

func (t Tree) Increase(key []byte, amt sdk.Int) {
	value := t.Get(key)
	t.Set(key, value.Add(amt))
}

func (t Tree) Decrease(key []byte, amt sdk.Int) {
	t.Increase(key, amt.Neg())
}

func (t Tree) Clear() {
	iter := t.store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		t.store.Delete(iter.Key())
	}
}

// ptr is pointer to a specific node inside the tree.
type ptr struct {
	tree  Tree
	level uint16
	key   []byte
	// XXX: cache stored value?
}

// ptrIterator iterates over ptrs in a given level. It only iterates directly over the pointers
// to the nodes, not the actual nodes themselves, to save loading additional data into memory.
type ptrIterator struct {
	tree  Tree
	level uint16
	store.Iterator
}

func (iter ptrIterator) ptr() *ptr {
	if !iter.Valid() {
		return nil
	}
	res := ptr{
		tree:  iter.tree,
		level: iter.level,
		key:   iter.Key()[7:],
	}
	// ptrIterator becomes invalid once retrieve ptr
	err := iter.Close()
	if err != nil {
		panic(err)
	}
	return &res
}

// nodeKey takes in a nodes layer, and its key, and constructs the
// its key in the underlying datastore.
func (t Tree) nodeKey(level uint16, key []byte) []byte {
	// node key prefix is of len 7
	bz := make([]byte, nodeKeyPrefixLen+2+len(key))
	copy(bz, nodeKeyPrefix)
	binary.BigEndian.PutUint16(bz[5:], level)
	copy(bz[nodeKeyPrefixLen+2:], key)
	return bz
}

// leafKey constructs a key for a node pointer representing a leaf node.
func (t Tree) leafKey(key []byte) []byte {
	return t.nodeKey(0, key)
}

func (t Tree) root() *ptr {
	iter := stypes.KVStoreReversePrefixIterator(t.store, nodeKeyPrefix)
	defer iter.Close()
	if !iter.Valid() {
		return nil
	}
	key := iter.Key()[5:]
	return &ptr{
		tree:  t,
		level: binary.BigEndian.Uint16(key[:2]),
		key:   key[2:],
	}
}

// Get returns the (sdk.Int) accumulation value at a given leaf.
func (t Tree) Get(key []byte) sdk.Int {
	res := new(Leaf)
	keybz := t.leafKey(key)
	if !t.store.Has(keybz) {
		return sdk.ZeroInt()
	}
	bz := t.store.Get(keybz)
	err := proto.Unmarshal(bz, res)
	if err != nil {
		panic(err)
	}
	return res.Leaf.Accumulation
}

func (ptr *ptr) create(node *Node) {
	keybz := ptr.tree.nodeKey(ptr.level, ptr.key)
	bz, err := proto.Marshal(node)
	if err != nil {
		panic(err)
	}
	ptr.tree.store.Set(keybz, bz)
}

func (t Tree) ptrGet(level uint16, key []byte) *ptr {
	return &ptr{
		tree:  t,
		level: level,
		key:   key,
	}
}

func (t Tree) ptrIterator(level uint16, begin, end []byte) ptrIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodeKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodeKey(level, nil))
	}
	return ptrIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.Iterator(t.nodeKey(level, begin), endBytes),
	}
}

func (t Tree) ptrReverseIterator(level uint16, begin, end []byte) ptrIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodeKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodeKey(level, nil))
	}
	return ptrIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.ReverseIterator(t.nodeKey(level, begin), endBytes),
	}
}

func (t Tree) Iterator(begin, end []byte) store.Iterator {
	return t.ptrIterator(0, begin, end)
}

func (t Tree) ReverseIterator(begin, end []byte) store.Iterator {
	return t.ptrReverseIterator(0, begin, end)
}

// accumulationSplit returns the accumulated value for all of the following:
// left: all leaves under nodePointer with key < provided key
// exact: leaf with key = provided key
// right: all leaves under nodePointer with key > provided key
// Note that the equalities here are _exclusive_.
func (ptr *ptr) accumulationSplit(key []byte) (left sdk.Int, exact sdk.Int, right sdk.Int) {
	left, exact, right = sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	if ptr.isLeaf() {
		var leaf Leaf
		bz := ptr.tree.store.Get(ptr.tree.leafKey(ptr.key))
		err := proto.Unmarshal(bz, &leaf)
		if err != nil {
			panic(err)
		}
		// Check if the leaf key is to the left of the input key,
		// if so this value is on the left. Similar for the other cases.
		// Recall that all of the output arguments default to 0, if unset internally.
		switch bytes.Compare(ptr.key, key) {
		case -1:
			left = leaf.Leaf.Accumulation
		case 0:
			exact = leaf.Leaf.Accumulation
		case 1:
			right = leaf.Leaf.Accumulation
		}
		return
	}

	node := ptr.node()
	idx, match := node.find(key)
	if !match {
		idx--
	}
	left, exact, right = ptr.tree.ptrGet(ptr.level-1, node.Children[idx].Index).accumulationSplit(key)
	left = left.Add(NewNode(node.Children[:idx]...).accumulate())
	right = right.Add(NewNode(node.Children[idx+1:]...).accumulate())
	return left, exact, right
}

// TotalAccumulatedValue returns the sum of the weights for all leaves.
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

func (ptr *ptr) visualize(depth int, acc sdk.Int) {
	if !ptr.exists() {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("- ")
	fmt.Printf("{%d %+v %v}\n", ptr.level, ptr.key, acc)
	for i, child := range ptr.node().Children {
		childnode := ptr.child(uint16(i))
		childnode.visualize(depth+1, child.Accumulation)
	}
}

// DebugVisualize prints the entire tree to stdout.
func (t Tree) DebugVisualize() {
	t.root().visualize(0, sdk.Int{})
}
