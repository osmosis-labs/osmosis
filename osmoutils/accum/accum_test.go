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

var (
	testAddressOne   = sdk.AccAddress([]byte("addr1_______________"))
	testAddressTwo   = sdk.AccAddress([]byte("addr2_______________"))
	testAddressThree = sdk.AccAddress([]byte("addr3_______________"))

	emptyPositionOptions = accumPackage.PositionOptions{}
	testNameOne          = "myaccumone"
	testNameTwo          = "myaccumtwo"
	testNameThree        = "myaccumthree"
	denomOne             = "denomone"
	denomTwo             = "denomtwo"

	emptyCoins = sdk.DecCoins(nil)

	initialValueOne      = sdk.MustNewDecFromStr("100.1")
	initialCoinDenomOne  = sdk.NewDecCoinFromDec(denomOne, initialValueOne)
	initialCoinDenomTwo  = sdk.NewDecCoinFromDec(denomTwo, initialValueOne)
	initialCoinsDenomOne = sdk.NewDecCoins(initialCoinDenomOne)

	positionOne = accumPackage.Record{
		NumShares:        sdk.NewDec(100),
		InitAccumValue:   emptyCoins,
		UnclaimedRewards: emptyCoins,
	}

	positionOneV2 = accumPackage.Record{
		NumShares:        sdk.NewDec(150),
		InitAccumValue:   emptyCoins,
		UnclaimedRewards: emptyCoins,
	}

	positionTwo = accumPackage.Record{
		NumShares:        sdk.NewDec(200),
		InitAccumValue:   emptyCoins,
		UnclaimedRewards: emptyCoins,
	}

	positionThree = accumPackage.Record{
		NumShares:        sdk.NewDec(300),
		InitAccumValue:   emptyCoins,
		UnclaimedRewards: emptyCoins,
	}
)

func withInitialAccumValue(record accumPackage.Record, initialAccum sdk.DecCoins) accumPackage.Record {
	record.InitAccumValue = initialAccum
	return record
}

func withUnclaimedRewards(record accumPackage.Record, unclaimedRewards sdk.DecCoins) accumPackage.Record {
	record.UnclaimedRewards = unclaimedRewards
	return record
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
		tc := tc
		suite.Run(name, func() {
			// Creates raw accumulator object with test case's accum name and zero initial value
			expAccum := accumPackage.CreateRawAccumObject(suite.store, tc.accumName, emptyCoins)

			err := accumPackage.MakeAccumulator(suite.store, tc.accumName)

			if !tc.expSetPass {
				suite.Require().Error(err)
			}

			retrievedAccum, err := accumPackage.GetAccumulator(suite.store, tc.accumName)

			if tc.expGetPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expAccum, retrievedAccum)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AccumTestSuite) TestNewPosition() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	accObject := accumPackage.CreateRawAccumObject(suite.store, testNameOne, emptyCoins)

	tests := map[string]struct {
		accObject        accumPackage.AccumulatorObject
		addr             sdk.AccAddress
		numShareUnits    sdk.Dec
		options          accumPackage.PositionOptions
		expectedPosition accumPackage.Record
	}{
		"test address one - position created": {
			accObject:        accObject,
			addr:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			options:          emptyPositionOptions,
			expectedPosition: positionOne,
		},
		"test address two - position created": {
			accObject:        accObject,
			addr:             testAddressTwo,
			numShareUnits:    positionTwo.NumShares,
			options:          emptyPositionOptions,
			expectedPosition: positionTwo,
		},
		"test address one - position overwritten": {
			accObject:        accObject,
			addr:             testAddressOne,
			numShareUnits:    positionOneV2.NumShares,
			options:          emptyPositionOptions,
			expectedPosition: positionOneV2,
		},
		"test address three - added": {
			accObject:        accObject,
			addr:             testAddressThree,
			numShareUnits:    positionThree.NumShares,
			options:          emptyPositionOptions,
			expectedPosition: positionThree,
		},
		"test address one with non-empty accumulator - position created": {
			accObject:        accumPackage.CreateRawAccumObject(suite.store, testNameTwo, initialCoinsDenomOne),
			addr:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			options:          emptyPositionOptions,
			expectedPosition: withInitialAccumValue(positionOne, initialCoinsDenomOne),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			// System under test.
			tc.accObject.NewPosition(tc.addr, tc.numShareUnits, tc.options)

			// Assertions.
			positions := tc.accObject.GetPosition(tc.addr)
			suite.Require().Equal(tc.expectedPosition, positions)
		})
	}
}

