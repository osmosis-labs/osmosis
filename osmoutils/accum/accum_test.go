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

func (suite *AccumTestSuite) GetAccumulator(name string) *accumPackage.AccumulatorObject {
	accum, err := accumPackage.GetAccumulator(suite.store, name)
	suite.Require().NoError(err)
	return accum
}

func (suite *AccumTestSuite) MakeAndGetAccumulator(name string) *accumPackage.AccumulatorObject {
	err := accumPackage.MakeAccumulator(suite.store, name)
	suite.Require().NoError(err)
	accum, err := accumPackage.GetAccumulator(suite.store, name)
	suite.Require().NoError(err)
	return accum
}

func (suite *AccumTestSuite) TotalSharesCheck(accum *accumPackage.AccumulatorObject, expected sdk.Dec) {
	shareCount, err := accum.GetTotalShares()
	suite.Require().NoError(err)
	suite.Require().Equal(expected.String(), shareCount.String())
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
		NumShares:             sdk.NewDec(100),
		AccumValuePerShare:    emptyCoins,
		UnclaimedRewardsTotal: emptyCoins,
	}

	positionOneV2 = accumPackage.Record{
		NumShares:             sdk.NewDec(150),
		AccumValuePerShare:    emptyCoins,
		UnclaimedRewardsTotal: emptyCoins,
	}

	positionTwo = accumPackage.Record{
		NumShares:             sdk.NewDec(200),
		AccumValuePerShare:    emptyCoins,
		UnclaimedRewardsTotal: emptyCoins,
	}

	positionThree = accumPackage.Record{
		NumShares:             sdk.NewDec(300),
		AccumValuePerShare:    emptyCoins,
		UnclaimedRewardsTotal: emptyCoins,
	}

	validPositionName   = testAddressThree
	invalidPositionName = testAddressTwo
	negativeCoins       = sdk.DecCoins{sdk.DecCoin{Denom: initialCoinsDenomOne[0].Denom, Amount: sdk.OneDec().Neg()}}
)

func withInitialAccumValue(record accumPackage.Record, initialAccum sdk.DecCoins) accumPackage.Record {
	record.AccumValuePerShare = initialAccum
	return record
}

