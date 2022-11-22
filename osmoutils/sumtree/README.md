# Prefix-Sum B-Tree specification

This module implements a B-Tree suitable for efficiently computing a
random prefix sum of data, while allowing the data to be efficiently
updated.

The prefix sums for N elements x\_1, x\_2, ... x\_N, each with a weight
field, is the sequence y\_1, y\_2, ... y\_N, where
`y_i = sum_{0 <= j <= i} x_j.weight`. This data structure allows one to
edit, insert, and delete entries in the x sequence efficiently, and
efficiently retrieve the prefix sum at any index, where efficiently is
`O(log(N))` state operations. (Note that in the cosmos SDK stack, a
state operation is itself liable to take `O(log(N))` time)

This is built for the use-case of we have a series of data that is
sorted by time. Each given time has an associated weight field. We want
to be able to very quickly find the total weight for all leaves with
times less than or equal to `t`. The actual implementation is agnostic
to what is the field we sort by.

## Data structure idea

The idea underlying this can be decomposed into two parts:

1. Do some extra (`O(log(N))`) work when modifying the data, to allow
    efficiently computing any prefix sum
2. Allow the data entries to be inserted into and deleted from, while
    remaining sorted

The solution for 1. is to build a balanced tree on top of the data.
Every inner node in the tree will be augmented with an "accumulated
weight" field, which contains the sum of the weights of all entries
below it. Notice that upon updating the weight of any leaf, the weights
of all nodes that are "above" this node can be updated efficiently. (As
there ought to only be log(N) such nodes) Furthermore, the root of the
tree's augmented value is the sum of all weights in the tree.

Then to query the jth prefix sum, you first identify the path to the jth
node in the tree. You keep a running tally of "prefix sum thus far",
which is initialized to the sum of all weights in the tree. Then as you
walk the path from the tree root to the jth leaf, whenever there are
siblings on the right, you subtract their weight from your running
total. The weight when you arrive at the leaf is then the prefix sum.

Lets illustrate this with a binary tree.

![binary
tree](https://user-images.githubusercontent.com/6440154/116960474-142bf980-ac66-11eb-9a07-af84ab6d0bfa.png)

If we want the prefix sum for leaf `12` we compute it as
`1.weight - 7.weight - 13.weight`. (We took a subtraction every time we
took a left)

Now notice that this solution works for any tree type, where efficiency
just depends on `number of siblings * depth` and as long as that is
parameterized to be `O(log(N))` we maintain our definition of efficient.

Thus we immediately get a solution to 2., by using a tree that supports
efficient inserts, and deletions, while still maintaining log depth and
a constant number of siblings. We opt for using a B+ tree for this, as
it performs the task well and does not require rebalance operations.
(Which can be costly in an adversarial environment)

```{=html}
<!---
TODO: Improve diagrams showing this accumulated weight concept with a binary tree, and show how you query a random prefix sum.
-->
```

## Implementation Details

The B-Tree implementation under `osmoutils/sumtree` is designed specifically
for allowing efficient computation of a random prefix sum, with the
underlying data being updatable as explained above.

Every Leaf has a `Weight` field, and the address its stored at in state
is the key which we want to sort by. The implementation sorts leaves as
byteslices, Leafs are sorted under their byteslice key, and the branch
nodes have accumulation for each childs.

A node is pointed by a `node` struct, used internally.

``` {.go}
type node struct {
    t Tree
    level uint8
    key []byte
}
```

A `node` struct is a pointer to the key-value pair under
`nodeKey(node.level, node.key)`

A leaf node is simply a `uint64` integer value stored under
`nodeKey(0, key)`. The key is arbitrary length of byte slice.

``` {.go}
type Leaf struct {
    Value uint64
}
```

A branch node consists of keys and accumulation of the children nodes.

``` {.go}
type Branch struct {
    Children []Child    
}

type Child struct {
    Key []byte
    Acc uint64
}
```

The following constraints are valid for all branch nodes:

1. For `c` in `node.Branch().Children`, the node corresponding to `c`
    is stored under `nodeKey(node.level-1, c.Key)`.
2. For `c` in `node.Branch().Children`, `c.Acc` is the sum of all
    `c'.Acc` where `c'` is in `c.Children`. If `c'` is leaf node,
    substitute `c'.Acc` to `Leaf.Value`.
3. For `c` in `node.Branch().Children`, `c.Key` is equal or greater
    than `node.key` and lesser than `node.rightSibling().key`.
4. There are no duplicate child stored in more than one of node's
    `.Children`.

### Example

Here is an example tree data:

    - Level 2 nil
      - Level 1 0xaaaa 
        - Level 0 0xaaaa Value 10
        - Level 0 0xaaaa01 Value 20
        - Level 0 0xaabb Value 30
      - Level 1 0xbb44
        - Level 0 0xbb55 Value 100
        - Level 0 0xbe Value 200
      - Level 1 0xeeaaaa
        - Level 0 0xef1234 Value 300
        - Level 0 0xffff Value 400

The branch nodes will have the following childrens:

``` {.go}
require.Equal(sumtree.Get(nodeKey(2, nil)), Children{{0xaaaa, 60}, {0xbb44, 300}, {0xeeaaaa, 700}})
require.Equal(sumtree.Get(nodeKey(1, 0xaaaa)), Children{{0xaaaa, 10}, {0xaaaa01, 20}, {0xaabb, 30}})
require.Equal(sumtree.Get(nodeKey(1, 0xbb44)), Children{{0xbb55, 100}, {0xbe, 200}})
require.Equal(sumtree.Get(nodeKey(1, 0xeeaaaa)), Children{{0xef1234, 300}, {0xffff, 400}})
```
