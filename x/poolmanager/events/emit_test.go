package events_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/events"
)

type PoolManagerEventsTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	addressString = "addr1---------------"
	testDenomA    = "denoma"
	testDenomB    = "denomb"
	testDenomC    = "denomc"
	testDenomD    = "denomd"
)

func TestPoolManagerEventsTestSuite(t *testing.T) {
	suite.Run(t, new(PoolManagerEventsTestSuite))
}

func (suite *PoolManagerEventsTestSuite) TestEmitSwapEvent() {
	testcases := map[string]struct {
		ctx             sdk.Context
		testAccountAddr sdk.AccAddress
		poolId          uint64
		tokensIn        sdk.Coins
		tokensOut       sdk.Coins
	}{
		"basic valid": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          1,
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(1234))),
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomB, osmomath.NewInt(5678))),
		},
		"valid with multiple tokens in and out": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          200,
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(12)), sdk.NewCoin(testDenomB, osmomath.NewInt(99))),
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomC, osmomath.NewInt(88)), sdk.NewCoin(testDenomD, osmomath.NewInt(34))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtTokenSwapped,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(sdk.AttributeKeySender, tc.testAccountAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(tc.poolId, 10)),
					sdk.NewAttribute(types.AttributeKeyTokensIn, tc.tokensIn.String()),
					sdk.NewAttribute(types.AttributeKeyTokensOut, tc.tokensOut.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSwapEvent(tc.ctx, tc.testAccountAddr, tc.poolId, tc.tokensIn, tc.tokensOut)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *PoolManagerEventsTestSuite) TestEmitAddLiquidityEvent() {
	testcases := map[string]struct {
		ctx             sdk.Context
		testAccountAddr sdk.AccAddress
		poolId          uint64
		tokensIn        sdk.Coins
	}{
		"basic valid": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          1,
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(1234))),
		},
		"valid with multiple tokens in": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          200,
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(12)), sdk.NewCoin(testDenomB, osmomath.NewInt(99))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtPoolJoined,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(sdk.AttributeKeySender, tc.testAccountAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(tc.poolId, 10)),
					sdk.NewAttribute(types.AttributeKeyTokensIn, tc.tokensIn.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitAddLiquidityEvent(tc.ctx, tc.testAccountAddr, tc.poolId, tc.tokensIn)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *PoolManagerEventsTestSuite) TestEmitRemoveLiquidityEvent() {
	testcases := map[string]struct {
		ctx             sdk.Context
		testAccountAddr sdk.AccAddress
		poolId          uint64
		tokensOut       sdk.Coins
	}{
		"basic valid": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          1,
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(1234))),
		},
		"valid with multiple tokens out": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          200,
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(12)), sdk.NewCoin(testDenomB, osmomath.NewInt(99))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtPoolExited,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(sdk.AttributeKeySender, tc.testAccountAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(tc.poolId, 10)),
					sdk.NewAttribute(types.AttributeKeyTokensOut, tc.tokensOut.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitRemoveLiquidityEvent(tc.ctx, tc.testAccountAddr, tc.poolId, tc.tokensOut)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}
