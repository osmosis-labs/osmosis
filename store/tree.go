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

// Tree is a modified B+ tree implementation.
// Branches have m sized key index slice. Each key index represents
// the starting index of the child node's index(inclusive), and the
// ending index of the previous node of the child node's index(exclusive).
type Tree struct {
	store store.KVStore
	m     uint8
}

func NewTree(store store.KVStore, m uint8) Tree {
	return Tree{store, m}
}

// node is pointer to a specific node inside the tree
type node struct {
	tree  Tree
	level uint16
	key   []byte
	// XXX: cache stored value?
}

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
		key:   iter.Key(),
	}
	return &res
}

func (node *node) children() (res children) {
	bz := node.tree.store.Get(node.tree.nodeKey(node.level, node.key))
	if bz != nil {
		json.Unmarshal(bz, &res)
	}
	return
}

func (node *node) set(children children) {
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	node.tree.store.Set(node.tree.nodeKey(node.level, node.key), bz)
}

func (node *node) setLeaf(acc uint64) {
	if node.level != 0 {
		panic("setLeaf should not be called on branch node")
	}
	bz, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	node.tree.store.Set(node.tree.leafKey(node.key), bz)
}

func (node *node) delete() {
	fmt.Printf("q")
	key := node.tree.nodeKey(node.level, node.key)
	fmt.Printf("%x", key)
	node.tree.store.Delete(node.tree.nodeKey(node.level, node.key))
	fmt.Printf("w")
}

func (node *node) leftSibling() *node {
	// TODO: set start to prefix start
	return node.tree.nodeReverseIterator(node.level, nil, node.key).node()
}

func (node *node) rightSibling() *node {
	fmt.Printf("%+v\n", node)
	// TODO: set end to prefix iterator end
	iter := node.tree.nodeIterator(node.level, node.key, nil)
	fmt.Printf("a")
	if !iter.Valid() {
		fmt.Printf("b")
		return nil
	}
	fmt.Printf("c")
	iter.Next()
	fmt.Printf("d")
	return iter.node()
}

func (node *node) child(n uint16) *node {
	// TODO: set end to prefix iterator end
	return node.tree.nodeIterator(node.level-1, node.children()[n].Index, nil).node()
}

func (node *node) parent() *node {
	// first child inclusive case
	parent := node.tree.nodeGet(node.level+1, node.key)
	if parent.exists() {
		return parent
	}
	leftSibling := parent.leftSibling()
	if leftSibling.exists() {
		return leftSibling
	}
	// edge case: only happens when pushing new node from the leftmost side
	return parent.rightSibling()
}

func (node *node) exists() bool {
	if node == nil {
		fmt.Printf("www")
		return false
	}
	return node.tree.store.Has(node.tree.nodeKey(node.level, node.key))
}

func (node *node) updateAccumulation(c child) {
	if !node.exists() {
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
	fmt.Printf("push %+v %+v\n", node, c)
	if !node.exists() {
		node.tree.nodeNew(node.level, children{c})
		return
	}

	children := node.children()
	idx, match := children.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		node.updateAccumulation(c)
		return
	}

	// inserting new child node
	children = children.insert(idx, c)

	// split and push-up if overflow
	if len(children) > int(node.tree.m) {
		split := node.tree.m/2 + 1
		parent := node.parent()
		// XXX: do we need this?
		if !parent.exists() {
			parent = node.tree.nodeGet(node.level+1, children[split].Index)
		}
		leftchildren, rightchildren := children.split(int(split))
		// constructing right child
		node.tree.nodeNew(node.level, rightchildren)
		parent.push(child{rightchildren.key(), rightchildren.accumulate()})
		children = leftchildren
		parent.updateAccumulation(child{node.key, leftchildren.accumulate()})
	}

	node.set(children)
}