func withUnclaimedRewards(record accumPackage.Record, unclaimedRewards sdk.DecCoins) accumPackage.Record {
	record.UnclaimedRewardsTotal = unclaimedRewards
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
			accumName:  "spread-reward-accumulator",
			expSetPass: true,
			expGetPass: true,
		},
		{
			testName:   "create duplicate accumulator",
			accumName:  "spread-reward-accumulator",
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

func (suite *AccumTestSuite) TestMakeAccumulatorWithValueAndShares() {
	// We set up store once at beginning so we can test duplicates
	suite.SetupTest()

	type testcase struct {
		testName    string
		accumName   string
		accumValue  sdk.DecCoins
		totalShares sdk.Dec
		expAccum    accumPackage.AccumulatorObject
		expSetPass  bool
		expGetPass  bool
	}

	tests := []testcase{
		{
			testName:    "create valid accumulator",
			accumName:   "spread-reward-accumulator",
			accumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(20))),
			totalShares: sdk.NewDec(30),
			expSetPass:  true,
			expGetPass:  true,
		},
		{
			testName:    "create duplicate accumulator",
			accumName:   "spread-reward-accumulator",
			accumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10)), sdk.NewDecCoin("bar", sdk.NewInt(20))),
			totalShares: sdk.NewDec(30),
			expSetPass:  false,
			expGetPass:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.testName, func() {
			// Creates raw accumulator object with test case's accum name and zero initial value
			expAccum := accumPackage.MakeTestAccumulator(suite.store, tc.accumName, emptyCoins, emptyDec)

			err := accumPackage.MakeAccumulatorWithValueAndShare(suite.store, tc.accumName, tc.accumValue, tc.totalShares)

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
	defaultAccObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, emptyCoins, emptyDec)

	nonEmptyAccObject := accumPackage.MakeTestAccumulator(suite.store, testNameTwo, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		accObject        *accumPackage.AccumulatorObject
		name             string
		numShareUnits    sdk.Dec
		options          *accumPackage.Options
		expectedPosition accumPackage.Record
	}{
		"test address one - position created": {
			accObject:        defaultAccObject,
			name:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			expectedPosition: positionOne,
		},
		"test address two (non-nil options) - position created": {
			accObject:        defaultAccObject,
			name:             testAddressTwo,
			numShareUnits:    positionTwo.NumShares,
			expectedPosition: positionTwo,
			options:          &emptyPositionOptions,
		},
		"test address one - position overwritten": {
			accObject:        defaultAccObject,
			name:             testAddressOne,
			numShareUnits:    positionOneV2.NumShares,
			expectedPosition: positionOneV2,
		},
		"test address three - added": {
			accObject:        defaultAccObject,
			name:             testAddressThree,
			numShareUnits:    positionThree.NumShares,
			expectedPosition: positionThree,
		},
		"test address one with non-empty accumulator - position created": {
			accObject:        nonEmptyAccObject,
			name:             testAddressOne,
			numShareUnits:    positionOne.NumShares,
			expectedPosition: withInitialAccumValue(positionOne, initialCoinsDenomOne),
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			originalAccValue := tc.accObject.GetTotalShareField()
			expectedAccValue := originalAccValue.Add(tc.numShareUnits)

			// System under test.
			err := tc.accObject.NewPosition(tc.name, tc.numShareUnits, tc.options)
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, position.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, position.UnclaimedRewardsTotal)

			// ensure receiver was mutated
			suite.Require().Equal(expectedAccValue, tc.accObject.GetTotalShareField())
			// ensure state was mutated
			suite.TotalSharesCheck(tc.accObject, expectedAccValue)

			if tc.options == nil {
				suite.Require().Nil(position.Options)
				return
			}

			suite.Require().Equal(*tc.options, *position.Options)
		})
	}
}

func (suite *AccumTestSuite) TestNewPositionIntervalAccumulation() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	defaultAccObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		accObject                    *accumPackage.AccumulatorObject
		name                         string
		numShareUnits                sdk.Dec
		intervalAccumulationPerShare sdk.DecCoins
		options                      *accumPackage.Options
		expectedPosition             accumPackage.Record
		expectedError                error
	}{
		"interval acc value equals to acc": {
			accObject:                    defaultAccObject,
			name:                         testAddressOne,
			numShareUnits:                positionOne.NumShares,
			intervalAccumulationPerShare: defaultAccObject.GetValue(),
			expectedPosition: accumPackage.Record{
				NumShares:             positionOne.NumShares,
				AccumValuePerShare:    defaultAccObject.GetValue(),
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
		"interval acc value does not equal to acc": {
			accObject:                    defaultAccObject,
			name:                         testAddressTwo,
			numShareUnits:                positionTwo.NumShares,
			intervalAccumulationPerShare: defaultAccObject.GetValue().MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:             positionTwo.NumShares,
				AccumValuePerShare:    defaultAccObject.GetValue().MulDec(sdk.NewDec(2)),
				UnclaimedRewardsTotal: emptyCoins,
			},
			options: &emptyPositionOptions,
		},
		"negative acc value - error": {
			accObject:                    defaultAccObject,
			name:                         testAddressOne,
			numShareUnits:                positionOne.NumShares,
			intervalAccumulationPerShare: defaultAccObject.GetValue().MulDec(sdk.NewDec(-1)),
			expectedError:                accumPackage.NegativeIntervalAccumulationPerShareError{defaultAccObject.GetValue().MulDec(sdk.NewDec(-1))},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			originalAccValue := tc.accObject.GetTotalShareField()
			expectedAccValue := originalAccValue.Add(tc.numShareUnits)

			// System under test.
			err := tc.accObject.NewPositionIntervalAccumulation(tc.name, tc.numShareUnits, tc.intervalAccumulationPerShare, tc.options)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, position.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, position.UnclaimedRewardsTotal)

			// ensure receiver was mutated
			suite.Require().Equal(expectedAccValue, tc.accObject.GetTotalShareField())
			// ensure state was mutated
			suite.TotalSharesCheck(tc.accObject, expectedAccValue)
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
		accObject             *accumPackage.AccumulatorObject
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
			actualResult, _, err := tc.accObject.ClaimRewards(tc.accName)

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
				suite.Require().Equal(emptyCoins, finalPosition.UnclaimedRewardsTotal)
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
				accumPackage.InitOrUpdatePosition(curAccum, curAccum.GetValue(), positionName, tc.startingNumShares, tc.startingUnclaimedRewards, nil)
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
				suite.Require().Equal(tc.accumInit.Add(tc.expAccumDelta...), newPosition.AccumValuePerShare)

				// Ensure accrued rewards are moved into UnclaimedRewardsTotal (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.expAccumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewardsTotal)

				// Ensure address's position properly reflects new number of shares
				suite.Require().Equal(tc.startingNumShares.Add(tc.newShares), newPosition.NumShares)

				// Ensure a new position isn't created or removed from memory
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)

				// Note: only new because we create a raw accumulator.
				expectedAccValue := tc.newShares
				// ensure receiver was mutated
				suite.Require().Equal(expectedAccValue, curAccum.GetTotalShareField())
				// ensure state was mutated
				suite.TotalSharesCheck(curAccum, expectedAccValue)
			} else {
				suite.Require().Error(err)

				// Further checks to ensure state was not mutated upon error
				if !tc.addrDoesNotExist {
					// Get new position for comparison
					newPosition, err := accumPackage.GetPosition(curAccum, positionName)
					suite.Require().NoError(err)

					// Ensure that numShares, accumulator value, and unclaimed rewards are unchanged
					suite.Require().Equal(tc.startingNumShares, newPosition.NumShares)
					suite.Require().Equal(tc.accumInit, newPosition.AccumValuePerShare)
					suite.Require().Equal(tc.startingUnclaimedRewards, newPosition.UnclaimedRewardsTotal)
				}
			}
		})
	}
}

