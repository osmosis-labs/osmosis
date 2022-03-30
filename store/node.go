package store

import (
	"bytes"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (node *node) isLeaf() bool {
	return node.level == 0
}

func (node *node) children() (res children) {
	bz := node.tree.store.Get(node.tree.nodeKey(node.level, node.key))
	if bz != nil {
		err := json.Unmarshal(bz, &res)
		if err != nil {
			panic(err)
		}
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

func (node *node) setLeaf(acc sdk.Int) {
	if !node.isLeaf() {
		panic("setLeaf should not be called on branch node")
	}
	bz, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}
	node.tree.store.Set(node.tree.leafKey(node.key), bz)
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
	if node.exists() {
		// exclude node itself
		iter.Next()
	}
	return iter.node()
}

func (node *node) child(n uint16) *node {
	// TODO: set end to prefix iterator end
	return node.tree.nodeIterator(node.level-1, node.children()[n].Index, nil).node()
}

// parent returns the parent of the provided node pointer.
// Behavior is not well defined if the calling node pointer does not exist in the tree.
func (node *node) parent() *node {
	// See if there is a parent with the same 'key' as this node.
	parent := node.tree.nodeGet(node.level+1, node.key)
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
	return node.tree.nodeGet(node.level+1, nil)
}

// exists returns if the calling node is in the tree.
func (node *node) exists() bool {
	if node == nil {
		return false
	}
	return node.tree.store.Has(node.tree.nodeKey(node.level, node.key))
}

// updateAccumulation changes the accumulation value of a node in the tree,
// and handles updating the accumulation for all of its parent's augmented data.
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

	// inserting new child node
	cs = cs.insert(idx, c)
	parent := node.parent()

	// split and push-up if overflow
	if len(cs) > int(node.tree.m) {
		split := node.tree.m/2 + 1
		leftchildren, rightchildren := cs.split(int(split))
		node.tree.nodeGet(node.level, cs[split].Index).create(rightchildren)
		if !parent.exists() {
			parent.create(children{
				child{node.key, leftchildren.accumulate()},
				child{cs[split].Index, rightchildren.accumulate()},
			})
			node.set(leftchildren)
			return
		}
		// constructing right childdd
		parent.push(child{cs[split].Index, rightchildren.accumulate()})
		cs = leftchildren
		parent = node.parent() // parent might be changed during the pushing process
	}

	parent.updateAccumulation(child{node.key, cs.accumulate()})
	node.set(cs)
}

func (node *node) pull(key []byte) {
	if !node.exists() {
		return // reached at the root
	}
	children := node.children()
	idx, match := children.find(key)

	if !match {
		panic("pulling non existing child")
	}

	children = children.delete(idx)
	// For sake of efficiently on our use case, we pull only when a node gets
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
			leftchildren := left.children()
			rightchildren := right.children()
			if len(leftchildren)+len(rightchildren) < int(node.tree.m) {
				left.set(leftchildren.merge(rightchildren))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(child{left.key, leftchildren.accumulate()})
			}
		}
	}
}

func (children children) accumulate() (res sdk.Int) {
	res = sdk.ZeroInt()
	for _, child := range children {
		res = res.Add(child.Acc)
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

func (children children) set(idx int, child child) children {
	children[idx] = child
	return children
}

func (children children) setAcc(idx int, acc sdk.Int) children {
	children[idx] = child{children[idx].Index, acc}
	return children
}

func (cs children) insert(idx int, c child) children {
	return append(cs[:idx], append(children{c}, cs[idx:]...)...)
}

// delete removes the corresponding node from the underlying data store,
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
