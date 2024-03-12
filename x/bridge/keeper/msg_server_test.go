package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgInboundTransfer() {
	var asset = types.Asset{
		SourceChain: "bitcoin",
		Denom:       "wbtc1",
		Precision:   10,
	}

	testCases := []struct {
		name     string
		sender   string
		destAddr string
		asset    types.Asset
		amount   math.Int
		valid    bool
	}{
		{
			name:     "happy path",
			sender:   s.TestAccs[0].String(),
			destAddr: s.TestAccs[1].String(),
			asset:    asset,
			amount:   math.NewInt(100),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))

			// Test inbound transfer message
			_, err := s.msgServer.InboundTransfer(
				sdk.WrapSDKContext(ctx),
				types.NewMsgInboundTransfer(tc.sender, tc.destAddr, tc.asset, tc.amount),
			)

			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, new(types.EventInboundTransfer).String(), 1)
		})
	}
}
