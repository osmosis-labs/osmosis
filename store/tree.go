// B+ tree implementation on KVStore

package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	store "github.com/cosmos/cosmos-sdk/store"
)

// Tree is a B+ tree implementation.
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

func (node *node) delete() {
	node.tree.store.Delete(node.tree.nodeKey(node.level, node.key))
}

func (node *node) prev() *node {
	return node.tree.nodeReverseIterator(node.level, nil, node.key).node()
}

func (node *node) next() *node {
	return node.tree.nodeIterator(node.level, node.key, nil).node()
}

func (node *node) child(n uint16) *node {
	index := node.get().Index
	if n == 0 {
		node.tree.nodeReverseIterator(node.level-1, nil, index[0]).node()
	}
	return node.tree.nodeIterator(node.level-1, index[n-1], index[n]).node()
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

	// sandwitch case
	iter := node.tree.nodeReverseIterator(node.level+1, nil, node.key)
	parent = iter.node()
	index := parent.get().Index
	lastindex := index[len(index)-1]
	if bytes.Compare(lastindex, node.key) == 1 {
		return parent
	}

	// edge case, left parent
	if bytes.Compare(lastindex, node.prev().key) == 1 {
		return parent
	}

	// edge case, right parent
	if bytes.Compare(lastindex, node.prev().key) == -1 {
		iter.Next()
		return iter.node()
	}

	panic("should not reach here")
}

func (node *node) exists() bool {
	return node.tree.store.Has(node.tree.nodeKey(node.level, node.key))
}

func (node *node) push(key []byte) {
	data := node.get()
	for i, idx := range data.Index
		// Push new key to the appropriate position
		if bytes.Compare(idx, key) >= 0 {
			// XXX: look tomorrow, maybe i-1 instead of i, brain not functioning now
			data.Index = append(append(data.Index[:i], key), data.Index[i:]...)
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

// XXX: or lets simply pull only when a node gets empty.
// 
func (node *node) pull(key []byte) {
	data := node.get()
	for i, idx := range data.Index {
		if bytes.Compare(idx, key) >= 0 {
			data.Index = append(data.Index[:i], data.Index[i+1:]...)
			break
		}
	}

	if len(data.Index) >= node.tree.m/2 {
		node.set(data)
		return
	}

	// redistribute and pull-down if underflow
	// XXX: redistribute
	
	// XXX: merge

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
	// set the leaf node
	keybz := t.leafKey(key)
	t.store.Set(keybz, value)



	// push if not overflow
	
}
