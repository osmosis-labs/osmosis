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

// Tree is a modified B+ tree implementation.
// Branches have m sized key index slice. Each key index represents
// the starting index of the child node's index(inclusive), and the
// ending index of the previous node of the child node's index(exclusive).
type Tree struct {
	store store.KVStore
	m     uint8
}

func NewTree(store store.KVStore, m uint8) Tree {
	tree := Tree{store, m}
	tree.Set(nil, 0)
	return tree
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
	start, end := iter.Domain()
	fmt.Printf("domain %+v %+v\n", start, end)
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
	fmt.Printf("set %+v %+v\n", node.tree.nodeKey(node.level, node.key), bz)
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
	fmt.Printf("setLeaf %+v %+v\n", node.tree.leafKey(node.key), bz)
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
	return node.tree.nodeReverseIterator(node.level, nil, node.key).node()
}

func (node *node) rightSibling() *node {
	iter := node.tree.nodeIterator(node.level, node.key, nil)
	fmt.Printf("t")
	if !iter.Valid() {
		return nil
	}
	fmt.Printf("g")
	if node.exists() {
		fmt.Printf("h")
		// exclude node itself
		iter.Next()
	}
	fmt.Printf("j")
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
	parent = parent.leftSibling()
	if parent.exists() {
		return parent
	}
	return node.tree.nodeGet(node.level+1, nil)
}

func (node *node) exists() bool {
	if node == nil {
		fmt.Printf("www")
		return false
	}
	fmt.Printf("exists %+v\n", node.tree.nodeKey(node.level, node.key))
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
		node.create(children{c})
		return
	}

	cs := node.children()
	idx, match := cs.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		node.updateAccumulation(c)
		return
	}

	fmt.Printf("prev %+v\n", cs)
	// inserting new child node
	cs = cs.insert(idx, c)

	fmt.Printf("next %+v\n", cs)
	// split and push-up if overflow
	if len(cs) > int(node.tree.m) {
		split := node.tree.m/2 + 1
		parent := node.parent()
		leftchildren, rightchildren := cs.split(int(split))
		node.tree.nodeGet(node.level, rightchildren.key()).create(rightchildren)
		if !parent.exists() {
			parent.create(children{
				child{node.key, leftchildren.accumulate()},
				child{cs[split].Index, rightchildren.accumulate()},
			})
			node.set(leftchildren)
			return
		}
		// constructing right childdd
		parent.push(child{rightchildren.key(), rightchildren.accumulate()})
		cs = leftchildren
		parent.updateAccumulation(child{node.key, leftchildren.accumulate()})
	}

	node.set(cs)
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

	return len(children), false
}

func (children children) set(idx int, child child) children {
	children[idx] = child
	return children
}

func (children children) setAcc(idx int, acc uint64) children {
	children[idx] = child{children[idx].Index, acc}
	return children
}

func (cs children) insert(idx int, c child) children {
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

// Root: (level, key) of the root node
func (t Tree) rootKey() []byte {
	return []byte("root")
}

// key of the node is always the first element of the node.Index
func (t Tree) nodeKey(level uint16, key []byte) []byte {
	bz := make([]byte, 2)
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

func (node *node) create(children children) {
	keybz := node.tree.nodeKey(node.level, node.key)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	fmt.Printf("nodeNew %+v %+v\n", keybz, bz)
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

func (t Tree) Set(key []byte, acc uint64) {
	node := t.nodeGet(0, key)
	node.setLeaf(acc)

	node.parent().push(child{key, acc})
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
		var err error
		bz := node.tree.store.Get(node.tree.leafKey(node.key))
		switch bytes.Compare(node.key, key) {
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

	children := node.children()
	idx, match := children.find(key)
	if !match {
		idx--
	}
	fmt.Printf("acc %+v, %+v, %+v %d\n", key, node, children, idx)
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
