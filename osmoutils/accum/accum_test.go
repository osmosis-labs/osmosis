package accum_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils"
	accumPackage "github.com/osmosis-labs/osmosis/osmoutils/accum"

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
		accumName  string
		expAccum   accumPackage.AccumulatorObject
		expSetPass bool
		expGetPass bool
	}

	tests := map[string]testcase{
		"create valid accumulator": {
			accumName:  "fee-accumulator",
			expSetPass: true,
			expGetPass: true,
		},
		"create duplicate accumulator": {
			accumName:  "fee-accumulator",
			expSetPass: false,
			expGetPass: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// Creates raw accumulator object with test case's accum name and zero initial value
			expAccum := accumPackage.CreateRawAccumObject(suite.store, tc.accumName, sdk.DecCoins(nil))

			err := accumPackage.MakeAccumulator(suite.store, tc.accumName)

			if !tc.expSetPass {
				suite.Require().Error(err)
			}

			actualAccum, err := accumPackage.GetAccumulator(suite.store, tc.accumName)

			if tc.expGetPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expAccum, actualAccum)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AccumTestSuite) TestAddToPosition() {
	type testcase struct {
		startingNumShares sdk.Dec
		startingUnclaimedRewards sdk.DecCoins
		newShares   sdk.Dec

		// accumInit and accumDelta specify the initial accum value 
		// and how much it has changed since the position being added 
		// to was created
		accumInit	sdk.DecCoins
		accumDelta  sdk.DecCoins

		// Address does not exist
		addrDNE bool
		expPass bool
	}

	tests := map[string]testcase{
		"zero shares with no new rewards": {
			startingNumShares: sdk.ZeroDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: true,
		},
		"non-zero shares with no new rewards": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: true,
		},
		"non-zero shares with new rewards in one denom": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unclaimed rewards since last update
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
			expPass: true,
		},
		"non-zero shares with new rewards in two denoms": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			expPass: true,
		},
		"non-zero shares with both existing and new rewards": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(11)), sdk.NewDecCoin("bar", sdk.NewInt(11))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			expPass: true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two denoms)": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			expPass: true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two new denoms)": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(10)), sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with same denoms": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with new denoms": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(10))),
			accumDelta: sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(10)), sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},

		// error catching
		"account does not exist": {
			addrDNE: true,
			expPass: false,

			startingNumShares: sdk.ZeroDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
		},
		"attempt to add zero shares": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.ZeroDec(),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: false,
		},
		"attempt to add negative shares": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec().Neg(),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: false,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// We reset the store for each test
			suite.SetupTest()
			addr := osmoutils.CreateRandomAccounts(1)[0]

			// Create a new accumulator with initial value specified by test case
			curAccum := accumPackage.CreateRawAccumObject(suite.store, "test-accum", tc.accumInit)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDNE {
				accumPackage.CreateRawPosition(curAccum, addr, tc.startingNumShares, tc.startingUnclaimedRewards, accumPackage.PositionOptions{})
			}

			// Update accumulator with accumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.CreateRawAccumObject(suite.store, "test-accum", tc.accumInit.Add(tc.accumDelta...))

			// Add newShares to position
			err := curAccum.AddToPosition(addr, tc.newShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(accumPackage.GetStore(curAccum), addr)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.accumDelta...), newPosition.InitAccumValue)

				// Ensure accrued rewards are moved into UnclaimedRewards (both when it starts empty and not)
				// Note: assumes only one position for accumulator, so new unclaimed rewards = accumDelta
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.accumDelta...), newPosition.UnclaimedRewards)

				// Ensure address's position properly reflects new number of shares
				suite.Require().Equal(tc.startingNumShares.Add(tc.newShares), newPosition.NumShares)

				// TODO: ensure a new position isn't created in memory (only old one is overwritten)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
