package accum_test

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/iavl"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils"
	accumPackage "github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
)

type AccumTestSuite struct {
	suite.Suite

	store store.KVStore
}

var (
	testAddressOne   = sdk.AccAddress([]byte("addr1_______________")).String()
	testAddressTwo   = sdk.AccAddress([]byte("addr2_______________")).String()
	testAddressThree = sdk.AccAddress([]byte("addr3_______________")).String()

	emptyPositionOptions = accumPackage.Options{}
	testNameOne          = "myaccumone"
	testNameTwo          = "myaccumtwo"
	testNameThree        = "myaccumthree"
	denomOne             = "denomone"
	denomTwo             = "denomtwo"
	denomThree           = "denomthree"

	emptyCoins = sdk.DecCoins(nil)
	emptyDec   = sdk.NewDec(0)

	initialValueOne       = sdk.MustNewDecFromStr("100.1")
	initialCoinDenomOne   = sdk.NewDecCoinFromDec(denomOne, initialValueOne)
	initialCoinDenomTwo   = sdk.NewDecCoinFromDec(denomTwo, initialValueOne)
	initialCoinDenomThree = sdk.NewDecCoinFromDec(denomThree, initialValueOne)
	initialCoinsDenomOne  = sdk.NewDecCoins(initialCoinDenomOne)

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

func TestAccumTestSuite(t *testing.T) {
	suite.Run(t, new(AccumTestSuite))
}

func (suite *AccumTestSuite) TestMakeAndGetAccum() {
	// We set up store once at beginning so we can test duplicates
	suite.SetupTest()

	type testcase struct {
		testName   string
		accumName  string
		expAccum   accumPackage.AccumulatorObject
		expSetPass bool
		expGetPass bool
	}

	tests := []testcase{
		{
			testName:   "create valid accumulator",
			accumName:  "fee-accumulator",
			expSetPass: true,
			expGetPass: true,
		},
		{
			testName:   "create duplicate accumulator",
			accumName:  "fee-accumulator",
			expSetPass: false,
			expGetPass: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.testName, func() {
			// Creates raw accumulator object with test case's accum name and zero initial value
			expAccum := accumPackage.MakeTestAccumulator(suite.store, tc.accumName, emptyCoins, emptyDec)

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
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, emptyCoins, emptyDec)

	tests := map[string]struct {
		accObject        accumPackage.AccumulatorObject
		name             string
		numShareUnits    sdk.Dec
		options          *accumPackage.Options
		expectedPosition accumPackage.Record
	}{
		"test address one - position created": {
			accObject:        accObject,
			name:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			expectedPosition: positionOne,
		},
		"test address two (non-nil options) - position created": {
			accObject:        accObject,
			name:             testAddressTwo,
			numShareUnits:    positionTwo.NumShares,
			expectedPosition: positionTwo,
			options:          &emptyPositionOptions,
		},
		"test address one - position overwritten": {
			accObject:        accObject,
			name:             testAddressOne,
			numShareUnits:    positionOneV2.NumShares,
			expectedPosition: positionOneV2,
		},
		"test address three - added": {
			accObject:        accObject,
			name:             testAddressThree,
			numShareUnits:    positionThree.NumShares,
			expectedPosition: positionThree,
		},
		"test address one with non-empty accumulator - position created": {
			accObject:        accumPackage.MakeTestAccumulator(suite.store, testNameTwo, initialCoinsDenomOne, emptyDec),
			name:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			expectedPosition: withInitialAccumValue(positionOne, initialCoinsDenomOne),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// System under test.
			tc.accObject.NewPosition(tc.name, tc.numShareUnits, tc.options)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, position.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, position.UnclaimedRewards)

			if tc.options == nil {
				suite.Require().Nil(position.Options)
				return
			}

			suite.Require().Equal(*tc.options, *position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestNewPositionCustomAcc() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		accObject        accumPackage.AccumulatorObject
		name             string
		numShareUnits    sdk.Dec
		customAcc        sdk.DecCoins
		options          *accumPackage.Options
		expectedPosition accumPackage.Record
		expectedError    error
	}{
		"custom acc value equals to acc": {
			accObject:     accObject,
			name:          testAddressOne,
			numShareUnits: positionOne.NumShares,
			customAcc:     accObject.GetValue(),
			expectedPosition: accumPackage.Record{
				NumShares:        positionOne.NumShares,
				InitAccumValue:   accObject.GetValue(),
				UnclaimedRewards: emptyCoins,
			},
		},
		"custom acc value does not equal to acc": {
			accObject:     accObject,
			name:          testAddressTwo,
			numShareUnits: positionTwo.NumShares,
			customAcc:     accObject.GetValue().MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:        positionTwo.NumShares,
				InitAccumValue:   accObject.GetValue().MulDec(sdk.NewDec(2)),
				UnclaimedRewards: emptyCoins,
			},
			options: &emptyPositionOptions,
		},
		"negative acc value - error": {
			accObject:     accObject,
			name:          testAddressOne,
			numShareUnits: positionOne.NumShares,
			customAcc:     accObject.GetValue().MulDec(sdk.NewDec(-1)),
			expectedError: accumPackage.NegativeCustomAccError{accObject.GetValue().MulDec(sdk.NewDec(-1))},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// System under test.
			err := tc.accObject.NewPositionCustomAcc(tc.name, tc.numShareUnits, tc.customAcc, tc.options)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, position.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, position.UnclaimedRewards)

			if tc.options == nil {
				suite.Require().Nil(position.Options)
				return
			}

			suite.Require().Equal(*tc.options, *position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestClaimRewards() {
	var (
		doubleCoinsDenomOne = sdk.NewDecCoinFromDec(denomOne, initialValueOne.MulInt64(2))

		tripleDenomOneAndTwo = sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denomOne, initialValueOne),
			sdk.NewDecCoinFromDec(denomTwo, sdk.NewDec(3)))
	)

	// single output convenience wrapper.
	toCoins := func(decCoins sdk.DecCoins) sdk.Coins {
		coins, _ := decCoins.TruncateDecimal()
		return coins
	}

	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.

	// 1. No rewards, 2 position accumulator.
	accumNoRewards := accumPackage.MakeTestAccumulator(suite.store, testNameOne, emptyCoins, emptyDec)

	// Create positions at testAddressOne and testAddressTwo.
	accumNoRewards.NewPosition(testAddressOne, positionOne.NumShares, nil)
	accumNoRewards.NewPosition(testAddressTwo, positionTwo.NumShares, nil)

	// 2. One accumulator reward coin, 1 position accumulator, no unclaimed rewards in position.
	accumOneReward := accumPackage.MakeTestAccumulator(suite.store, testNameTwo, initialCoinsDenomOne, emptyDec)

	// Create position at testAddressThree.
	accumOneReward = accumPackage.WithPosition(accumOneReward, testAddressThree, withInitialAccumValue(positionThree, initialCoinsDenomOne))

	// Double the accumulator value.
	accumOneReward.SetValue(sdk.NewDecCoins(doubleCoinsDenomOne))

	// 3. Multi accumulator rewards, 2 position accumulator, some unclaimed rewards.
	accumThreeRewards := accumPackage.MakeTestAccumulator(suite.store, testNameThree, sdk.NewDecCoins(), emptyDec)

	// Create positions at testAddressOne
	// This position has unclaimed rewards set.
	accumThreeRewards = accumPackage.WithPosition(accumThreeRewards, testAddressOne, withUnclaimedRewards(positionOne, initialCoinsDenomOne))

	// Create positions at testAddressThree with no unclaimed rewards.
	accumThreeRewards.NewPosition(testAddressTwo, positionTwo.NumShares, nil)

	// Triple the accumulator value.
	accumThreeRewards.SetValue(tripleDenomOneAndTwo)

	tests := []struct {
		testName              string
		accObject             accumPackage.AccumulatorObject
		accName               string
		expectedResult        sdk.Coins
		updateNumSharesToZero bool
		expectError           error
	}{
		{
			testName:       "claim at testAddressOne with no rewards - success",
			accObject:      accumNoRewards,
			accName:        testAddressOne,
			expectedResult: toCoins(emptyCoins),
		},
		{
			testName:              "delete accum - claim at testAddressOne with no rewards - success",
			accObject:             accumNoRewards,
			accName:               testAddressOne,
			updateNumSharesToZero: true,
			expectedResult:        toCoins(emptyCoins),
		},
		{
			testName:       "claim at testAddressTwo with no rewards - success",
			accObject:      accumNoRewards,
			accName:        testAddressTwo,
			expectedResult: toCoins(emptyCoins),
		},
		{
			testName:    "claim at testAddressTwo with no rewards - error - no position",
			accObject:   accumNoRewards,
			accName:     testAddressThree,
			expectError: accumPackage.NoPositionError{Name: testAddressThree},
		},
		{
			testName:  "claim at testAddressThree with single reward token - success",
			accObject: accumOneReward,
			accName:   testAddressThree,
			// denomOne: (200.2 - 100.1) * 300 (accum diff * share count) = 30030
			expectedResult: toCoins(initialCoinsDenomOne.MulDec(positionThree.NumShares)),
		},
		{
			testName:  "claim at testAddressOne with multiple reward tokens and unclaimed rewards - success",
			accObject: accumThreeRewards,
			accName:   testAddressOne,
			// denomOne: (300.3 - 0) * 100 (accum diff * share count) + 100.1 (unclaimed rewards) = 30130.1
			// denomTwo: (3 - 0) * 100 (accum diff * share count) = 300
			expectedResult: toCoins(tripleDenomOneAndTwo.MulDec(positionOne.NumShares).Add(initialCoinDenomOne)),
		},
		{
			testName:              "delete accum - claim at testAddressOne with multiple reward tokens and unclaimed rewards - success",
			accObject:             accumThreeRewards,
			accName:               testAddressOne,
			updateNumSharesToZero: true,
			// all claimed during the previous test
			expectedResult: toCoins(emptyCoins),
		},
		{
			testName:  "claim at testAddressTwo with multiple reward tokens and no unclaimed rewards - success",
			accObject: accumThreeRewards,
			accName:   testAddressTwo,
			// denomOne: (100.1 - 0) * 200 (accum diff * share count) = 200020
			// denomTwo: (3 - 0) * 200  (accum diff * share count) = 600
			expectedResult: toCoins(tripleDenomOneAndTwo.MulDec(positionTwo.NumShares)),
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.testName, func() {
			if tc.updateNumSharesToZero {
				positionSize, err := tc.accObject.GetPositionSize(tc.accName)
				suite.Require().NoError(err)
				err = tc.accObject.UpdatePosition(tc.accName, positionSize.Neg())
				suite.Require().NoError(err)
			}
			// System under test.
			actualResult, err := tc.accObject.ClaimRewards(tc.accName)

			// Assertions.

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectError, err)
				return
			}

			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedResult.String(), actualResult.String())

			osmoassert.ConditionalPanic(suite.T(), tc.updateNumSharesToZero, func() {
				finalPosition := tc.accObject.MustGetPosition(tc.accName)
				suite.Require().NoError(err)

				// Unclaimed rewards are reset.
				suite.Require().Equal(emptyCoins, finalPosition.UnclaimedRewards)
			})
		})
	}
}

func (suite *AccumTestSuite) TestAddToPosition() {
	type testcase struct {
		startingNumShares        sdk.Dec
		startingUnclaimedRewards sdk.DecCoins
		newShares                sdk.Dec

		// accumInit and expAccumDelta specify the initial accum value
		// and how much it has changed since the position being added
		// to was created
		accumInit     sdk.DecCoins
		expAccumDelta sdk.DecCoins

		addrDoesNotExist bool
		expPass          bool
	}

	tests := map[string]testcase{
		"zero shares with no new rewards": {
			startingNumShares:        sdk.ZeroDec(),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			expAccumDelta: sdk.NewDecCoins(),
			expPass:       true,
		},
		"non-zero shares with no new rewards": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares:                sdk.NewDec(10),
			accumInit:                sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			expAccumDelta: sdk.NewDecCoins(),
			expPass:       true,
		},
		"non-zero shares with new rewards in one denom": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			// unclaimed rewards since last update
			expAccumDelta: sdk.NewDecCoins(initialCoinDenomOne),
			expPass:       true,
		},
		"non-zero shares with new rewards in two denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"non-zero shares with both existing and new rewards": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin(denomOne, sdk.NewInt(11)), sdk.NewDecCoin(denomTwo, sdk.NewInt(11))),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two denoms)": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"non-zero shares with both existing (one denom) and new rewards (two new denoms)": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomTwo, initialCoinDenomThree),
			expPass:                  true,
		},
		"nonzero accumulator starting value, delta with same denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"nonzero accumulator starting value, delta with new denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			newShares:                sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomTwo, initialCoinDenomThree),
			expPass:                  true,
		},
		"decimal shares with new rewards in two denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			newShares:                sdk.NewDecWithPrec(983429874321, 5),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},

		// error catching
		"account does not exist": {
			addrDoesNotExist: true,
			expPass:          false,

			startingNumShares:        sdk.OneDec(),
			startingUnclaimedRewards: emptyCoins,
			newShares:                sdk.OneDec(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
		},
		"attempt to add zero shares": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: emptyCoins,
			newShares:                sdk.ZeroDec(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
			expPass:                  false,
		},
		"attempt to add negative shares": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: emptyCoins,
			newShares:                sdk.OneDec().Neg(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
			expPass:                  false,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// We reset the store for each test
			suite.SetupTest()
			positionName := osmoutils.CreateRandomAccounts(1)[0].String()

			// Create a new accumulator with initial value specified by test case
			curAccum := accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit, emptyDec)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDoesNotExist {
				accumPackage.CreateRawPosition(curAccum, positionName, tc.startingNumShares, tc.startingUnclaimedRewards, nil)
			}

			// Update accumulator with expAccumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit.Add(tc.expAccumDelta...), emptyDec)

			// Add newShares to position
			err := curAccum.AddToPosition(positionName, tc.newShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(curAccum, positionName)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.expAccumDelta...), newPosition.InitAccumValue)

				// Ensure accrued rewards are moved into UnclaimedRewards (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.expAccumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewards)

				// Ensure address's position properly reflects new number of shares
				suite.Require().Equal(tc.startingNumShares.Add(tc.newShares), newPosition.NumShares)

				// Ensure a new position isn't created or removed from memory
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)
			} else {
				suite.Require().Error(err)

				// Further checks to ensure state was not mutated upon error
				if !tc.addrDoesNotExist {
					// Get new position for comparison
					newPosition, err := accumPackage.GetPosition(curAccum, positionName)
					suite.Require().NoError(err)

					// Ensure that numShares, accumulator value, and unclaimed rewards are unchanged
					suite.Require().Equal(tc.startingNumShares, newPosition.NumShares)
					suite.Require().Equal(tc.accumInit, newPosition.InitAccumValue)
					suite.Require().Equal(tc.startingUnclaimedRewards, newPosition.UnclaimedRewards)
				}
			}
		})
	}
}

