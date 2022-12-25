package accum_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	iavlstore "github.com/cosmos/cosmos-sdk/store/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/iavl"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
)

type AccumTestSuite struct {
	suite.Suite

	store store.KVStore
}

var (
	testAddressOne   = sdk.AccAddress([]byte("addr1_______________"))
	testAddressTwo   = sdk.AccAddress([]byte("addr2_______________"))
	testAddressThree = sdk.AccAddress([]byte("addr3_______________"))
)

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
		expAccum   accum.AccumulatorObject
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
			expAccum := accum.CreateRawAccumObject(suite.store, tc.accumName, sdk.DecCoins(nil))

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

func (suite *AccumTestSuite) TestNewPosition() {
	var (
		emptyPositionOptions = accum.PositionOptions{}
		testName             = "myaccum"

		emptyCoins = sdk.DecCoins(nil)

		positionOne = accum.Record{
			NumShares:        sdk.NewDec(100),
			InitAccumValue:   emptyCoins,
			UnclaimedRewards: emptyCoins,
		}

		positionOneV2 = accum.Record{
			NumShares:        sdk.NewDec(150),
			InitAccumValue:   emptyCoins,
			UnclaimedRewards: emptyCoins,
		}

		positionTwo = accum.Record{
			NumShares:        sdk.NewDec(200),
			InitAccumValue:   emptyCoins,
			UnclaimedRewards: emptyCoins,
		}

		positionThree = accum.Record{
			NumShares:        sdk.NewDec(300),
			InitAccumValue:   emptyCoins,
			UnclaimedRewards: emptyCoins,
		}
	)

	// We setup store and accum
	// once at beginning so we can test duplicate positions
	suite.SetupTest()

	// Setup.
	err := accum.MakeAccumulator(suite.store, testName)
	suite.Require().NoError(err)

	accObject, err := accum.GetAccumulator(suite.store, testName)
	suite.Require().NoError(err)

	tests := map[string]struct {
		addr              sdk.AccAddress
		numShareUnits     sdk.Dec
		options           accum.PositionOptions
		expectedPositions []accum.Record
	}{
		"test address one - position created": {
			addr:          testAddressOne,
			numShareUnits: positionOne.NumShares,
			options:       emptyPositionOptions,
			expectedPositions: []accum.Record{
				positionOne,
			},
		},
		"test address two - position created": {
			addr:          testAddressTwo,
			numShareUnits: positionTwo.NumShares,
			options:       emptyPositionOptions,
			expectedPositions: []accum.Record{
				positionOne,
				positionTwo,
			},
		},
		"test address one - position overwritten": {
			addr:          testAddressOne,
			numShareUnits: positionOneV2.NumShares,
			options:       emptyPositionOptions,
			expectedPositions: []accum.Record{
				positionOneV2,
				positionTwo,
			},
		},
		"test address three - added": {
			addr:          testAddressThree,
			numShareUnits: positionThree.NumShares,
			options:       emptyPositionOptions,
			expectedPositions: []accum.Record{
				positionOneV2,
				positionTwo,
				positionThree,
			},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {

			// System under test.
			accObject.NewPosition(tc.addr, tc.numShareUnits, tc.options)

			// Assertions.
			positions, err := accObject.GetAllPositions()
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedPositions, positions)
		})
	}
}
