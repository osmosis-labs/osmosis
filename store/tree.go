// B+ tree implementation on KVStore

package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	store "github.com/cosmos/cosmos-sdk/store"
)

// Tree is a modified B+ tree implementation.
// Branches have m sized key index slice. Each key index represents
// the starting index of the child node's index(inclusive), and the
// ending index of the previous node of the child node's index(exclusive).
type Tree struct {
	store store.KVStore
	m uint8
}

// node is pointer to a specific node inside the tree
type node struct {
	tree Tree
	level uint16
	key []byte
	// XXX: cache stored value?
}

type nodeIterator struct {
	tree Tree
	level uint16
	store.Iterator
}

func (iter nodeIterator) node() *node {
	res := node{
		tree: iter.tree,
		level: iter.level,
		key: iter.Key(),
	}
	return &res
}

func (node *node) get() (res nodeData) {
	bz := node.tree.store.Get(node.tree.nodeKey(node.level, node.key))
	if bz != nil {
		json.Unmarshal(bz, &res)
	}
	return
}

func (node *node) set(data nodeData) {
	bz, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	node.tree.store.Set(node.tree.nodeKey(node.level, node.key), bz)
}

func (node *node) setLeaf(value []byte) {
	if node.level != 0 {
		panic("setLeaf should not be called on branch node")
	}
	node.tree.store.Set(node.tree.leafKey(node.key), value)
}

func (node *node) delete() {
	node.tree.store.Delete(node.tree.nodeKey(node.level, node.key))
}

func (node *node) left() *node {
	return node.tree.nodeReverseIterator(node.level, nil, node.key).node()
}

func (node *node) right() *node {
	iter := node.tree.nodeIterator(node.level, node.key, nil)
	iter.Next()
	return iter.node()
}

func (node *node) child(n uint16) *node {
	return node.tree.nodeIterator(node.level-1, index[n], nil).node()
}

func (node *node) customDataUpdate() {
	// XXX
}

func (node *node) parent() *node {	
	// first child inclusive case
	parent := node.tree.nodeGet(node.level+1, node.key)
	if parent.exists() {
		return parent
	}
	return parent.left()

	/*
	// sandwitch case
	iter := node.tree.nodeReverseIterator(node.level+1, nil, node.key)
	parent = iter.node()
	index := parent.get().Index
	lastindex := index[len(index)-1]
	if bytes.Compare(lastindex, node.key) == 1 {
		return parent
	}
	*/
}

func (node *node) exists() bool {
	return node.tree.store.Has(node.tree.nodeKey(node.level, node.key))
}

func (node *node) push(key []byte) {
	data := node.get()
	for i, idx := range data.Index {
		// ignore if key already exists
		if bytes.Compare(idx, key) == 0 {
			return
		}
		// Push new key to the appropriate position
		if bytes.Compare(idx, key) > 0 {
			data.Index = append(append(data.Index[:i+1], key), data.Index[i+1:]...)
			break
		}
	}

	// split and push-up if overflow
	if len(data.Index) > int(node.tree.m) {
		split := node.tree.m/2+1
		parent := node.parent()
		if !parent.exists() {
			parent = node.tree.nodeGet(node.level+1, data.Index[split])
		}
		parent.push(data.Index[split])
		node.delete()
		node.tree.nodeNew(node.level, data.Index[:split])
		node.tree.nodeNew(node.level, data.Index[split:])
		return
	}

	node.customDataUpdate()

	node.set(data)
}

func (node *node) pull(key []byte) {
	data := node.get()
	for i, idx := range data.Index {
		if bytes.Compare(idx, key) == 0 {
			data.Index = append(data.Index[:i], data.Index[i+1:]...)
			break
		}
	}

	// For sake of efficienty on our use case, we pull only when a node gets
	// empty.
	// if len(data.Index) >= int(node.tree.m/2) {
	if len(data.Index) > 0 {
		node.set(data)
		return
	}

	// merge if possible
	left := node.left()
	right := node.right()
	node.delete()
	parent.pull(node.key)
	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		if bytes.Equal(parent.key, right.parent().key)) {
			leftIndex := left.get().Index
			rightIndex := right.get().Index
			if len(leftIndex)+len(rightIndex) < int(node.tree.m) {
				leftIndex = append(leftIndex, rightIndex...)
				left.set(nodeData{Index: leftIndex})
				right.delete()
				parent.pull(right.key)
			}
		}
	}
}

// nodeData is struct for internal nodes
// marshaled and stored inside the state.
type nodeData struct {
	Index [][]byte // max length M slice of key bytes, sorted
	// XXX: custom data interface
}

// Root: (level, key) of the root node
func (t Tree) rootKey() []byte {
	return []byte("root")
}

// key of the node is always the first element of the node.Index
func (t Tree) nodeKey(level uint16, key []byte) []byte {
	bz := make([]byte, 4)
	binary.BigEndian.PutUint16(bz, level)
	return append(append([]byte("node/"), bz...), key...)
}

func (t Tree) leafKey(key []byte) []byte {
	return t.nodeKey(0, key)
}

func (t Tree) Get(key []byte) []byte {
	keybz := t.leafKey(key)
	if !t.store.Has(keybz) {
		return nil
	}
	return t.store.Get(keybz)
}

func (t Tree) nodeNew(level uint16, index [][]byte) *node {
	keybz := t.nodeKey(level, index[0])
	bz, err := json.Marshal(nodeData{
		Index: index,
	})
	if err != nil {
		panic(err)
	}
	t.store.Set(keybz, bz)

	node := node{
		tree: t,
		level: level,
		key: index[0],
	}

	node.customDataUpdate()

	return &node
}

func (t Tree) nodeGet(level uint16, key []byte) *node {
	return &node{
		tree: t,
		level: level,
		key: key,
	}
}

// XXX: store.Iterator -> custom node iterator
func (t Tree) nodeIterator(level uint16, begin, end []byte) nodeIterator {
	return nodeIterator{
		tree: t,
		level: level,
		Iterator: t.store.Iterator(t.nodeKey(level, begin), t.nodeKey(level, end)),
	}
}

func (t Tree) nodeReverseIterator(level uint16, begin, end []byte) nodeIterator {
	return nodeIterator{
		tree: t,
		level: level,
		Iterator: t.store.ReverseIterator(t.nodeKey(level, begin), t.nodeKey(level, end)),
	}
}

func (t Tree) Iterator(begin, end []byte) store.Iterator {
	return t.nodeIterator(0, begin, end)
}

func (t Tree) ReverseIterator(begin, end []byte) store.Iterator {
	return t.nodeReverseIterator(0, begin, end)
}

func (t Tree) Set(key, value []byte) {
	node := t.nodeGet(0, key)
	if !node.exists() {
		node.setLeaf(value)
	}
	node.parent().push(key)
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