// TestAddToPositionCustomAcc this test only focuses on testing the
// custom accumulator value functionality of adding to position.
func (suite *AccumTestSuite) TestAddToPositionCustomAcc() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		accObject        accumPackage.AccumulatorObject
		name             string
		numShareUnits    sdk.Dec
		customAcc        sdk.DecCoins
		expectedPosition accumPackage.Record
		expectedError    error
	}{
		"custom acc value equals to acc": {
			accObject:     accObject,
			name:          testAddressOne,
			numShareUnits: positionOne.NumShares,
			customAcc:     accObject.GetValue(),
			expectedPosition: accumPackage.Record{
				NumShares:        positionOne.NumShares,
				InitAccumValue:   accObject.GetValue(),
				UnclaimedRewards: emptyCoins,
			},
		},
		"custom acc value does not equal to acc": {
			accObject:     accObject,
			name:          testAddressTwo,
			numShareUnits: positionTwo.NumShares,
			customAcc:     accObject.GetValue().MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:        positionTwo.NumShares,
				InitAccumValue:   accObject.GetValue().MulDec(sdk.NewDec(2)),
				UnclaimedRewards: emptyCoins,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// Setup
			err := tc.accObject.NewPositionCustomAcc(tc.name, sdk.ZeroDec(), tc.accObject.GetValue(), nil)
			suite.Require().NoError(err)

			// System under test.
			err = tc.accObject.AddToPositionCustomAcc(tc.name, tc.numShareUnits, tc.customAcc)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, position.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, position.UnclaimedRewards)
			suite.Require().Nil(position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestRemoveFromPosition() {
	type testcase struct {
		startingNumShares        sdk.Dec
		startingUnclaimedRewards sdk.DecCoins
		removedShares            sdk.Dec

		// accumInit and expAccumDelta specify the initial accum value
		// and how much it has changed since the position being added
		// to was created
		accumInit     sdk.DecCoins
		expAccumDelta sdk.DecCoins

		addrDoesNotExist bool
		expPass          bool
	}

	tests := map[string]testcase{
		"no new rewards": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			// unchanged accum value, so no unclaimed rewards
			expAccumDelta: sdk.NewDecCoins(),
			expPass:       true,
		},
		"new rewards in one denom": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			// unclaimed rewards since last update
			expAccumDelta: sdk.NewDecCoins(initialCoinDenomOne),
			expPass:       true,
		},
		"new rewards in two denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"both existing and new rewards": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin(denomOne, sdk.NewInt(11)), sdk.NewDecCoin(denomTwo, sdk.NewInt(11))),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"both existing (one denom) and new rewards (two denoms, one overlapping)": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"both existing (one denom) and new rewards (two new denoms)": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomTwo, initialCoinDenomThree),
			expPass:                  true,
		},
		"nonzero accumulator starting value, delta with same denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"nonzero accumulator starting value, delta with new denoms": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: sdk.NewDecCoins(initialCoinDenomOne),
			removedShares:            sdk.OneDec(),
			accumInit:                sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomTwo, initialCoinDenomThree),
			expPass:                  true,
		},
		"remove decimal shares with new rewards in two denoms": {
			startingNumShares:        sdk.NewDec(1000000),
			startingUnclaimedRewards: sdk.NewDecCoins(),
			removedShares:            sdk.NewDecWithPrec(7489274134, 5),
			accumInit:                sdk.NewDecCoins(),
			expAccumDelta:            sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			expPass:                  true,
		},
		"attempt to remove exactly numShares": {
			startingNumShares:        sdk.OneDec(),
			startingUnclaimedRewards: emptyCoins,
			removedShares:            sdk.OneDec(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
			expPass:                  true,
		},

		// error catching
		"account does not exist": {
			addrDoesNotExist: true,
			expPass:          false,

			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: emptyCoins,
			removedShares:            sdk.OneDec(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
		},
		"attempt to remove zero shares": {
			startingNumShares:        initialValueOne,
			startingUnclaimedRewards: emptyCoins,
			removedShares:            sdk.ZeroDec(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
			expPass:                  false,
		},
		"attempt to remove negative shares": {
			startingNumShares:        sdk.OneDec(),
			startingUnclaimedRewards: emptyCoins,
			removedShares:            sdk.OneDec().Neg(),
			accumInit:                emptyCoins,
			expAccumDelta:            sdk.NewDecCoins(),
			expPass:                  false,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// We reset the store for each test
			suite.SetupTest()
			positionName := osmoutils.CreateRandomAccounts(1)[0].String()

			// Create a new accumulator with initial value specified by test case
			curAccum := accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit, emptyDec)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDoesNotExist {
				accumPackage.CreateRawPosition(curAccum, positionName, tc.startingNumShares, tc.startingUnclaimedRewards, nil)
			}

			// Update accumulator with expAccumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit.Add(tc.expAccumDelta...), emptyDec)

			// Remove removedShares from position
			err := curAccum.RemoveFromPosition(positionName, tc.removedShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(curAccum, positionName)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.expAccumDelta...), newPosition.InitAccumValue)

				// Ensure accrued rewards are moved into UnclaimedRewards (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.expAccumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewards)

				// Ensure address's position properly reflects new number of shares
				if (tc.startingNumShares.Sub(tc.removedShares)).Equal(sdk.ZeroDec()) {
					suite.Require().Equal(emptyDec, newPosition.NumShares)
				} else {
					suite.Require().Equal(tc.startingNumShares.Sub(tc.removedShares), newPosition.NumShares)
				}

				// Ensure a new position isn't created in memory (only old one is overwritten)
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)
			} else {
				suite.Require().Error(err)

				// Further checks to ensure state was not mutated upon error
				if !tc.addrDoesNotExist {
					// Get new position for comparison
					newPosition, err := accumPackage.GetPosition(curAccum, positionName)
					suite.Require().NoError(err)

					// Ensure that numShares, accumulator value, and unclaimed rewards are unchanged
					suite.Require().Equal(tc.startingNumShares, newPosition.NumShares)
					suite.Require().Equal(tc.accumInit, newPosition.InitAccumValue)
					suite.Require().Equal(tc.startingUnclaimedRewards, newPosition.UnclaimedRewards)
				}
			}
		})
	}
}

