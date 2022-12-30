package v101_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/iavl"

	dbm "github.com/tendermint/tm-db"

	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/sumtree"
	v101 "github.com/osmosis-labs/osmosis/osmoutils/sumtree/legacy/v101"
)

func setupStore() sdk.KVStore {
	db := dbm.NewMemDB()
	tree, _ := iavl.NewMutableTree(db, 100, false)
	_, _, err := tree.SaveVersion()
	if err != nil {
		panic(err)
	}
	kvstore := iavlstore.UnsafeNewStore(tree)
	return kvstore
}

func compareBranch(oldValueBz []byte, valueBz []byte) (err error) {
	oldValue := v101.Children{}
	value := sumtree.Node{}
	err = json.Unmarshal(oldValueBz, &oldValue)
	if err != nil {
		return
	}
	err = proto.Unmarshal(valueBz, &value)
	if err != nil {
		return
	}

	for i, c := range oldValue {
		c2 := value.Children[i]
		if !bytes.Equal(c.Index, c2.Index) || !c.Acc.Equal(c2.Accumulation) {
			err = fmt.Errorf("branch value mismatch: %+v / %+v", oldValue, value)
			return
		}
	}
	return
}

func compareLeaf(oldValueBz []byte, valueBz []byte) (err error) {
	oldValue := sdk.ZeroInt()
	value := sumtree.Leaf{}
	err = json.Unmarshal(oldValueBz, &oldValue)
	if err != nil {
		return
	}
	err = proto.Unmarshal(valueBz, &value)
	if err != nil {
		return
	}

	if !oldValue.Equal(value.Leaf.Accumulation) {
		return fmt.Errorf("leaf value mismatch: %+v / %+v", oldValue, value)
	}
	return
}

func comparePair(oldKeyBz, oldValueBz, keyBz, valueBz []byte) (err error) {
	if !bytes.Equal(oldKeyBz, keyBz) {
		err = fmt.Errorf("key bytes mismatch: %x / %x", oldKeyBz, keyBz)
	}

	// TODO: properly select error
	err = compareBranch(oldValueBz, valueBz)
	if err == nil {
		return nil
	}
	err = compareLeaf(oldValueBz, valueBz)
	return err
}

type kvPair struct {
	key   []byte
	value []byte
}

func pair(iter sdk.Iterator) kvPair {
	res := kvPair{iter.Key(), iter.Value()}
	iter.Next()
	return res
}

func extract(store sdk.KVStore) (res []kvPair) {
	res = []kvPair{}
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for iter.Valid() {
		res = append(res, pair(iter))
	}
	return
}

func readold() []kvPair {
	bz, err := os.ReadFile("./old_tree.json")
	if err != nil {
		panic(err)
	}
	var data [][][]byte
	err = json.Unmarshal(bz, &data)
	if err != nil {
		panic(err)
	}
	res := make([]kvPair, len(data))
	for i, pair := range data {
		res[i] = kvPair{pair[0], pair[1]}
	}
	return res
}

func TestMigrate(t *testing.T) {
	store := setupStore()

	oldpairs := readold()
	for _, pair := range oldpairs {
		fmt.Println("set", pair.key, pair.value)
		store.Set(pair.key, pair.value)
	}

	v101.MigrateTree(store)

	newpairs := extract(store)

	for i, oldpair := range oldpairs {
		fmt.Println(i)
		newpair := newpairs[i]
		err := comparePair(oldpair.key, oldpair.value, newpair.key, newpair.value)
		require.NoError(t, err)
	}
}