// TestAddToPositionIntervalAccumulation this test only focuses on testing the
// interval accumulation functionality of adding to position.
func (suite *AccumTestSuite) TestAddToPositionIntervalAccumulation() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

	tests := map[string]struct {
		accObject                    *accumPackage.AccumulatorObject
		name                         string
		numShareUnits                sdk.Dec
		intervalAccumulationPerShare sdk.DecCoins
		expectedPosition             accumPackage.Record
		expectedError                error
	}{
		"interval acc value equals to acc": {
			accObject:                    accObject,
			name:                         testAddressOne,
			numShareUnits:                positionOne.NumShares,
			intervalAccumulationPerShare: accObject.GetValue(),
			expectedPosition: accumPackage.Record{
				NumShares:             positionOne.NumShares,
				AccumValuePerShare:    accObject.GetValue(),
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
		"interval acc value does not equal to acc": {
			accObject:                    accObject,
			name:                         testAddressTwo,
			numShareUnits:                positionTwo.NumShares,
			intervalAccumulationPerShare: accObject.GetValue().MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:             positionTwo.NumShares,
				AccumValuePerShare:    accObject.GetValue().MulDec(sdk.NewDec(2)),
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			expectedAccValue := tc.accObject.GetTotalShareField()
			expectedAccValue = expectedAccValue.Add(tc.numShareUnits)

			// Setup
			err := tc.accObject.NewPositionIntervalAccumulation(tc.name, sdk.ZeroDec(), tc.accObject.GetValue(), nil)
			suite.Require().NoError(err)

			// System under test.
			err = tc.accObject.AddToPositionIntervalAccumulation(tc.name, tc.numShareUnits, tc.intervalAccumulationPerShare)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, position.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, position.UnclaimedRewardsTotal)
			suite.Require().Nil(position.Options)

			// ensure receiver was mutated
			suite.Require().Equal(expectedAccValue, tc.accObject.GetTotalShareField())
			// ensure state was mutated
			suite.TotalSharesCheck(tc.accObject, expectedAccValue)
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
				accumPackage.InitOrUpdatePosition(curAccum, curAccum.GetValue(), positionName, tc.startingNumShares, tc.startingUnclaimedRewards, nil)
			}

			// Update accumulator with expAccumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit.Add(tc.expAccumDelta...), tc.startingNumShares)
			expectedGlobalNumShares := tc.startingNumShares.Sub(tc.removedShares)

			// Remove removedShares from position
			err := curAccum.RemoveFromPosition(positionName, tc.removedShares)

			if tc.expPass {
				suite.Require().NoError(err)

				// Get updated position for comparison
				newPosition, err := accumPackage.GetPosition(curAccum, positionName)
				suite.Require().NoError(err)

				// Ensure position's accumulator value is moved up to init + delta
				suite.Require().Equal(tc.accumInit.Add(tc.expAccumDelta...), newPosition.AccumValuePerShare)

				// Ensure accrued rewards are moved into UnclaimedRewardsTotal (both when it starts empty and not)
				suite.Require().Equal(tc.startingUnclaimedRewards.Add(tc.expAccumDelta.MulDec(tc.startingNumShares)...), newPosition.UnclaimedRewardsTotal)

				// Ensure address's position properly reflects new number of shares
				if (tc.startingNumShares.Sub(tc.removedShares)).IsZero() {
					suite.Require().Equal(emptyDec, newPosition.NumShares)
				} else {
					suite.Require().Equal(tc.startingNumShares.Sub(tc.removedShares), newPosition.NumShares)
				}

				// Ensure a new position isn't created in memory (only old one is overwritten)
				allAccumPositions, err := curAccum.GetAllPositions()
				suite.Require().NoError(err)
				suite.Require().True(len(allAccumPositions) == 1)

				// ensure receiver was mutated
				suite.Require().Equal(expectedGlobalNumShares, curAccum.GetTotalShareField())
				// ensure state was mutated
				suite.TotalSharesCheck(curAccum, expectedGlobalNumShares)
			} else {
				suite.Require().Error(err)

				// Further checks to ensure state was not mutated upon error
				if !tc.addrDoesNotExist {
					// Get new position for comparison
					newPosition, err := accumPackage.GetPosition(curAccum, positionName)
					suite.Require().NoError(err)

					// Ensure that numShares, accumulator value, and unclaimed rewards are unchanged
					suite.Require().Equal(tc.startingNumShares, newPosition.NumShares)
					suite.Require().Equal(tc.accumInit, newPosition.AccumValuePerShare)
					suite.Require().Equal(tc.startingUnclaimedRewards, newPosition.UnclaimedRewardsTotal)
				}
			}
		})
	}
}