// TestRemoveFromPositionCustomAcc this test only focuses on testing the
// custom accumulator value functionality of removing from a position.
func (suite *AccumTestSuite) TestRemoveFromPositionCustomAcc() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	baseAccumValue := initialCoinsDenomOne

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, baseAccumValue, emptyDec)

	tests := map[string]struct {
		accObject        accumPackage.AccumulatorObject
		name             string
		numShareUnits    sdk.Dec
		customAcc        sdk.DecCoins
		expectedPosition accumPackage.Record
		expectedError    error
	}{
		"custom acc value equals to acc": {
			accObject:     accObject,
			name:          testAddressOne,
			numShareUnits: positionOne.NumShares,
			customAcc:     baseAccumValue,
			expectedPosition: accumPackage.Record{
				NumShares:      sdk.ZeroDec(),
				InitAccumValue: baseAccumValue,
				// base value - 0.5 * base = base value
				UnclaimedRewards: baseAccumValue.MulDec(sdk.NewDecWithPrec(5, 1)).MulDec(positionOne.NumShares),
			},
		},
		"custom acc value does not equal to acc": {
			accObject:     accObject,
			name:          testAddressTwo,
			numShareUnits: positionTwo.NumShares,
			customAcc:     baseAccumValue.MulDec(sdk.NewDecWithPrec(75, 2)),
			expectedPosition: accumPackage.Record{
				NumShares:      sdk.ZeroDec(),
				InitAccumValue: baseAccumValue.MulDec(sdk.NewDecWithPrec(75, 2)),
				// base value - 0.75 * base = 0.25 * base
				UnclaimedRewards: baseAccumValue.MulDec(sdk.NewDecWithPrec(25, 2)).MulDec(positionTwo.NumShares),
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// Setup

			// Original position's accum is always set to 0.5 * base value.
			err := tc.accObject.NewPositionCustomAcc(tc.name, tc.numShareUnits, initialCoinsDenomOne.MulDec(sdk.NewDecWithPrec(5, 1)), nil)
			suite.Require().NoError(err)

			tc.accObject.SetValue(tc.customAcc)

			// System under test.
			err = tc.accObject.RemoveFromPositionCustomAcc(tc.name, tc.numShareUnits, tc.customAcc)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, position.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, position.UnclaimedRewards)
			suite.Require().Nil(position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestGetPositionSize() {
	type testcase struct {
		numShares     sdk.Dec
		changedShares sdk.Dec

		// accumInit and expAccumDelta specify the initial accum value
		// and how much it has changed since the position being added
		// to was created
		accumInit     sdk.DecCoins
		expAccumDelta sdk.DecCoins

		addrDoesNotExist bool
		expPass          bool
	}

	tests := map[string]testcase{
		"unchanged accumulator": {
			numShares:     sdk.OneDec(),
			accumInit:     sdk.NewDecCoins(),
			expAccumDelta: sdk.NewDecCoins(),
			changedShares: sdk.ZeroDec(),
			expPass:       true,
		},
		"changed accumulator": {
			numShares:     sdk.OneDec(),
			accumInit:     sdk.NewDecCoins(),
			expAccumDelta: sdk.NewDecCoins(initialCoinDenomOne, initialCoinDenomTwo),
			changedShares: sdk.ZeroDec(),
			expPass:       true,
		},
		"changed number of shares": {
			numShares:     sdk.OneDec(),
			accumInit:     sdk.NewDecCoins(),
			expAccumDelta: sdk.NewDecCoins(),
			changedShares: sdk.OneDec(),
			expPass:       true,
		},
		"account does not exist": {
			addrDoesNotExist: true,
			expPass:          false,

			numShares:     sdk.OneDec(),
			accumInit:     sdk.NewDecCoins(),
			expAccumDelta: sdk.NewDecCoins(),
			changedShares: sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// We reset the store for each test
			suite.SetupTest()
			positionName := osmoutils.CreateRandomAccounts(1)[0].String()

			// Create a new accumulator with initial value specified by test case
			curAccum := accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit, emptyDec)

			// Create new position in store (raw to minimize dependencies)
			if !tc.addrDoesNotExist {
				accumPackage.CreateRawPosition(curAccum, positionName, tc.numShares, sdk.NewDecCoins(), nil)
			}

			// Update accumulator with expAccumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit.Add(tc.expAccumDelta...), emptyDec)

			// Get position size from valid address (or from nonexistant if address does not exist)
			positionSize, err := curAccum.GetPositionSize(positionName)

			if tc.changedShares.GT(sdk.ZeroDec()) {
				accumPackage.CreateRawPosition(curAccum, positionName, tc.numShares.Add(tc.changedShares), sdk.NewDecCoins(), nil)
			}

			positionSize, err = curAccum.GetPositionSize(positionName)

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

// TestMarhsalUnmarshalRecord displays that we may use Records without options
// For records with nil options, adding new fields to `Options`, should not
// require future migrations.
func (suite *AccumTestSuite) TestMarhsalUnmarshalRecord() {
	suite.SetupTest()

	recordNoOptions := accumPackage.Record{
		NumShares: sdk.OneDec(),
		InitAccumValue: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denomOne, sdk.OneDec()),
		),
		UnclaimedRewards: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denomOne, sdk.OneDec()),
		),
	}

	bz, err := proto.Marshal(&recordNoOptions)
	suite.Require().NoError(err)

	var unmarshaledRecord accumPackage.Record
	proto.Unmarshal(bz, &unmarshaledRecord)
	// Options should be nil, not an empty struct
	suite.Require().True(unmarshaledRecord.Options == nil)
}

func (suite *AccumTestSuite) TestAddToAccumulator() {
	tests := map[string]struct {
		updateAmount sdk.DecCoins

		expectedValue sdk.DecCoins
	}{
		"positive": {
			updateAmount: initialCoinsDenomOne,

			expectedValue: initialCoinsDenomOne,
		},
		"negative": {
			updateAmount: initialCoinsDenomOne.MulDec(sdk.NewDec(-1)),

			expectedValue: initialCoinsDenomOne.MulDec(sdk.NewDec(-1)),
		},
		"multiple coins": {
			updateAmount: initialCoinsDenomOne.Add(initialCoinDenomTwo),

			expectedValue: initialCoinsDenomOne.Add(initialCoinDenomTwo),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			// Setup
			suite.SetupTest()

			err := accumPackage.MakeAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)
			originalAccum, err := accumPackage.GetAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			// System under test.
			originalAccum.AddToAccumulator(tc.updateAmount)

			// Validations.

			// validate that the reciever is mutated.
			suite.Require().Equal(tc.expectedValue, originalAccum.GetValue())

			accumFromStore, err := accumPackage.GetAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			// validate that store is updated.
			suite.Require().Equal(tc.expectedValue, accumFromStore.GetValue())
		})
	}
}

