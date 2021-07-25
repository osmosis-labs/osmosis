package v101

import (
  "github.com/gogo/protobuf/proto"

  sdk "github.com/cosmos/cosmos-sdk/types"

  "github.com/osmosis-labs/osmosis/store"
)

func parseKey(key []byte) (duration []byte) {
  duration = key[:8]
  return
}

func extract(denomAccStore sdk.KVStore) map[string]sdk.Int {
  amountPerDuration := make(map[string]sdk.Int)

  // denomAccStore should be already denom prefixed 
  tree := store.NewTree(denomAccStore, 10)

  iter := tree.Iterator(nil, nil)
  defer iter.Close()
  for ; iter.Valid(); iter.Next() {
    key, valuebz := iter.Key(), iter.Value()
    // ignore nil placeholder
    if key == nil {
      continue
    }
    duration := string(parseKey(key))
    amt, ok := amountPerDuration[duration]
    if !ok {
      amt = sdk.ZeroInt()
    }

    leaf := new(store.Leaf)
    err := proto.Unmarshal(valuebz, leaf)
    if err != nil {
      panic(err)
    }
    value := leaf.Leaf.Accumulation
    amt = amt.Add(value)

    amountPerDuration[duration] = amt
  }

  return amountPerDuration
}

func wipe(denomAccStore sdk.KVStore) {
  iter := denomAccStore.Iterator(nil, nil)
  defer iter.Close()

  for ; iter.Valid(); iter.Next() {
    denomAccStore.Delete(iter.Key())
  }
}

func override(denomAccStore sdk.KVStore, amountPerDuration map[string]sdk.Int) {
  tree := store.NewTree(denomAccStore, 10)
  for key, value := range amountPerDuration {
    tree.Set([]byte(key), value)
  }
}

func MigrateAccStore(denomAccStore sdk.KVStore) {
  amountPerDuration := extract(denomAccStore)
  wipe(denomAccStore)
  override(denomAccStore, amountPerDuration)
}
