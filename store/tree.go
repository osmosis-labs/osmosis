// B+ tree implementation on KVStore

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
	key   []byte
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

func (iter nodeIterator) nodePointer() *nodePointer {
	if !iter.Valid() {
		return nil
	}
	res := nodePointer{
		tree:  iter.tree,
		level: iter.level,
		key:   iter.Key()[7:],
	}
	return &res
}

func (nodePointer *nodePointer) children() (res children) {
	bz := nodePointer.tree.store.Get(nodePointer.tree.nodePointerKey(nodePointer.level, nodePointer.key))
	if bz != nil {
		json.Unmarshal(bz, &res)
	}
	return
}

func (nodePointer *nodePointer) set(children children) {
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	nodePointer.tree.store.Set(nodePointer.tree.nodePointerKey(nodePointer.level, nodePointer.key), bz)
}

func (nodePointer *nodePointer) setLeaf(acc uint64) {
	if nodePointer.level != 0 {
		panic("setLeaf should not be called on branch nodePointer")
	}
	bz, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	nodePointer.tree.store.Set(nodePointer.tree.leafKey(nodePointer.key), bz)
}

func (nodePointer *nodePointer) delete() {
	nodePointer.tree.store.Delete(nodePointer.tree.nodePointerKey(nodePointer.level, nodePointer.key))
}

func (nodePointer *nodePointer) leftSibling() *nodePointer {
	return nodePointer.tree.nodePointerReverseIterator(nodePointer.level, nil, nodePointer.key).nodePointer()
}

func (nodePointer *nodePointer) rightSibling() *nodePointer {
	iter := nodePointer.tree.nodeIterator(nodePointer.level, nodePointer.key, nil)
	if !iter.Valid() {
		return nil
	}
	if nodePointer.exists() {
		// exclude nodePointer itself
		iter.Next()
	}
	return iter.nodePointer()
}

func (nodePointer *nodePointer) child(n uint16) *nodePointer {
	// TODO: set end to prefix iterator end
	return nodePointer.tree.nodeIterator(nodePointer.level-1, nodePointer.children()[n].Index, nil).nodePointer()
}

func (nodePointer *nodePointer) parent() *nodePointer {
	// first child inclusive case
	parent := nodePointer.tree.nodePointerGet(nodePointer.level+1, nodePointer.key)
	if parent.exists() {
		return parent
	}
	parent = parent.leftSibling()
	if parent.exists() {
		return parent
	}
	return nodePointer.tree.nodePointerGet(nodePointer.level+1, nil)
}

func (nodePointer *nodePointer) exists() bool {
	if nodePointer == nil {
		return false
	}
	return nodePointer.tree.store.Has(nodePointer.tree.nodePointerKey(nodePointer.level, nodePointer.key))
}

func (nodePointer *nodePointer) updateAccumulation(c node) {
	if !nodePointer.exists() {
		return // reached at the root
	}

	children := nodePointer.children()
	idx, match := children.find(c.Index)
	if !match {
		panic("non existing key pushed from the child")
	}
	children = children.setAcc(idx, c.Acc)
	nodePointer.set(children)
	nodePointer.parent().updateAccumulation(node{nodePointer.key, children.accumulate()})
}

func (nodePointer *nodePointer) push(c node) {
	if !nodePointer.exists() {
		nodePointer.create(children{c})
		return
	}

	cs := nodePointer.children()
	idx, match := cs.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		nodePointer.updateAccumulation(c)
		return
	}

	// inserting new child nodePointer
	cs = cs.insert(idx, c)
	parent := nodePointer.parent()

	// split and push-up if overflow
	if len(cs) > int(nodePointer.tree.m) {
		split := nodePointer.tree.m/2 + 1
		leftchildren, rightchildren := cs.split(int(split))
		nodePointer.tree.nodePointerGet(nodePointer.level, cs[split].Index).create(rightchildren)
		if !parent.exists() {
			parent.create(children{
				node{nodePointer.key, leftchildren.accumulate()},
				node{cs[split].Index, rightchildren.accumulate()},
			})
			nodePointer.set(leftchildren)
			return
		}
		// constructing right childdd
		parent.push(node{cs[split].Index, rightchildren.accumulate()})
		cs = leftchildren
		parent = nodePointer.parent() // parent might be changed during the pushing process
	}

	parent.updateAccumulation(node{nodePointer.key, cs.accumulate()})
	nodePointer.set(cs)
}

func (nodePointer *nodePointer) pull(key []byte) {

	if !nodePointer.exists() {
		return // reached at the root
	}
	children := nodePointer.children()
	idx, match := children.find(key)

	if !match {
		panic("pulling non existing child")
	}

	children = children.delete(idx)
	// For sake of efficienty on our use case, we pull only when a nodePointer gets
	// empty.
	// if len(data.Index) >= int(nodePointer.tree.m/2) {
	if len(children) > 0 {
		nodePointer.set(children)
		nodePointer.parent().updateAccumulation(node{nodePointer.key, children.accumulate()})
		return
	}

	// merge if possible
	left := nodePointer.leftSibling()
	right := nodePointer.rightSibling()
	parent := nodePointer.parent()
	nodePointer.delete()
	parent.pull(nodePointer.key)

	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		if bytes.Equal(parent.key, right.parent().key) {
			leftchildren := left.children()
			rightchildren := right.children()
			if len(leftchildren)+len(rightchildren) < int(nodePointer.tree.m) {
				left.set(leftchildren.merge(rightchildren))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(node{left.key, leftchildren.accumulate()})
			}
		}
	}
}

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
		if bytes.Equal(child.Index, key) {
			return idx, true
		}
		// Push new key to the appropriate position
		if bytes.Compare(child.Index, key) > 0 {
			return idx, false
		}
	}

	return len(children), false
}

func (children children) set(idx int, child node) children {
	children[idx] = child
	return children
}

func (children children) setAcc(idx int, acc uint64) children {
	children[idx] = node{children[idx].Index, acc}
	return children
}

func (cs children) insert(idx int, c node) children {
	return append(cs[:idx], append(children{c}, cs[idx:]...)...)
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

// Root: (level, key) of the root nodePointer
func (t Tree) rootKey() []byte {
	return []byte("root")
}

// key of the nodePointer is always the first element of the nodePointer.Index
func (t Tree) nodePointerKey(level uint16, key []byte) []byte {
	bz := make([]byte, 2)
	binary.BigEndian.PutUint16(bz, level)
	return append(append([]byte("nodePointer/"), bz...), key...)
}

func (t Tree) leafKey(key []byte) []byte {
	return t.nodePointerKey(0, key)
}

func (t Tree) root() *nodePointer {
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

func (nodePointer *nodePointer) create(children children) {
	keybz := nodePointer.tree.nodePointerKey(nodePointer.level, nodePointer.key)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	nodePointer.tree.store.Set(keybz, bz)
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

func (nodePointer *nodePointer) accSplit(key []byte) (left uint64, exact uint64, right uint64) {
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