func (suite *AccumTestSuite) TestUpdatePosition() {
	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		name             string
		numShares        sdk.Dec
		expectedPosition accumPackage.Record
		expectError      error
	}{
		"positive - acts as AddToPosition": {
			name:      testAddressOne,
			numShares: sdk.OneDec(),

			expectedPosition: accumPackage.Record{
				NumShares:        sdk.OneDec().MulInt64(2),
				InitAccumValue:   initialCoinsDenomOne,
				UnclaimedRewards: emptyCoins,
			},
		},
		"negative - acts as RemoveFromPosition": {
			name:      testAddressOne,
			numShares: sdk.OneDec().Neg(),

			expectedPosition: accumPackage.Record{
				NumShares:        sdk.ZeroDec(),
				InitAccumValue:   initialCoinsDenomOne,
				UnclaimedRewards: emptyCoins,
			},
		},
		"zero - error": {
			name:      testAddressOne,
			numShares: sdk.ZeroDec(),

			expectError: accumPackage.ZeroSharesError,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			err := accObject.NewPosition(tc.name, sdk.OneDec(), nil)
			suite.Require().NoError(err)

			err = accObject.UpdatePosition(tc.name, tc.numShares)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}
			suite.Require().NoError(err)

			updatedPosition := accObject.MustGetPosition(tc.name)

			// Assertions.
			position := accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, updatedPosition.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, updatedPosition.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, updatedPosition.UnclaimedRewards)
			suite.Require().Nil(position.Options)
		})
	}
}

