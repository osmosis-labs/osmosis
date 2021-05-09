package store_test

import (
	"bytes"
	"math/rand"
	"testing"
	"sort"

	"github.com/stretchr/testify/suite"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"

	"github.com/c-osmosis/osmosis/store"
)

type TreeTestSuite struct {
	suite.Suite

	tree store.Tree
}

func (suite *TreeTestSuite) SetupTest() {
	kvstore := dbadapter.Store{DB: dbm.NewMemDB()}
	suite.tree = store.NewTree(kvstore, 10)
}

func TestTreeTestSuite(t *testing.T) {
	suite.Run(t, new(TreeTestSuite))
}

type pair struct {
	key []byte
	value []byte
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

func (suite *TreeTestSuite) TestTreeInvariants() {
	suite.SetupTest()

	pairs := pairs{pair{[]byte("hello"), []byte("world")}}
	suite.tree.Set([]byte("hello"), []byte("world"))

	for i := 0; i < 2000; i++ {
		// add a single element
		key := make([]byte, rand.Int()%20)
		value := make([]byte, rand.Int()%20)
		rand.Read(key)
		rand.Read(value)
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
			suite.Require().Equal(suite.tree.Get(pair.key), pair.value)
			// XXX: check all branch nodes
		}

		// XXX: remove randomly by coin flip
		if rand.Int() % 2 == 0 {
			idx := rand.Int()%len(pairs)
			pair := pairs[idx]
			pairs = append(pairs[:idx], pairs[idx+1:]...)
			suite.tree.Remove(pair.key)
		}
	}
}
