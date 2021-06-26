package v101_test

import (
	"encoding/json"
	"fmt"
	"bytes"
	"testing"
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/store/legacy/v101"
	"github.com/osmosis-labs/osmosis/store"
)

func compareBranch(oldValue v101.Children, value store.Node) (ok bool, err error) {
	for i, c := range oldValue {
		c2 := value.Children[i]
		if !bytes.Equal(c.Index, c2.Index) || !c.Acc.Equal(c2.Accumulation) {
			return
		}
	}
	ok = true
	return
}

func compareLeaf(oldValue sdk.Int, value store.Leaf) (ok bool, err error) {
	if !oldValue.Equal(value.Leaf.Accumulation) {
		return
	}
	ok = true
	return
}

func comparePair(isLeaf bool, oldKeyBz, oldValueBz, keyBz, valueBz []byte) (err error) {
	if !bytes.Equal(oldKeyBz, keyBz) {
		return fmt.Errorf("key bytes mismatch: %x / %x", oldKeyBz, keyBz)
	}
	if isLeaf {
		oldValue := sdk.ZeroInt()
		value := store.Leaf{}
		err = json.Unmarshal(oldValueBz, &oldValue)
		if err != nil {
			return
		}
		err = proto.Unmarshal(valueBz, &value)
		if err != nil {
			return
		}
		ok, err := compareLeaf(oldValue, value)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("leaf value mismatch: %+v / %+v", oldValue, value)
		}
	} else {
		oldValue := v101.Children{}
		value := store.Node{}
		err = json.Unmarshal(oldValueBz, &oldValue)
		if err != nil {
			return
		}
		err = proto.Unmarshal(valueBz, &value)
		if err != nil {
			return
		}

		ok, err := compareBranch(oldValue, value)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("branch value mismatch: %+v / %+v", oldValue, value)
		}
	}

	return nil
}

type kvPair struct {
	key []byte
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

func testTree() {
	
}

func TestMigrate(t *testing.T) {
	
}