// TestRemoveFromPositionIntervalAccumulation this test only focuses on testing the
// custom accumulator value functionality of removing from a position.
func (suite *AccumTestSuite) TestRemoveFromPositionIntervalAccumulation() {
	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	baseAccumValue := initialCoinsDenomOne

	// Setup.
	accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, baseAccumValue, emptyDec)

	tests := map[string]struct {
		accObject                    *accumPackage.AccumulatorObject
		name                         string
		numShareUnits                sdk.Dec
		intervalAccumulationPerShare sdk.DecCoins
		expectedPosition             accumPackage.Record
		expectedError                error
	}{
		"interval acc value equals to acc": {
			accObject:                    accObject,
			name:                         testAddressOne,
			numShareUnits:                positionOne.NumShares,
			intervalAccumulationPerShare: baseAccumValue,
			expectedPosition: accumPackage.Record{
				NumShares:          sdk.ZeroDec(),
				AccumValuePerShare: baseAccumValue,
				// base value - 0.5 * base = base value
				UnclaimedRewardsTotal: baseAccumValue.MulDec(sdk.NewDecWithPrec(5, 1)).MulDec(positionOne.NumShares),
			},
		},
		"interval acc value does not equal to acc": {
			accObject:                    accObject,
			name:                         testAddressTwo,
			numShareUnits:                positionTwo.NumShares,
			intervalAccumulationPerShare: baseAccumValue.MulDec(sdk.NewDecWithPrec(75, 2)),
			expectedPosition: accumPackage.Record{
				NumShares:          sdk.ZeroDec(),
				AccumValuePerShare: baseAccumValue.MulDec(sdk.NewDecWithPrec(75, 2)),
				// base value - 0.75 * base = 0.25 * base
				UnclaimedRewardsTotal: baseAccumValue.MulDec(sdk.NewDecWithPrec(25, 2)).MulDec(positionTwo.NumShares),
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			// Setup

			expectedGlobalAccValue := tc.accObject.GetTotalShareField()

			// Original position's accum is always set to 0.5 * base value.
			err := tc.accObject.NewPositionIntervalAccumulation(tc.name, tc.numShareUnits, initialCoinsDenomOne.MulDec(sdk.NewDecWithPrec(5, 1)), nil)
			suite.Require().NoError(err)

			tc.accObject.SetValue(tc.intervalAccumulationPerShare)

			// System under test.
			err = tc.accObject.RemoveFromPositionIntervalAccumulation(tc.name, tc.numShareUnits, tc.intervalAccumulationPerShare)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			// Assertions.
			position := tc.accObject.MustGetPosition(tc.name)

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, position.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, position.UnclaimedRewardsTotal)
			suite.Require().Nil(position.Options)

			// ensure receiver was mutated
			suite.Require().Equal(expectedGlobalAccValue.String(), tc.accObject.GetTotalShareField().String())
			// ensure state was mutated
			suite.TotalSharesCheck(tc.accObject, expectedGlobalAccValue)
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
				accumPackage.InitOrUpdatePosition(curAccum, curAccum.GetValue(), positionName, tc.numShares, sdk.NewDecCoins(), nil)
			}

			// Update accumulator with expAccumDelta (increasing position's rewards by a proportional amount)
			curAccum = accumPackage.MakeTestAccumulator(suite.store, testNameOne, tc.accumInit.Add(tc.expAccumDelta...), emptyDec)

			// Get position size from valid address (or from nonexistant if address does not exist)
			positionSize, err := curAccum.GetPositionSize(positionName)

			if tc.changedShares.IsPositive() {
				accumPackage.InitOrUpdatePosition(curAccum, curAccum.GetValue(), positionName, tc.numShares.Add(tc.changedShares), sdk.NewDecCoins(), nil)
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
		AccumValuePerShare: sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denomOne, sdk.OneDec()),
		),
		UnclaimedRewardsTotal: sdk.NewDecCoins(
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

			originalAccum := suite.MakeAndGetAccumulator(testNameOne)

			// System under test.
			originalAccum.AddToAccumulator(tc.updateAmount)

			// Validations.

			// validate that the reciever is mutated.
			suite.Require().Equal(tc.expectedValue, originalAccum.GetValue())

			accumFromStore, err := accumPackage.GetAccumulator(suite.store, testNameOne)
			suite.Require().NoError(err)

			// validate that store is updated.
			suite.Require().Equal(tc.expectedValue, accumFromStore.GetValue())
			// ensure receiver was mutated
			suite.Require().Equal(tc.expectedValue.String(), originalAccum.GetValueField().String())
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
				NumShares:             sdk.OneDec().MulInt64(2),
				AccumValuePerShare:    initialCoinsDenomOne,
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
		"negative - acts as RemoveFromPosition": {
			name:      testAddressOne,
			numShares: sdk.OneDec().Neg(),

			expectedPosition: accumPackage.Record{
				NumShares:             sdk.ZeroDec(),
				AccumValuePerShare:    initialCoinsDenomOne,
				UnclaimedRewardsTotal: emptyCoins,
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

			expectedGlobalAccValue := accObject.GetTotalShareField().Add(tc.numShares).Add(sdk.OneDec())

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
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, updatedPosition.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, updatedPosition.UnclaimedRewardsTotal)
			suite.Require().Nil(position.Options)

			// ensure receiver was mutated
			suite.Require().Equal(expectedGlobalAccValue.String(), accObject.GetTotalShareField().String())
			// ensure state was mutated
			suite.TotalSharesCheck(accObject, expectedGlobalAccValue)
		})
	}
}