// TestUpdatePositionCustomAcc this test only focuses on testing the
// custom accumulator value functionality of updating a position.
func (suite *AccumTestSuite) TestUpdatePositionCustomAcc() {
	tests := []struct {
		testName         string
		initialShares    sdk.Dec
		initialAccum     sdk.DecCoins
		accName          string
		numShareUnits    sdk.Dec
		customAcc        sdk.DecCoins
		expectedPosition accumPackage.Record
		expectedError    error
	}{
		{
			testName:      "custom acc value equals to acc; positive shares -> acts as AddToPosition",
			initialShares: sdk.ZeroDec(),
			initialAccum:  initialCoinsDenomOne,
			accName:       testAddressOne,
			numShareUnits: positionOne.NumShares,
			customAcc:     initialCoinsDenomOne,
			expectedPosition: accumPackage.Record{
				NumShares:        positionOne.NumShares,
				InitAccumValue:   initialCoinsDenomOne,
				UnclaimedRewards: emptyCoins,
			},
		},
		{
			testName:      "custom acc value does not equal to acc; remove same amount -> acts as RemoveFromPosition",
			initialShares: positionTwo.NumShares,
			initialAccum:  initialCoinsDenomOne,
			accName:       testAddressTwo,
			numShareUnits: positionTwo.NumShares.Neg(), // note: negative shares
			customAcc:     initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:        sdk.ZeroDec(), // results in 0 shares (200 - 200)
				InitAccumValue:   initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
				UnclaimedRewards: emptyCoins,
			},
		},
		{
			testName:      "custom acc value does not equal to acc; remove diff amount -> acts as RemoveFromPosition",
			initialShares: positionTwo.NumShares,
			initialAccum:  initialCoinsDenomOne,
			accName:       testAddressTwo,
			numShareUnits: positionOne.NumShares.Neg(), // note: negative shares
			customAcc:     initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:        positionOne.NumShares, // results in 100 shares (200 - 100)
				InitAccumValue:   initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
				UnclaimedRewards: emptyCoins,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.testName, func() {
			suite.SetupTest()

			// make accumualtor based off of tc.accObject
			err := accumPackage.MakeAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			accumObject, err := accumPackage.GetAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			// manually update accumulator value
			accumObject.AddToAccumulator(initialCoinsDenomOne)

			// Setup
			err = accumObject.NewPositionCustomAcc(tc.accName, tc.initialShares, tc.initialAccum, nil)
			suite.Require().NoError(err)

			// System under test.
			err = accumObject.UpdatePositionCustomAcc(tc.accName, tc.numShareUnits, tc.customAcc)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			accumObject, err = accumPackage.GetAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			position := accumObject.MustGetPosition(tc.accName)
			// Assertions.

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.InitAccumValue, position.InitAccumValue)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewards, position.UnclaimedRewards)
			suite.Require().Nil(position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestHasPosition() {
	// We setup store and accum
	// once at beginning.
	suite.SetupTest()

	const (
		defaultPositionName = "posname"
	)

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := []struct {
		name              string
		preCreatePosition bool
	}{
		{
			name:              "position does not exist -> false",
			preCreatePosition: false,
		},
		{
			name:              "position exists -> true",
			preCreatePosition: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			// Setup
			if tc.preCreatePosition {
				err := accObject.NewPosition(defaultPositionName, sdk.ZeroDec(), nil)
				suite.Require().NoError(err)
			}

			hasPosition, err := accObject.HasPosition(defaultPositionName)
			suite.NoError(err)

			suite.Equal(tc.preCreatePosition, hasPosition)
		})
	}
}

func (suite *AccumTestSuite) TestSetPositionCustomAcc() {
	// We setup store and accum
	// once at beginning.
	suite.SetupTest()

	// Setup.
	var (
		accObject           = accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)
		validPositionName   = testAddressThree
		invalidPositionName = testAddressTwo
	)

	tests := map[string]struct {
		positionName           string
		customAccumulatorValue sdk.DecCoins
		expectedError          error
	}{
		"valid update greater than initial value": {
			positionName:           validPositionName,
			customAccumulatorValue: initialCoinsDenomOne.Add(initialCoinDenomOne),
		},
		"valid update equal to the initial value": {
			positionName:           validPositionName,
			customAccumulatorValue: initialCoinsDenomOne,
		},
		"invalid position - different name": {
			positionName:  invalidPositionName,
			expectedError: accumPackage.NoPositionError{Name: invalidPositionName},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {

			// Setup
			err := accObject.NewPositionCustomAcc(validPositionName, sdk.OneDec(), initialCoinsDenomOne, nil)
			suite.Require().NoError(err)

			// System under test.
			err = accObject.SetPositionCustomAcc(tc.positionName, tc.customAccumulatorValue)

			// Assertions.
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			position := accObject.MustGetPosition(tc.positionName)
			suite.Require().Equal(tc.customAccumulatorValue, position.GetInitAccumValue())
			// unchanged
			suite.Require().Equal(sdk.OneDec(), position.NumShares)
			suite.Require().Equal(emptyCoins, position.GetUnclaimedRewards())
		})
	}
}

