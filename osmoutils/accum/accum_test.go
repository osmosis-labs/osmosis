package accum_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"

	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
)

type AccumTestSuite struct {
	suite.Suite

	store store.KVStore
}

// Sets/resets KVStore to use for tests under `suite.store`
func (suite *AccumTestSuite) SetupTest() {
	db := dbm.NewMemDB()
	tree, err := iavl.NewMutableTree(db, 100, false)
	suite.Require().NoError(err)
	_, _, err = tree.SaveVersion()
	suite.Require().Nil(err)
	kvstore := iavlstore.UnsafeNewStore(tree)
	suite.store = kvstore
}

func TestTreeTestSuite(t *testing.T) {
	suite.Run(t, new(AccumTestSuite))
}

func (suite *AccumTestSuite) TestMakeAndGetAccum() {
	// We set up store once at beginning so we can test duplicates
	suite.SetupTest()

	type testcase struct {
		accumName string
		expAccum accum.AccumulatorObject
		expSetPass	bool
		expGetPass	bool
	}

	tests := map[string]testcase{
		"create valid accumulator": {
			accumName: "fee-accumulator",
			expSetPass: true,
			expGetPass: true,
		},
		"create duplicate accumulator": {
			accumName: "fee-accumulator",
			expSetPass: false,
			expGetPass: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			expAccum := accum.AccumulatorObject{
				Store: suite.store,
				Name: tc.accumName,
				Value: sdk.ZeroDec(),
			}

			err := accum.MakeAccumulator(suite.store, tc.accumName)

			if !tc.expSetPass {
				suite.Require().Error(err)
			}

			actualAccum, err := accum.GetAccumulator(suite.store, tc.accumName)

			if tc.expGetPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expAccum, actualAccum)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}