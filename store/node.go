package store

import (
	"bytes"
	"encoding/json"
)

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
	bz := nodePointer.tree.store.Get(nodePointer.tree.nodeKey(nodePointer.level, nodePointer.key))
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
	nodePointer.tree.store.Set(nodePointer.tree.nodeKey(nodePointer.level, nodePointer.key), bz)
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

// delete removes the corresponding node from the underlying data store,
func (nodePointer *nodePointer) delete() {
	nodePointer.tree.store.Delete(nodePointer.tree.nodeKey(nodePointer.level, nodePointer.key))
}

func (nodePointer *nodePointer) leftSibling() *nodePointer {
	return nodePointer.tree.nodeReverseIterator(nodePointer.level, nil, nodePointer.key).nodePointer()
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
	return nodePointer.tree.store.Has(nodePointer.tree.nodeKey(nodePointer.level, nodePointer.key))
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
		// constructing right child
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
	// For sake of efficienty of our use case, we pull only when a nodePointer gets
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

// accumulate returns the sum of the values of all the children.
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

func (nodePointer *nodePointer) create(children children) {
	keybz := nodePointer.tree.nodeKey(nodePointer.level, nodePointer.key)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	nodePointer.tree.store.Set(keybz, bz)
}