// We run a series of partially random operations on two accumulators to ensure that total shares are properly tracked in state
func (suite *AccumTestSuite) TestGetTotalShares() {
	suite.SetupTest()

	// Set seed to make tests deterministic
	rand.Seed(1)

	// Set up two accumulators to ensure we can test all relevant behavior
	err := accumPackage.MakeAccumulator(suite.store, testNameOne)
	suite.Require().NoError(err)
	err = accumPackage.MakeAccumulator(suite.store, testNameTwo)
	suite.Require().NoError(err)

	// Fetch both accumulators for initial sanity tests
	accumOne, err := accumPackage.GetAccumulator(suite.store, testNameOne)
	suite.Require().NoError(err)
	accumTwo, err := accumPackage.GetAccumulator(suite.store, testNameTwo)
	suite.Require().NoError(err)

	// Ensure that both accums start at zero shares
	accumOneShares, err := accumOne.GetTotalShares()
	suite.Require().NoError(err)
	accumTwoShares, err := accumTwo.GetTotalShares()
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroDec(), accumOneShares)
	suite.Require().Equal(sdk.ZeroDec(), accumTwoShares)

	// Create position on first accum and pull new accum objects from state
	err = accumOne.NewPosition(testAddressOne, sdk.OneDec(), nil)
	suite.Require().NoError(err)
	accumOne, err = accumPackage.GetAccumulator(suite.store, testNameOne)
	suite.Require().NoError(err)
	accumTwo, err = accumPackage.GetAccumulator(suite.store, testNameTwo)
	suite.Require().NoError(err)

	// Check that total shares for accum one has updated properly and accum two shares are unchanged
	accumOneShares, err = accumOne.GetTotalShares()
	suite.Require().NoError(err)
	accumTwoShares, err = accumTwo.GetTotalShares()
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.OneDec(), accumOneShares)
	suite.Require().Equal(sdk.ZeroDec(), accumTwoShares)

	// Run a number of NewPosition, AddToPosition, and RemoveFromPosition operations on each accum
	testAddresses := []string{testAddressOne, testAddressTwo, testAddressThree}
	accums := []accumPackage.AccumulatorObject{accumOne, accumTwo}
	expectedShares := []sdk.Dec{sdk.OneDec(), sdk.ZeroDec()}

	for i := 1; i <= 10; i++ {
		// Cycle through accounts and accumulators
		curAddr := testAddresses[i%3]
		curAccum := accums[i%2]

		// We set a baseAmt that varies with the iteration to increase coverage
		baseAmt := sdk.NewDec(int64(i)).Mul(sdk.NewDec(10))

		// If addr doesn't have a position yet, we make one
		positionExists, err := curAccum.HasPosition(curAddr)
		suite.Require().NoError(err)
		if !positionExists {
			err = curAccum.NewPosition(curAddr, baseAmt, nil)
			suite.Require().NoError(err)
		}

		// We generate a random binary value (0 or 1) to determine
		// whether we will add and/or remove liquidity this loop
		addShares := sdk.NewDec(int64(rand.Int()) % 2)
		removeShares := sdk.NewDec(int64(rand.Int()) % 2)

		// Half the time, we add to the new position
		addAmt := baseAmt.Mul(addShares)
		if addAmt.GT(sdk.ZeroDec()) {
			err = curAccum.AddToPosition(curAddr, addAmt)
			suite.Require().NoError(err)
		}

		// Half the time, we remove one share from the position
		amtToRemove := sdk.OneDec().Mul(removeShares)
		if amtToRemove.GT(sdk.ZeroDec()) {
			err = curAccum.RemoveFromPosition(curAddr, amtToRemove)
			suite.Require().NoError(err)
		}

		// Finally, we update our expected number of shares for the accumulator
		// we targeted in this loop
		if !positionExists {
			// If a new position was created, we factor its new shares in
			expectedShares[i%2] = expectedShares[i%2].Add(baseAmt)
		}
		expectedShares[i%2] = expectedShares[i%2].Add(addAmt).Sub(amtToRemove)
	}

	// Get updated accums from state to validate results
	accumOne, err = accumPackage.GetAccumulator(suite.store, testNameOne)
	suite.Require().NoError(err)
	accumTwo, err = accumPackage.GetAccumulator(suite.store, testNameTwo)
	suite.Require().NoError(err)

	// Ensure that total shares in each accum matches our expected number of shares
	accumOneShares, err = accumOne.GetTotalShares()
	suite.Require().NoError(err)
	accumTwoShares, err = accumTwo.GetTotalShares()
	suite.Require().NoError(err)
	suite.Require().Equal(expectedShares[0], accumOneShares)
	suite.Require().Equal(expectedShares[1], accumTwoShares)
}
