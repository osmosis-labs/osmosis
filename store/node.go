package store

import (
	"bytes"
	"encoding/json"
)

func (iter nodeIterator) nodePtr() *nodePtr {
	if !iter.Valid() {
		return nil
	}
	res := nodePtr{
		tree:  iter.tree,
		level: iter.level,
		key:   iter.Key()[7:],
	}
	return &res
}

func (nodePtr *nodePtr) isLeaf() bool {
	return nodePtr.level == 0
}

func (nodePtr *nodePtr) children() (res children) {
	bz := nodePtr.tree.store.Get(nodePtr.tree.nodeKey(nodePtr.level, nodePtr.key))
	if bz != nil {
		json.Unmarshal(bz, &res)
	}
	return
}

func (nodePtr *nodePtr) set(children children) {
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	nodePtr.tree.store.Set(nodePtr.tree.nodeKey(nodePtr.level, nodePtr.key), bz)
}

func (nodePtr *nodePtr) setLeaf(acc uint64) {
	if !nodePtr.isLeaf() {
		panic("setLeaf should not be called on branch nodePtr")
	}
	bz, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	nodePtr.tree.store.Set(nodePtr.tree.leafKey(nodePtr.key), bz)
}

// delete removes the corresponding node from the underlying data store,
func (nodePtr *nodePtr) delete() {
	nodePtr.tree.store.Delete(nodePtr.tree.nodeKey(nodePtr.level, nodePtr.key))
}

func (nodePtr *nodePtr) leftSibling() *nodePtr {
	return nodePtr.tree.nodeReverseIterator(nodePtr.level, nil, nodePtr.key).nodePtr()
}

func (nodePtr *nodePtr) rightSibling() *nodePtr {
	iter := nodePtr.tree.nodeIterator(nodePtr.level, nodePtr.key, nil)
	if !iter.Valid() {
		return nil
	}
	if nodePtr.exists() {
		// exclude nodePtr itself
		iter.Next()
	}
	return iter.nodePtr()
}

func (nodePtr *nodePtr) child(n uint16) *nodePtr {
	// TODO: set end to prefix iterator end
	return nodePtr.tree.nodeIterator(nodePtr.level-1, nodePtr.children()[n].Index, nil).nodePtr()
}

// parent returns the parent of the provided node pointer.
// Behavior is not well defined if the calling node pointer does not exist in the tree.
func (nodePtr *nodePtr) parent() *nodePtr {
	// See if there is a parent with the same 'key' as this node.
	parent := nodePtr.tree.nodePtrGet(nodePtr.level+1, nodePtr.key)
	if parent.exists() {
		return parent
	}
	// If not, take the node in the above layer that is lexicographically the closest
	// from the left of the key.
	parent = parent.leftSibling()
	if parent.exists() {
		return parent
	}
	// If there is no such node (this node is not in the tree), return nil
	return nodePtr.tree.nodePtrGet(nodePtr.level+1, nil)
}

// exists returns if the calling node is in the tree.
func (nodePtr *nodePtr) exists() bool {
	if nodePtr == nil {
		return false
	}
	return nodePtr.tree.store.Has(nodePtr.tree.nodeKey(nodePtr.level, nodePtr.key))
}

// updateAccumulation changes the accumulation value of a node in the tree,
// and handles updating the accumulation for all of its parent's augmented data.
func (nodePtr *nodePtr) updateAccumulation(c node) {
	if !nodePtr.exists() {
		return // reached at the root
	}

	children := nodePtr.children()
	idx, match := children.find(c.Index)
	if !match {
		panic("non existing key pushed from the child")
	}
	children = children.setAcc(idx, c.Acc)
	nodePtr.set(children)
	nodePtr.parent().updateAccumulation(node{nodePtr.key, children.accumulate()})
}

func (nodePtr *nodePtr) push(c node) {
	if !nodePtr.exists() {
		nodePtr.create(children{c})
		return
	}

	cs := nodePtr.children()
	idx, match := cs.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		nodePtr.updateAccumulation(c)
		return
	}

	// inserting new child nodePtr
	cs = cs.insert(idx, c)
	parent := nodePtr.parent()

	// split and push-up if overflow
	if len(cs) > int(nodePtr.tree.m) {
		split := nodePtr.tree.m/2 + 1
		leftchildren, rightchildren := cs.split(int(split))
		nodePtr.tree.nodePtrGet(nodePtr.level, cs[split].Index).create(rightchildren)
		if !parent.exists() {
			parent.create(children{
				node{nodePtr.key, leftchildren.accumulate()},
				node{cs[split].Index, rightchildren.accumulate()},
			})
			nodePtr.set(leftchildren)
			return
		}
		// constructing right child
		parent.push(node{cs[split].Index, rightchildren.accumulate()})
		cs = leftchildren
		parent = nodePtr.parent() // parent might be changed during the pushing process
	}

	parent.updateAccumulation(node{nodePtr.key, cs.accumulate()})
	nodePtr.set(cs)
}

func (nodePtr *nodePtr) pull(key []byte) {
	if !nodePtr.exists() {
		return // reached at the root
	}
	children := nodePtr.children()
	idx, match := children.find(key)

	if !match {
		panic("pulling non existing child")
	}

	children = children.delete(idx)
	// For sake of efficienty of our use case, we pull only when a nodePtr gets
	// empty.
	// if len(data.Index) >= int(nodePtr.tree.m/2) {
	if len(children) > 0 {
		nodePtr.set(children)
		nodePtr.parent().updateAccumulation(node{nodePtr.key, children.accumulate()})
		return
	}

	// merge if possible
	left := nodePtr.leftSibling()
	right := nodePtr.rightSibling()
	parent := nodePtr.parent()
	nodePtr.delete()
	parent.pull(nodePtr.key)

	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		if bytes.Equal(parent.key, right.parent().key) {
			leftchildren := left.children()
			rightchildren := right.children()
			if len(leftchildren)+len(rightchildren) < int(nodePtr.tree.m) {
				left.set(leftchildren.merge(rightchildren))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(node{left.key, leftchildren.accumulate()})
			}
		}
	}
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

func (nodePtr *nodePtr) create(children children) {
	keybz := nodePtr.tree.nodeKey(nodePtr.level, nodePtr.key)
	bz, err := json.Marshal(children)
	if err != nil {
		panic(err)
	}
	nodePtr.tree.store.Set(keybz, bz)
}
