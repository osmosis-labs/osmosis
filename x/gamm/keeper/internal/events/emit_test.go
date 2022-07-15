package events_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

type GammEventsTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	addressString = "addr1---------------"
	testDenomA    = "denoma"
	testDenomB    = "denomb"
	testDenomC    = "denomc"
	testDenomD    = "denomd"
)

func TestGammEventsTestSuite(t *testing.T) {
	suite.Run(t, new(GammEventsTestSuite))
}

func (suite *GammEventsTestSuite) TestEmitSwapEvent() {
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
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, sdk.NewInt(1234))),
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomB, sdk.NewInt(5678))),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
		"valid with multiple tokens in and out": {
			ctx:             suite.CreateTestContext(),
			testAccountAddr: sdk.AccAddress([]byte(addressString)),
			poolId:          200,
			tokensIn:        sdk.NewCoins(sdk.NewCoin(testDenomA, sdk.NewInt(12)), sdk.NewCoin(testDenomB, sdk.NewInt(99))),
			tokensOut:       sdk.NewCoins(sdk.NewCoin(testDenomC, sdk.NewInt(88)), sdk.NewCoin(testDenomD, sdk.NewInt(34))),
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
