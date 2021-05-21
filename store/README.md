# Modified B-Tree specification

## Implementation Details

B-Tree implementation under `osmosis/store` is designed specifically for fast-calculating subslice deposit accumulation. Leafs are sorted under their byteslice key, and the branch nodes have accumulation for each childs.

A node is pointed by a `node` struct, used internally.

```go=
type node struct {
    t Tree
    level uint8
    key []byte
}
```

A `node` struct is a pointer to the key-value pair under `nodeKey(node.level, node.key)`

A leaf node is simply a `uint64` integer value stored under `nodeKey(0, key)`. The key is arbitrary length of byte slice. 

```go=
type Leaf struct {
    Value uint64
}
```

A branch node consists of keys and accumulation of the children nodes.

```go=
type Branch struct {
    Children []Child    
}

type Child struct {
    Key []byte
    Acc uint64
}
```

The following constraints are valid for all branch nodes:

1. For `c` in `node.Branch().Children`, the node corresponding to `c` is stored under `nodeKey(node.level-1, c.Key)`.
2. For `c` in `node.Branch().Children`, `c.Acc` is the sum of all `c'.Acc` where `c'` is in `c.Children`. If `c'` is leaf node, substitute `c'.Acc` to `Leaf.Value`.
3. For `c` in `node.Branch().Children`, `c.Key` is equal or greater than `node.key` and lesser than `node.rightSibling().key`.
4. There are no duplicate child stored in more than one of node's `.Children`.

## Example

Here is an example tree data:

```
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
```

The branch nodes will have the following childrens:
```go=
require.Equal(store.Get(nodeKey(2, nil)), Children{{0xaaaa, 60}, {0xbb44, 300}, {0xeeaaaa, 700}})
require.Equal(store.Get(nodeKey(1, 0xaaaa)), Children{{0xaaaa, 10}, {0xaaaa01, 20}, {0xaabb, 30}})
require.Equal(store.Get(nodeKey(1, 0xbb44)), Children{{0xbb55, 100}, {0xbe, 200}})
require.Equal(store.Get(nodeKey(1, 0xeeaaaa)), Children{{0xef1234, 300}, {0xffff, 400}})
```