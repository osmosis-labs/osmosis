/// B+ tree implementation on KVStore

package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	store "github.com/cosmos/cosmos-sdk/store"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Tree is an augmented B+ tree implementation.
// Branches have m sized key index slice. Each key index represents
// the starting index of the child nodePointer's index(inclusive), and the
// ending index of the previous nodePointer of the child nodePointer's index(exclusive).
// TODO: We should abstract out the leaves of this tree to allow more data aside from
// the accumulation value to go there.
type Tree struct {
	store store.KVStore
	m     uint8
}

func NewTree(store store.KVStore, m uint8) Tree {
	tree := Tree{store, m}
	tree.Set(nil, 0)
	return tree
}

// nodePointer is a pointer to a node inside of the tree
// This specifies how we access a node in the tree, and gets pointers to the nodes children.
// TODO: Revisit architecture of this.
type nodePointer struct {
	tree  Tree
	level uint16
	// key of the nodePointer is always the first element of the nodePointer.Index
	key []byte
	// XXX: cache stored value?
}

// node is a node in the tree.
type node struct {
	Index []byte
	Acc   uint64
}

type children []node // max length M slice of key bytes, sorted by index

// nodeIterator iterates over nodes in a given level. It only iterates directly over the pointers
// to the nodes, not the actual nodes themselves, to save loading additional data into memory.
type nodeIterator struct {
	tree  Tree
	level uint16
	store.Iterator
}

// nodePointerKey takes in a nodes layer, and its key, and constructs the
// its key in the underlying datastore.
func (t Tree) nodePointerKey(level uint16, key []byte) []byte {
	bz := make([]byte, 2)
	binary.BigEndian.PutUint16(bz, level)
	return append(append([]byte("nodePointer/"), bz...), key...)
}

// leafKey constructs a key for a node pointer representing a leaf node.
func (t Tree) leafKey(key []byte) []byte {
	return t.nodePointerKey(0, key)
}

// root returns the node pointer of the root of the tree.
func (t Tree) root() *nodePointer {
	// TODO: Why does this work, what is the root key here?
	iter := stypes.KVStoreReversePrefixIterator(t.store, []byte("nodePointer/"))
	if !iter.Valid() {
		return nil
	}
	key := iter.Key()[5:]
	return &nodePointer{
		tree:  t,
		level: binary.BigEndian.Uint16(key[:2]),
		key:   key[2:],
	}
}

// Get returns the (uint64) value at a given leaf.
func (t Tree) Get(key []byte) (res uint64) {
	keybz := t.leafKey(key)
	if !t.store.Has(keybz) {
		return 0
	}
	err := json.Unmarshal(t.store.Get(keybz), &res)
	if err != nil {
		panic(err)
	}
	return
}

func (t Tree) Set(key []byte, acc uint64) {
	nodePointer := t.nodePointerGet(0, key)
	nodePointer.setLeaf(acc)

	nodePointer.parent().push(node{key, acc})
}

func (t Tree) Remove(key []byte) {
	nodePointer := t.nodePointerGet(0, key)
	if !nodePointer.exists() {
		return
	}
	parent := nodePointer.parent()
	nodePointer.delete()
	parent.pull(key)
}

func (t Tree) nodePointerGet(level uint16, key []byte) *nodePointer {
	return &nodePointer{
		tree:  t,
		level: level,
		key:   key,
	}
}

// XXX: store.Iterator -> custom nodePointer iterator
func (t Tree) nodeIterator(level uint16, begin, end []byte) nodeIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodePointerKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodePointerKey(level, nil))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.Iterator(t.nodePointerKey(level, begin), endBytes),
	}
}

func (t Tree) nodePointerReverseIterator(level uint16, begin, end []byte) nodeIterator {
	var endBytes []byte
	if end != nil {
		endBytes = t.nodePointerKey(level, end)
	} else {
		endBytes = stypes.PrefixEndBytes(t.nodePointerKey(level, nil))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.ReverseIterator(t.nodePointerKey(level, begin), endBytes),
	}
}

func (t Tree) Iterator(begin, end []byte) store.Iterator {
	return t.nodeIterator(0, begin, end)
}

func (t Tree) ReverseIterator(begin, end []byte) store.Iterator {
	return t.nodePointerReverseIterator(0, begin, end)
}

func (nodePointer *nodePointer) accSplit(key []byte) (left uint64, exact uint64, right uint64) {
	// If the current node is a leaf node, then ...
	if nodePointer.level == 0 {
		var err error
		bz := nodePointer.tree.store.Get(nodePointer.tree.leafKey(nodePointer.key))
		switch bytes.Compare(nodePointer.key, key) {
		case -1:
			err = json.Unmarshal(bz, &left)
		case 0:
			err = json.Unmarshal(bz, &exact)
		case 1:
			err = json.Unmarshal(bz, &right)
		}
		if err != nil {
			panic(err)
		}
		return
	}

	children := nodePointer.children()
	idx, match := children.find(key)
	if !match {
		idx--
	}
	left, exact, right = nodePointer.tree.nodePointerGet(nodePointer.level-1, children[idx].Index).accSplit(key)
	left += children[:idx].accumulate()
	right += children[idx+1:].accumulate()
	return
}

// TotalAccumulatedValue returns the sum of the weights for all leaves
func (t Tree) TotalAccumulatedValue() uint64 {
	return t.SubsetAccumulation(nil, nil)
}

// Prefix sum returns the total weight of all leaves with keys <= to the provided key.
func (t Tree) PrefixSum(key []byte) uint64 {
	return t.SubsetAccumulation(nil, key)
}

// SubsetAccumulation returns the total value of all leaves with keys
// between start and end (inclusive of both ends)
// if start is nil, it is the beginning of the tree.
// if end is nil, it is the end of the tree.
func (t Tree) SubsetAccumulation(start []byte, end []byte) uint64 {
	if start == nil {
		left, exact, _ := t.root().accSplit(end)
		return left + exact
	}
	if end == nil {
		_, exact, right := t.root().accSplit(start)
		return exact + right
	}
	_, leftexact, leftrest := t.root().accSplit(start)
	_, _, rightest := t.root().accSplit(end)
	return leftexact + leftrest - rightest
}

func (t Tree) SplitAcc(key []byte) (uint64, uint64, uint64) {
	return t.root().accSplit(key)
}

func (nodePointer *nodePointer) visualize(depth int, acc uint64) {
	if !nodePointer.exists() {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("- ")
	fmt.Printf("{%d %+v %d}\n", nodePointer.level, nodePointer.key, acc)
	for i, child := range nodePointer.children() {
		childnodePointer := nodePointer.child(uint16(i))
		childnodePointer.visualize(depth+1, child.Acc)
	}
}

// DebugVisualize prints out the entire tree to stdout
func (t Tree) DebugVisualize() {
	t.root().visualize(0, 0)
}