func (suite *AccumTestSuite) TestClaimRewards() {
	var (
		doubleCoinsDenomOne = sdk.NewDecCoinFromDec(denomOne, initialValueOne.MulInt64(2))

		tripleDenomOneAndTwo = sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denomOne, initialValueOne),
			sdk.NewDecCoinFromDec(denomTwo, sdk.OneDec())).MulDec(sdk.NewDec(3))
	)

	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.

	// 1. No rewards, 2 position accumulator.
	accumNoRewards := accumPackage.CreateRawAccumObject(suite.store, testNameOne, emptyCoins)

	// Create positions at testAddressOne and testAddressTwo.
	accumNoRewards.NewPosition(testAddressOne, positionOne.NumShares, emptyPositionOptions)
	accumNoRewards.NewPosition(testAddressTwo, positionTwo.NumShares, emptyPositionOptions)

	// 2. One accumulator reward coin, 1 position accumulator, no unclaimed rewards in position.
	accumOneReward := accumPackage.CreateRawAccumObject(suite.store, testNameTwo, initialCoinsDenomOne)

	// Create position at testAddressThree.
	accumOneReward = accumPackage.WithPosition(accumOneReward, testAddressThree, withInitialAccumValue(positionThree, initialCoinsDenomOne))

	// Double the accumulator value.
	accumOneReward.SetValue(sdk.NewDecCoins(doubleCoinsDenomOne))

	// 3. Multi accumulator rewards, 2 position accumulator, some unclaimed rewards.
	accumThreeRewards := accumPackage.CreateRawAccumObject(suite.store, testNameThree, sdk.NewDecCoins())

	// Create positions at testAddressOne
	// This position has unclaimed rewards set.
	accumThreeRewards = accumPackage.WithPosition(accumThreeRewards, testAddressOne, withUnclaimedRewards(positionOne, initialCoinsDenomOne))

	// Create positions at testAddressThree with no unclaimed rewards.
	accumThreeRewards.NewPosition(testAddressTwo, positionTwo.NumShares, emptyPositionOptions)

	// Triple the accumulator value.
	accumThreeRewards.SetValue(tripleDenomOneAndTwo)

	tests := map[string]struct {
		accObject      accumPackage.AccumulatorObject
		addr           sdk.AccAddress
		expectedResult sdk.DecCoins
		expectError    error
	}{
		"claim at testAddressOne with no rewards - success": {
			accObject:      accumNoRewards,
			addr:           testAddressOne,
			expectedResult: emptyCoins,
		},
		"claim at testAddressTwo with no rewards - success": {
			accObject:      accumNoRewards,
			addr:           testAddressTwo,
			expectedResult: emptyCoins,
		},
		"claim at testAddressTwo with no rewards - error - no position": {
			accObject:   accumNoRewards,
			addr:        testAddressThree,
			expectError: accumPackage.NoPositionError{Address: testAddressThree},
		},
		"claim at testAddressThree with single reward token - success": {
			accObject: accumOneReward,
			addr:      testAddressThree,
			// denomOne: (200.2 - 100.1) * 300 (accum diff * share count) = 30030
			expectedResult: initialCoinsDenomOne.MulDec(positionThree.NumShares),
		},
		"claim at testAddressOne with multiple reward tokens and unclaimed rewards - success": {
			accObject: accumThreeRewards,
			addr:      testAddressOne,
			// denomOne: (300.3 - 0) * 100 (accum diff * share count) + 100.1 (unclaimed rewards) = 30130.1
			// denomTwo: (3 - 0) * 100 (accum diff * share count) = 300
			expectedResult: tripleDenomOneAndTwo.MulDec(positionOne.NumShares).Add(initialCoinDenomOne),
		},
		"claim at testAddressTwo with multiple reward tokens and no unclaimed rewards - success": {
			accObject: accumThreeRewards,
			addr:      testAddressTwo,
			// denomOne: (300.3 - 0) * 200 (accum diff * share count) = 60060.6
			// denomTwo: (3 - 0) * 200  (accum diff * share count) = 600
			expectedResult: sdk.NewDecCoins(
				initialCoinDenomOne,
				sdk.NewDecCoinFromDec(denomTwo, sdk.OneDec()),
			).MulDec(positionTwo.NumShares).MulDec(sdk.NewDec(3)),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {

			// System under test.
			actualResult, err := tc.accObject.ClaimRewards(tc.addr)

			// Assertions.

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectError, err)
				return
			}

			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedResult, actualResult)

			finalPosition := tc.accObject.GetPosition(tc.addr)
			suite.Require().NoError(err)

			// Unclaimed rewards are reset.
			suite.Require().True(finalPosition.UnclaimedRewards.IsZero())
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
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.NewDec(10),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: true,
		},
		"non-zero shares with new rewards in one denom": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unclaimed rewards since last update
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne),
			expPass: true,
		},
		"non-zero shares with new rewards in two denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"non-zero shares with both existing and new rewards": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin(denomOne, sdk.NewInt(11)), sdk.NewDecCoin(denomTwo, sdk.NewInt(11))),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two denoms)": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two new denoms)": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomTwo, sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with same denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with new denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			accumDelta: sdk.NewDecCoins(initialCoinDenomTwo, sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},
		"decimal shares with new rewards in two denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.NewDecWithPrec(983429874321, 5),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
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
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.ZeroDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
			expPass: false,
		},
		"attempt to add negative shares": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares: sdk.OneDec().Neg(),
			accumInit: sdk.NewDecCoins(),
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
			curAccum := accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDNE {
				accumPackage.CreateRawPosition(curAccum, addr, tc.startingNumShares, tc.startingUnclaimedRewards, emptyPositionOptions)
			}

			// Update accumulator with accumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit.Add(tc.accumDelta...))

			// Add newShares to position
			err := curAccum.AddToPosition(addr, tc.newShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(curAccum, addr)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.accumDelta...), newPosition.InitAccumValue)

				// Ensure accrued rewards are moved into UnclaimedRewards (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.accumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewards)

				// Ensure address's position properly reflects new number of shares
				suite.Require().Equal(tc.startingNumShares.Add(tc.newShares), newPosition.NumShares)

				// Ensure a new position isn't created in memory (only old one is overwritten)
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AccumTestSuite) TestRemoveFromPosition() {
	type testcase struct {
		startingNumShares sdk.Dec
		startingUnclaimedRewards sdk.DecCoins
		removedShares   sdk.Dec

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
		"no new rewards": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			accumDelta: sdk.NewDecCoins(),
			expPass: true,
		},
		"new rewards in one denom": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			// unclaimed rewards since last update
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne),
			expPass: true,
		},
		"new rewards in two denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"both existing and new rewards": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin(denomOne, sdk.NewInt(11)), sdk.NewDecCoin(denomTwo, sdk.NewInt(11))),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"both existing (one denom) and new rewards (two denoms, one overlapping)": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"both existing (one denom) and new rewards (two new denoms)": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomTwo, sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with same denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},
		"nonzero accumulator starting value, delta with new denoms": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			accumDelta: sdk.NewDecCoins(initialCoinDenomTwo, sdk.NewDecCoin("baz", sdk.NewInt(10))),
			expPass: true,
		},
		"remove decimal shares with new rewards in two denoms": {
			startingNumShares: sdk.NewDec(1000000),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.NewDecWithPrec(7489274134, 5),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass: true,
		},

		// error catching
		"account does not exist": {
			addrDNE: true,
			expPass: false,

			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
		},
		"attempt to remove zero shares": {
			startingNumShares: initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.ZeroDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
			expPass: false,
		},
		"attempt to remove negative shares": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec().Neg(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
			expPass: false,
		},
		"attempt to remove exactly numShares": {
			startingNumShares: sdk.OneDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
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
			curAccum := accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDNE {
				accumPackage.CreateRawPosition(curAccum, addr, tc.startingNumShares, tc.startingUnclaimedRewards, emptyPositionOptions)
			}

			// Update accumulator with accumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit.Add(tc.accumDelta...))

			// Remove removedShares from position
			err := curAccum.RemoveFromPosition(addr, tc.removedShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(curAccum, addr)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.accumDelta...), newPosition.InitAccumValue)

				// Ensure accrued rewards are moved into UnclaimedRewards (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.accumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewards)

				// Ensure address's position properly reflects new number of shares
				suite.Require().Equal(tc.startingNumShares.Sub(tc.removedShares), newPosition.NumShares)

				// Ensure a new position isn't created in memory (only old one is overwritten)
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AccumTestSuite) TestGetPositionSize() {
	type testcase struct {
		numShares sdk.Dec
		changedShares sdk.Dec

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
		"unchanged accumulator": {
			numShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
			changedShares: sdk.ZeroDec(),
			expPass: true,
		},
		"changed accumulator": {
			numShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			changedShares: sdk.ZeroDec(),
			expPass: true,
		},
		"changed number of shares": {
			numShares: sdk.OneDec(),
			accumInit: sdk.NewDecCoins(),
			accumDelta: sdk.NewDecCoins(),
			changedShares: sdk.OneDec(),
			expPass: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// We reset the store for each test
			suite.SetupTest()
			addr := osmoutils.CreateRandomAccounts(1)[0]

			// Create a new accumulator with initial value specified by test case
			curAccum := accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDNE {
				accumPackage.CreateRawPosition(curAccum, addr, tc.numShares, sdk.NewDecCoins(), emptyPositionOptions)
			}

			// Update accumulator with accumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.CreateRawAccumObject(suite.store, testNameOne, tc.accumInit.Add(tc.accumDelta...))

			// Get position size from valid address (or from nonexistant if addrDNE)
			positionSize, err := curAccum.GetPositionSize(addr)

			if tc.changedShares.GT(sdk.ZeroDec()) {
				accumPackage.CreateRawPosition(curAccum, addr, tc.numShares.Add(tc.changedShares), sdk.NewDecCoins(), emptyPositionOptions)
			}

			positionSize, err = curAccum.GetPositionSize(addr)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.numShares.Add(tc.changedShares), positionSize)

				// Ensure nothing was added or removed from store
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}