package sumtree

import (
	"bytes"

	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewLeaf(key []byte, acc sdk.Int) *Leaf {
	return &Leaf{Leaf: &Child{
		Index:        key,
		Accumulation: acc,
	}}
}

func (ptr *ptr) isLeaf() bool {
	return ptr.level == 0
}

func (ptr *ptr) node() (res *Node) {
	res = new(Node)
	bz := ptr.tree.store.Get(ptr.tree.nodeKey(ptr.level, ptr.key))
	if bz != nil {
		if err := proto.Unmarshal(bz, res); err != nil {
			panic(err)
		}
	}
	return
}

func (ptr *ptr) set(node *Node) {
	bz, err := proto.Marshal(node)
	if err != nil {
		panic(err)
	}
	ptr.tree.store.Set(ptr.tree.nodeKey(ptr.level, ptr.key), bz)
}

func (ptr *ptr) setLeaf(leaf *Leaf) {
	if !ptr.isLeaf() {
		panic("setLeaf should only be called on pointers to leaf nodes. This ptr is a branch")
	}
	bz, err := proto.Marshal(leaf)
	if err != nil {
		panic(err)
	}
	ptr.tree.store.Set(ptr.tree.leafKey(ptr.key), bz)
}

func (ptr *ptr) delete() {
	ptr.tree.store.Delete(ptr.tree.nodeKey(ptr.level, ptr.key))
}

func (ptr *ptr) leftSibling() *ptr {
	iter := ptr.tree.ptrReverseIterator(ptr.level, nil, ptr.key)
	defer iter.Close()
	return iter.ptr()
}

func (ptr *ptr) rightSibling() *ptr {
	iter := ptr.tree.ptrIterator(ptr.level, ptr.key, nil)
	defer iter.Close()
	if !iter.Valid() {
		return nil
	}
	if ptr.exists() {
		// exclude ptr itself
		iter.Next()
	}
	return iter.ptr()
}

func (ptr *ptr) child(n uint16) *ptr {
	// TODO: set end to prefix iterator end
	iter := ptr.tree.ptrIterator(ptr.level-1, ptr.node().Children[n].Index, nil)
	defer iter.Close()
	return iter.ptr()
}

// parent returns the parent of the provided pointer.
// Behavior is not well defined if the calling pointer does not exist in the tree.
func (ptr *ptr) parent() *ptr {
	// See if there is a parent with the same 'key' as this ptr.
	parent := ptr.tree.ptrGet(ptr.level+1, ptr.key)
	if parent.exists() {
		return parent
	}
	// If not, take the node in the above layer that is lexicographically the closest
	// from the left of the key.
	parent = parent.leftSibling()
	if parent.exists() {
		return parent
	}
	// If there is no such ptr (the parent is not in the tree), return nil
	return ptr.tree.ptrGet(ptr.level+1, nil)
}

// exists returns true if the calling pointer has a node in the tree.
func (ptr *ptr) exists() bool {
	if ptr == nil {
		return false
	}
	return ptr.tree.store.Has(ptr.tree.nodeKey(ptr.level, ptr.key))
}

// updateAccumulation changes the accumulation value of a ptr in the tree,
// and handles updating the accumulation for all of its parent's augmented data.
func (ptr *ptr) updateAccumulation(c *Child) {
	if !ptr.exists() {
		return // reached at the root
	}

	node := ptr.node()
	idx, match := node.find(c.Index)
	if !match {
		panic("non existing key pushed from the child")
	}
	node = node.setAcc(idx, c.Accumulation)
	ptr.set(node)
	ptr.parent().updateAccumulation(&Child{ptr.key, node.accumulate()})
}

func (ptr *ptr) push(c *Child) {
	if !ptr.exists() {
		ptr.create(NewNode(c))
		return
	}

	cs := ptr.node()
	idx, match := cs.find(c.Index)

	// setting already existing child, move to updateAccumulation
	if match {
		ptr.updateAccumulation(c)
		return
	}

	// inserting new child ptr
	cs = cs.insert(idx, c)
	parent := ptr.parent()

	// split and push-up if overflow
	if len(cs.Children) > int(ptr.tree.m) {
		split := ptr.tree.m/2 + 1
		leftnode, rightnode := cs.split(int(split))
		ptr.tree.ptrGet(ptr.level, cs.Children[split].Index).create(rightnode)
		if !parent.exists() {
			parent.create(NewNode(
				&Child{ptr.key, leftnode.accumulate()},
				&Child{cs.Children[split].Index, rightnode.accumulate()},
			))
			ptr.set(leftnode)
			return
		}
		// constructing right child
		parent.push(&Child{cs.Children[split].Index, rightnode.accumulate()})
		cs = leftnode
		parent = ptr.parent() // parent might be changed during the pushing process
	}

	parent.updateAccumulation(&Child{ptr.key, cs.accumulate()})
	ptr.set(cs)
}

func (ptr *ptr) pull(key []byte) {
	if !ptr.exists() {
		return // reached at the root
	}
	node := ptr.node()
	idx, match := node.find(key)

	if !match {
		panic("pulling non existing child")
	}

	node = node.delete(idx)
	// For sake of efficiently on our use case, we pull only when a ptr gets
	// empty.
	// if len(data.Index) >= int(ptr.tree.m/2) {
	if len(node.Children) > 0 {
		ptr.set(node)
		ptr.parent().updateAccumulation(&Child{ptr.key, node.accumulate()})
		return
	}

	// merge if possible
	left := ptr.leftSibling()
	right := ptr.rightSibling()
	parent := ptr.parent()
	ptr.delete()
	parent.pull(ptr.key)

	if left.exists() && right.exists() {
		// parent might be deleted, retrieve from left
		parent = left.parent()
		if bytes.Equal(parent.key, right.parent().key) {
			leftnode := left.node()
			rightnode := right.node()
			if len(leftnode.Children)+len(rightnode.Children) < int(ptr.tree.m) {
				left.set(leftnode.merge(rightnode))
				right.delete()
				parent.pull(right.key)
				parent.updateAccumulation(&Child{left.key, leftnode.accumulate()})
			}
		}
	}
}

func (node Node) accumulate() (res sdk.Int) {
	res = sdk.ZeroInt()
	for _, child := range node.Children {
		res = res.Add(child.Accumulation)
	}
	return
}

func NewNode(cs ...*Child) *Node {
	return &Node{Children: cs}
}

// find returns the appropriate position that key should be inserted
// if match is true, idx is the exact position for the key
// if match is false, idx is the position where the key should be inserted.
func (node Node) find(key []byte) (idx int, match bool) {
	for idx, child := range node.Children {
		if bytes.Equal(child.Index, key) {
			return idx, true
		}
		// Push new key to the appropriate position
		if bytes.Compare(child.Index, key) > 0 {
			return idx, false
		}
	}

	return len(node.Children), false
}

func (node *Node) setAcc(idx int, acc sdk.Int) *Node {
	node.Children[idx] = &Child{node.Children[idx].Index, acc}
	return node
}

func (node *Node) insert(idx int, c *Child) *Node {
	arr := append(node.Children[:idx], append([]*Child{c}, node.Children[idx:]...)...)
	return NewNode(arr...)
}

func (node *Node) delete(idx int) *Node {
	node = NewNode(append(node.Children[:idx], node.Children[idx+1:]...)...)
	return node
}

func (node *Node) split(idx int) (*Node, *Node) {
	return NewNode(node.Children[:idx]...), NewNode(node.Children[idx:]...)
}

func (node *Node) merge(node2 *Node) *Node {
	return NewNode(append(node.Children, node2.Children...)...)
}