// TestUpdatePositionIntervalAccumulation this test only focuses on testing the
// custom accumulator value functionality of updating a position.
func (suite *AccumTestSuite) TestUpdatePositionIntervalAccumulation() {
	tests := []struct {
		testName                     string
		initialShares                sdk.Dec
		initialAccum                 sdk.DecCoins
		accName                      string
		numShareUnits                sdk.Dec
		intervalAccumulationPerShare sdk.DecCoins
		expectedPosition             accumPackage.Record
		expectedError                error
	}{
		{
			testName:                     "interval acc value equals to acc; positive shares -> acts as AddToPosition",
			initialShares:                sdk.ZeroDec(),
			initialAccum:                 initialCoinsDenomOne,
			accName:                      testAddressOne,
			numShareUnits:                positionOne.NumShares,
			intervalAccumulationPerShare: initialCoinsDenomOne,
			expectedPosition: accumPackage.Record{
				NumShares:             positionOne.NumShares,
				AccumValuePerShare:    initialCoinsDenomOne,
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
		{
			testName:                     "interval acc value does not equal to acc; remove same amount -> acts as RemoveFromPosition",
			initialShares:                positionTwo.NumShares,
			initialAccum:                 initialCoinsDenomOne,
			accName:                      testAddressTwo,
			numShareUnits:                positionTwo.NumShares.Neg(), // note: negative shares
			intervalAccumulationPerShare: initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:             sdk.ZeroDec(), // results in 0 shares (200 - 200)
				AccumValuePerShare:    initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
		{
			testName:                     "interval acc value does not equal to acc; remove diff amount -> acts as RemoveFromPosition",
			initialShares:                positionTwo.NumShares,
			initialAccum:                 initialCoinsDenomOne,
			accName:                      testAddressTwo,
			numShareUnits:                positionOne.NumShares.Neg(), // note: negative shares
			intervalAccumulationPerShare: initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
			expectedPosition: accumPackage.Record{
				NumShares:             positionOne.NumShares, // results in 100 shares (200 - 100)
				AccumValuePerShare:    initialCoinsDenomOne.MulDec(sdk.NewDec(2)),
				UnclaimedRewardsTotal: emptyCoins,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.testName, func() {
			suite.SetupTest()

			// make accumualtor based off of tc.accObject
			accumObject := suite.MakeAndGetAccumulator(testNameOne)

			expectedGlobalAccValue := tc.initialShares.Add(tc.numShareUnits)

			// manually update accumulator value
			accumObject.AddToAccumulator(initialCoinsDenomOne)

			// Setup
			err := accumObject.NewPositionIntervalAccumulation(tc.accName, tc.initialShares, tc.initialAccum, nil)
			suite.Require().NoError(err)

			// System under test.
			err = accumObject.UpdatePositionIntervalAccumulation(tc.accName, tc.numShareUnits, tc.intervalAccumulationPerShare)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			accumObject = suite.GetAccumulator(testNameOne)
			position := accumObject.MustGetPosition(tc.accName)
			// Assertions.

			suite.Require().Equal(tc.expectedPosition.NumShares, position.NumShares)
			suite.Require().Equal(tc.expectedPosition.AccumValuePerShare, position.AccumValuePerShare)
			suite.Require().Equal(tc.expectedPosition.UnclaimedRewardsTotal, position.UnclaimedRewardsTotal)
			suite.Require().Nil(position.Options)

			// ensure receiver was mutated
			suite.Require().Equal(expectedGlobalAccValue.String(), accumObject.GetTotalShareField().String())
			// ensure state was mutated
			suite.TotalSharesCheck(accumObject, expectedGlobalAccValue)
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

			hasPosition := accObject.HasPosition(defaultPositionName)
			suite.Equal(tc.preCreatePosition, hasPosition)
		})
	}
}

func (suite *AccumTestSuite) TestSetPositionIntervalAccumulation() {
	// We setup store and accum
	// once at beginning.
	suite.SetupTest()

	// Setup.
	var (
		accObject = accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)
	)

	tests := map[string]struct {
		positionName                 string
		intervalAccumulationPerShare sdk.DecCoins
		expectedError                error
	}{
		"valid update greater than initial value": {
			positionName:                 validPositionName,
			intervalAccumulationPerShare: initialCoinsDenomOne.Add(initialCoinDenomOne),
		},
		"valid update equal to the initial value": {
			positionName:                 validPositionName,
			intervalAccumulationPerShare: initialCoinsDenomOne,
		},
		"invalid position - different name": {
			positionName:  invalidPositionName,
			expectedError: accumPackage.NoPositionError{Name: invalidPositionName},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {

			// Setup
			err := accObject.NewPositionIntervalAccumulation(validPositionName, sdk.OneDec(), initialCoinsDenomOne, nil)
			suite.Require().NoError(err)

			// System under test.
			err = accObject.SetPositionIntervalAccumulation(tc.positionName, tc.intervalAccumulationPerShare)

			// Assertions.
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			position := accObject.MustGetPosition(tc.positionName)
			suite.Require().Equal(tc.intervalAccumulationPerShare, position.GetAccumValuePerShare())
			// unchanged
			suite.Require().Equal(sdk.OneDec(), position.NumShares)
			suite.Require().Equal(emptyCoins, position.GetUnclaimedRewardsTotal())
		})
	}
}

// We run a series of partially random operations on two accumulators to ensure that total shares are properly tracked in state
func (suite *AccumTestSuite) TestGetTotalShares() {
	suite.SetupTest()

	// Set seed to make tests deterministic
	rand.Seed(1)

	// Set up two accumulators to ensure we can test all relevant behavior
	accumOne, accumTwo := suite.MakeAndGetAccumulator(testNameOne), suite.MakeAndGetAccumulator(testNameTwo)

	// Sanity check initial accumulators (start at zero shares)
	suite.TotalSharesCheck(accumOne, sdk.ZeroDec())
	suite.TotalSharesCheck(accumTwo, sdk.ZeroDec())

	// Create position on first accum and pull new accum objects from state
	err := accumOne.NewPosition(testAddressOne, sdk.OneDec(), nil)
	suite.Require().NoError(err)
	accumOne, accumTwo = suite.GetAccumulator(testNameOne), suite.GetAccumulator(testNameTwo)

	// Check that total shares for accum one has updated properly and accum two shares are unchanged
	suite.TotalSharesCheck(accumOne, sdk.OneDec())
	suite.TotalSharesCheck(accumTwo, sdk.ZeroDec())

	// Run a number of NewPosition, AddToPosition, and RemoveFromPosition operations on each accum
	testAddresses := []string{testAddressOne, testAddressTwo, testAddressThree}
	accums := []*accumPackage.AccumulatorObject{accumOne, accumTwo}
	expectedShares := []sdk.Dec{sdk.OneDec(), sdk.ZeroDec()}

	for i := 1; i <= 10; i++ {
		// Cycle through accounts and accumulators
		curAddr := testAddresses[i%3]
		curAccum := accums[i%2]

		// We set a baseAmt that varies with the iteration to increase coverage
		baseAmt := sdk.NewDec(int64(i)).Mul(sdk.NewDec(10))

		// If addr doesn't have a position yet, we make one
		positionExists := curAccum.HasPosition(curAddr)
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
		if addAmt.IsPositive() {
			err = curAccum.AddToPosition(curAddr, addAmt)
			suite.Require().NoError(err)
		}

		// Half the time, we remove one share from the position
		amtToRemove := sdk.OneDec().Mul(removeShares)
		if amtToRemove.IsPositive() {
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
	accumOne, accumTwo = suite.GetAccumulator(testNameOne), suite.GetAccumulator(testNameTwo)

	// Ensure that total shares in each accum matches our expected number of shares
	suite.TotalSharesCheck(accumOne, expectedShares[0])
	suite.TotalSharesCheck(accumTwo, expectedShares[1])
}

func (suite *AccumTestSuite) TestAddToUnclaimedRewards() {
	// We setup store and accum
	// once at beginning.
	suite.SetupTest()

	// Setup.
	var (
		accObject = accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)
	)

	tests := map[string]struct {
		positionName             string
		unclaimedRewardsAddition sdk.DecCoins
		expectedError            error
	}{
		"valid update": {
			positionName:             validPositionName,
			unclaimedRewardsAddition: initialCoinsDenomOne,
		},
		"zero rewards - no op": {
			positionName:             validPositionName,
			unclaimedRewardsAddition: emptyCoins,
		},
		"error: negative addition": {
			positionName:             validPositionName,
			unclaimedRewardsAddition: negativeCoins,
			expectedError:            accumPackage.NegativeRewardsAdditionError{AccumName: accObject.GetName(), PositionName: validPositionName},
		},
		"invalid position - different name": {
			positionName:  invalidPositionName,
			expectedError: accumPackage.NoPositionError{Name: invalidPositionName},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			err := accObject.NewPositionIntervalAccumulation(validPositionName, sdk.OneDec(), initialCoinsDenomOne, nil)
			suite.Require().NoError(err)

			// Update global accumulator.
			accObject.AddToAccumulator(initialCoinsDenomOne)

			// System under test.
			err = accObject.AddToUnclaimedRewards(tc.positionName, tc.unclaimedRewardsAddition)

			// Assertions.
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().Equal(tc.expectedError, err)
				return
			}
			suite.Require().NoError(err)

			position := accObject.MustGetPosition(tc.positionName)
			suite.Require().Equal(initialCoinsDenomOne, position.GetAccumValuePerShare())
			//
			suite.Require().Equal(tc.unclaimedRewardsAddition, position.GetUnclaimedRewardsTotal())
		})
	}
}

func (suite *AccumTestSuite) TestDeletePosition() {
	tests := map[string]struct {
		positionName             string
		globalGrowth             sdk.DecCoins
		numShares                int64
		expectedUnclaimedRewards sdk.DecCoins
		expectedError            error
	}{
		"base": {
			positionName:             validPositionName,
			globalGrowth:             initialCoinsDenomOne,
			numShares:                1,
			expectedUnclaimedRewards: initialCoinsDenomOne,
		},
		"no global growth": {
			positionName:             validPositionName,
			globalGrowth:             emptyCoins,
			numShares:                1,
			expectedUnclaimedRewards: emptyCoins,
		},
		"2 shares": {
			positionName:             validPositionName,
			globalGrowth:             initialCoinsDenomOne,
			numShares:                2,
			expectedUnclaimedRewards: initialCoinsDenomOne.Add(initialCoinsDenomOne...),
		},
		"invalid position - different name": {
			positionName:  invalidPositionName,
			numShares:     1,
			expectedError: accumPackage.NoPositionError{Name: invalidPositionName},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			// Setup.
			accObject := accumPackage.MakeTestAccumulator(suite.store, testNameOne, initialCoinsDenomOne, emptyDec)

			err := accObject.NewPosition(validPositionName, sdk.NewDec(tc.numShares), nil)
			suite.Require().NoError(err)

			// Update global accumulator.
			accObject.AddToAccumulator(tc.globalGrowth)

			// System under test.
			unclaimedRewards, err := accObject.DeletePosition(tc.positionName)

			// Assertions.
			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectedError)
				return
			}
			suite.Require().NoError(err)

			hasPosition := accObject.HasPosition(tc.positionName)
			suite.Require().False(hasPosition)

			// Check rewards.
			suite.Require().Equal(tc.expectedUnclaimedRewards, unclaimedRewards)

			// Check that global accumulator value is updated
			suite.TotalSharesCheck(accObject, sdk.ZeroDec())
			// Check that accumulator receiver is mutated.
			suite.Require().Equal(sdk.ZeroDec().String(), accObject.GetTotalShareField().String())
		})
	}
}