func (node *node) pull(key []byte) {
	fmt.Printf("pull %x\n", key)

	if !node.exists() {
		return // reached at the root
	}
	children := node.children()
	idx, match := children.find(key)

	fmt.Printf("a")
	if !match {
		panic("pulling non existing child")
	}

	children = children.delete(idx)
	// For sake of efficienty on our use case, we pull only when a node gets
	// empty.
	// if len(data.Index) >= int(node.tree.m/2) {
	if len(children) > 0 {
		fmt.Println("nomerge")
		node.set(children)
		node.parent().updateAccumulation(child{node.key, children.accumulate()})
		return
	}

	fmt.Printf("b")
	// merge if possible
	left := node.leftSibling()
	fmt.Printf("f")
	right := node.rightSibling()
	fmt.Printf("g")
	parent := node.parent()
	fmt.Printf("h")
	fmt.Printf("%+v", node)
	node.delete()
	fmt.Printf("d")
	parent.pull(node.key)

	fmt.Printf("e")
	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		fmt.Printf("c")
		if bytes.Equal(parent.key, right.parent().key) {
			leftchildren := left.children()
			rightchildren := right.children()
			if len(leftchildren)+len(rightchildren) < int(node.tree.m) {
				fmt.Printf("merge %x %x\n", leftchildren.key(), rightchildren.key())
				left.set(leftchildren.merge(rightchildren))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(child{left.key, leftchildren.accumulate()})
			}
		}
	}
}

type child struct {
	Index []byte
	Acc   uint64
}

type children []child // max length M slice of key bytes, sorted by index

func (children children) key() []byte {
	return children[0].Index
}

func (children children) accumulate() (res uint64) {
	for _, child := range children {
		res += child.Acc
	}
	return
}

// find returns the appropriate position that key should be inserted
// if match is true, idx is the exact position for the key
// if match is false, idx is the position where the key should be inserted
func (children children) find(key []byte) (idx int, match bool) {
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

func (children children) set(idx int, child child) children {
	children[idx] = child
	return children
}

func (children children) setAcc(idx int, acc uint64) children {
	children[idx] = child{children[idx].Index, acc}
	return children
}

func (children children) insert(idx int, child child) children {
	children = append(append(children[:idx], child), children[idx:]...)
	return children
}

func (children children) delete(idx int) children {
	children = append(children[:idx], children[idx+1:]...)
	return children
}

func (children children) split(idx int) (children, children) {
	return children[:idx], children[idx:]
}

func (children children) merge(children2 children) children {
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

func (t Tree) root() *node {
	iter := stypes.KVStoreReversePrefixIterator(t.store, []byte("node/"))
	if !iter.Valid() {
		return nil
	}
	key := iter.Key()[5:]
	return &node{
		tree:  t,
		level: binary.BigEndian.Uint16(key[:4]),
		key:   key[4:],
	}
}

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

func (t Tree) nodeNew(level uint16, children children) *node {
	keybz := t.nodeKey(level, children[0].Index)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	t.store.Set(keybz, bz)

	node := node{
		tree:  t,
		level: level,
		key:   children.key(),
	}

	return &node
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
	if end == nil {
		end = stypes.PrefixEndBytes(t.nodeKey(level, end))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
		Iterator: t.store.Iterator(t.nodeKey(level, begin), t.nodeKey(level, end)),
	}
}

func (t Tree) nodeReverseIterator(level uint16, begin, end []byte) nodeIterator {
	if end == nil {
		end = stypes.PrefixEndBytes(t.nodeKey(level, end))
	}
	return nodeIterator{
		tree:     t,
		level:    level,
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
	node.setLeaf(acc)

	parent := node.parent()
	if !parent.exists() {
		fmt.Printf("qqq")
		parent = t.nodeGet(1, key)
	}
	parent.push(child{key, acc})
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

func (node *node) accSplit(key []byte) (left uint64, exact uint64, right uint64) {
	if node.level == 0 {
		if bytes.Equal(node.key, key) {
			// exact leaf node
			keybz := node.tree.leafKey(key)
			err := json.Unmarshal(node.tree.store.Get(keybz), &exact)
			if err != nil {
				panic(err)
			}
		}
		return
	}

	children := node.children()
	fmt.Printf("%+v, %+v, %+v\n", key, node, children)
	idx, match := children.find(key)
	if !match {
		idx--
	}
	left, exact, right = node.tree.nodeGet(node.level-1, children[idx].Index).accSplit(key)
	left += children[:idx].accumulate()
	right += children[idx+1:].accumulate()
	return
}

func (t Tree) SplitAcc(key []byte) (uint64, uint64, uint64) {
	return t.root().accSplit(key)
}

func (t Tree) SliceAcc(start []byte, end []byte) uint64 {
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
