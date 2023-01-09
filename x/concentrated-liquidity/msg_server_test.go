package concentrated_liquidity_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	cl "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// TestCreateConcentratedPool_Events tests that events are correctly emitted
// when calling CreateConcentratedPool.
func (suite *KeeperTestSuite) TestCreateConcentratedPool_Events() {
	testcases := map[string]struct {
		sender                    string
		denom0                    string
		denom1                    string
		tickSpacing               uint64
		precisionFactorAtPriceOne sdk.Int
		expectedPoolCreatedEvent  int
		expectedMessageEvents     int
		expectedError             error
	}{
		"happy path": {
			denom0:                    ETH,
			denom1:                    USDC,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: DefaultExponentAtPriceOne,
			expectedPoolCreatedEvent:  1,
			expectedMessageEvents:     3, // 1 for pool created, 1 for coin spent, 1 for coin received
		},
		"error: missing denom0": {
			denom1:                    USDC,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:             fmt.Errorf("received denom0 with invalid metadata: %s", ""),
		},
		"error: missing denom1": {
			denom0:                    ETH,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:             fmt.Errorf("received denom1 with invalid metadata: %s", ""),
		},
		"error: missing tickSpacing": {
			denom0:                    ETH,
			denom1:                    USDC,
			precisionFactorAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:             fmt.Errorf("tick spacing must be positive"),
		},
		"error: precision value below minimum": {
			denom0:                    ETH,
			denom1:                    USDC,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: cltypes.PrecisionValueAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:             fmt.Errorf("precision factor at price one must be greater than or equal to %s", cltypes.PrecisionValueAtPriceOneMin.String()),
		},
		"error: precision value above maximum": {
			denom0:                    ETH,
			denom1:                    USDC,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: cltypes.PrecisionValueAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:             fmt.Errorf("precision factor at price one must be less than or equal to %s", cltypes.PrecisionValueAtPriceOneMax.String()),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Retrieve the pool creation fee from swaprouter params.
			swaprouterParams := swaproutertypes.DefaultParams()

			// Fund account to pay for the pool creation fee.
			suite.FundAcc(suite.TestAccs[0], swaprouterParams.PoolCreationFee)

			msgServer := cl.NewMsgCreatorServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// set denom metadata
			if tc.denom0 != "" {
				denomMetaData := banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{{
						Denom:    tc.denom0,
						Exponent: 0,
					}},
					Base: tc.denom0,
				}
				suite.App.BankKeeper.SetDenomMetaData(ctx, denomMetaData)
			}
			if tc.denom1 != "" {
				denomMetaData := banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{{
						Denom:    tc.denom1,
						Exponent: 0,
					}},
					Base: tc.denom1,
				}
				suite.App.BankKeeper.SetDenomMetaData(ctx, denomMetaData)
			}

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.CreateConcentratedPool(sdk.WrapSDKContext(ctx), &clmodel.MsgCreateConcentratedPool{
				Sender:                    suite.TestAccs[0].String(),
				Denom0:                    tc.denom0,
				Denom1:                    tc.denom1,
				TickSpacing:               tc.tickSpacing,
				PrecisionFactorAtPriceOne: tc.precisionFactorAtPriceOne,
			})

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtPoolCreated, tc.expectedPoolCreatedEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
			}
		})
	}
}

// TODO: Add test cases for create and withdraw position messages
