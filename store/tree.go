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

func NewTree(store store.KVStore, m uint8) Tree {
	return Tree{store, m}
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
	if !iter.Valid() {
		return nil
	}
	res := node{
		tree: iter.tree,
		level: iter.level,
		key: iter.Key(),
	}
	return &res
}

func (node *node) children() (res Children) {
	bz := node.tree.store.Get(node.tree.nodeKey(node.level, node.key))
	if bz != nil {
		json.Unmarshal(bz, &res)
	}
	return
}

func (node *node) set(children Children) {
	bz, err := json.Marshal(children)
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

func (node *node) leftSibling() *node {
	return node.tree.nodeReverseIterator(node.level, nil, node.key).node()
}

func (node *node) rightSibling() *node {
	iter := node.tree.nodeIterator(node.level, node.key, nil)
	if !iter.Valid() {
		return nil
	}
	iter.Next()
	return iter.node()
}

func (node *node) child(n uint16) *node {
	return node.tree.nodeIterator(node.level-1, node.children()[n].Index, nil).node()
}

func (node *node) parent() *node {
	// first child inclusive case
	parent := node.tree.nodeGet(node.level+1, node.key)
	if parent.exists() {
		return parent
	}
	return parent.leftSibling()
}

func (node *node) exists() bool {
	return node.tree.store.Has(node.tree.nodeKey(node.level, node.key))
}

func (node *node) pushLeaf(key []byte) {
	node.push(key, nil)
}

func (node *node) updateAccumulation(c child) {
	if node == nil {
		return // reached at the root
	}

	children := node.children()
	idx, match := children.find(c.Index)
	if !match {
		panic("non existing key pushed from the child")
	}
	children = children.setAcc(idx, c.Acc)
	node.set(children)
	node.parent().updateAccumulation(child{node.key, children.accumulate()})
}

func (node *node) push(c child) {
	if node == nil {
		return // reached at the root
	}
	children := node.children()
	idx, match := children.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		node.updateAccumulation(c)
		return
	}

	// inserting new child node
	children = children.insert(c)

	// split and push-up if overflow
	if len(children) > int(node.tree.m) {
		split := node.tree.m/2+1
		parent := node.parent()
		// XXX: do we need this?
		if !parent.exists() {
			parent = node.tree.nodeGet(node.level+1, children[split].Index)
		}
		leftChildren, rightChildren := children.split(int(split))
		// constructing right child
		node.tree.nodeNew(node.level, rightChildren)
		parent.push(child{rightChildren.key(), rightChildren.accumulate()})
		children = leftChildren
		parent.updateAccumulation(child{node.key, leftChildren.accumulate()})
	}

	node.set(children)
}

func (node *node) pull(key []byte) {
	if node == nil {
		return // reached at the root
	}
	children := node.children()
	idx, match := children.find(key)

	if !match {
		panic("pulling non existing child")
	}

	children = children.delete(idx)
	// For sake of efficienty on our use case, we pull only when a node gets
	// empty.
	// if len(data.Index) >= int(node.tree.m/2) {
	if len(children) > 0 {
		node.set(children)
		node.parent().updateAccumulation(child{node.key, children.accumulate()})
		return
	}

	// merge if possible
	left := node.leftSibling()
	right := node.rightSibling()
	parent := node.parent()
	node.delete()
	parent.pull(node.key)

	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		if bytes.Equal(parent.key, right.parent().key) {
			leftChildren := left.children()
			rightChildren := right.children()
			if len(leftChildren)+len(rightChildren) < int(node.tree.m) {
				left.set(leftChildren.merge(rightChildren))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(child{left.key, leftChildren.accumulate()})
			}
		}
	}
}

type child struct {
	Index []byte
	Acc uint64
}

type Children []child // max length M slice of key bytes, sorted by index

func (children Children) key() []byte {
	return children[0].Index
}

func (children Children) accumulate() (res uint64) {
	for _, child := range children {
		res += child.Acc
	}
	return
}

// find returns the appropriate position that key should be inserted
// if match is true, idx is the exact position for the key
// if match is false, idx is the position where the key should be inserted
func (children Children) find(key []byte) (idx int, match bool) {
	for idx, child := range children {
		if bytes.Compare(child.Index, key) == 0 {
			return idx, true
		}
		// Push new key to the appropriate position
		if bytes.Compare(child.Index, key) > 0 {
			return idx, false 
		}
	}

	panic("should not reach here")
}

func (children Children) set(idx int, child child) Children {
	children[idx] = child
	return children
}

func (children Children) setAcc(idx int, acc uint64) Children {
	children[idx] = child{children[idx].Index, acc}
	return children
}

func (children Children) insert(idx int, child child) Children {
	children = append(append(children[:idx], child), children[idx:]...)
	return children
}

func (children Children) delete(idx int) Children {
	children = append(children[:idx], children[idx+1:]...)
	return children
}

func (children Children) split(idx int) (Children, Children) {
	return children[:idx], children[idx:]
}

func (children Children) merge(children2 Children) Children {
	return append(children, children2...)
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

func (t Tree) nodeNew(level uint16, children Children) *node {
	keybz := t.nodeKey(level, children[0].Index)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	t.store.Set(keybz, bz)

	node := node{
		tree: t,
		level: level,
		key: children.key(),
	}

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

func (t Tree) Set(key []byte, acc uint64) {
	node := t.nodeGet(0, key)
	node.setLeaf(value)

	parent := t.nodeGet(1, key)
	parent.pushLeaf(key, acc)
}

func (t Tree) Remove(key []byte) {
	node := t.nodeGet(0, key)
	if !node.exists() {
		return
	}
	parent := t.nodeGet(1, key)
	node.delete()
	parent.pull(key)
}
