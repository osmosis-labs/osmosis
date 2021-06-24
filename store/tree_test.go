package store_test

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/iavl"

	dbm "github.com/tendermint/tm-db"

	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/store"
)

type TreeTestSuite struct {
	suite.Suite

	tree store.Tree
}

func (suite *TreeTestSuite) SetupTest() {
	db := dbm.NewMemDB()
	tree, err := iavl.NewMutableTree(db, 100)
	suite.Require().NoError(err)
	_, _, err = tree.SaveVersion()
	suite.Require().Nil(err)
	kvstore := iavlstore.UnsafeNewStore(tree)
	suite.tree = store.NewTree(kvstore, 10)
}

func TestTreeTestSuite(t *testing.T) {
	suite.Run(t, new(TreeTestSuite))
}

type pair struct {
	key   []byte
	value sdk.Coins
}

type pairs []pair

var _ sort.Interface = pairs{}

func (p pairs) Len() int {
	return len(p)
}

func (p pairs) Less(i, j int) bool {
	return bytes.Compare(p[i].key, p[j].key) < 0
}

func (p pairs) Swap(i, j int) {
	temp := p[i]
	p[i] = p[j]
	p[j] = temp
}

func (p pairs) sum() (res sdk.Coins) {
	for _, pair := range p {
		res = res.Add(pair.value...)
	}
	return
}

func (suite *TreeTestSuite) TestTreeInvariants() {
	suite.SetupTest()

	denom := "denom"

	pairs := pairs{pair{[]byte("hello"), sdk.NewCoins(sdk.NewInt64Coin(denom, 100))}}
	suite.tree.Set([]byte("hello"), sdk.NewCoins(sdk.NewInt64Coin(denom, 100)))

	// tested up to 2000
	for i := 0; i < 500; i++ {
		// add a single element
		key := make([]byte, rand.Int()%20)
		value := sdk.NewCoins(sdk.NewInt64Coin(denom, rand.Int63()%100))
		rand.Read(key)
		idx := sort.Search(len(pairs), func(n int) bool { return bytes.Compare(pairs[n].key, key) >= 0 })
		if idx < len(pairs) {
			if bytes.Equal(pairs[idx].key, key) {
				pairs[idx] = pair{key, value}
			} else {
				pairs = append(pairs, pair{key, value})
				sort.Sort(pairs)
			}
		} else {
			pairs = append(pairs, pair{key, value})
		}

		suite.tree.Set(key, value)

		// check all is right
		for _, pair := range pairs {
			acc := suite.tree.Get(pair.key).Leaf.Accumulation
			suite.Require().True(acc.IsEqual(pair.value),
				"stored accumulation %v differ from %v", acc, pair.value)
			// XXX: check all branch nodes
		}

		// check accumulation calc is alright
		left, exact, right := sdk.Coins{}, pairs[0].value, pairs[1:].sum()
		for idx, pair := range pairs {
			tleft, texact, tright := suite.tree.SplitAcc(pair.key)
			suite.Require().True(left.IsEqual(tleft))
			suite.Require().True(exact.IsEqual(texact))
			suite.Require().True(right.IsEqual(tright))

			key := append(pair.key, 0x00)
			if idx == len(pairs)-1 {
				break
			}
			if bytes.Equal(key, pairs[idx+1].key) {
				break
			}

			tleft, texact, tright = suite.tree.SplitAcc(key)
			suite.Require().True(left.Add(exact...).IsEqual(tleft))
			suite.Require().True(sdk.Coins{}.IsEqual(texact))
			suite.Require().True(right.IsEqual(tright))

			left = left.Add(exact...)
			exact = pairs[idx+1].value
			right = right.Sub(exact)
		}

		if rand.Int()%2 == 0 {
			idx := rand.Int() % len(pairs)
			pair := pairs[idx]
			pairs = append(pairs[:idx], pairs[idx+1:]...)
			suite.tree.Remove(pair.key)
		}
	}
}
