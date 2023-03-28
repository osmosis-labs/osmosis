package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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
			precisionFactorAtPriceOne: cltypes.ExponentAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:             cltypes.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: cltypes.ExponentAtPriceOneMin.Sub(sdk.OneInt()), PrecisionValueAtPriceOneMin: cltypes.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: cltypes.ExponentAtPriceOneMax},
		},
		"error: precision value above maximum": {
			denom0:                    ETH,
			denom1:                    USDC,
			tickSpacing:               DefaultTickSpacing,
			precisionFactorAtPriceOne: cltypes.ExponentAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:             cltypes.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: cltypes.ExponentAtPriceOneMax.Add(sdk.OneInt()), PrecisionValueAtPriceOneMin: cltypes.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: cltypes.ExponentAtPriceOneMax},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Retrieve the pool creation fee from poolmanager params.
			poolmanagerParams := poolmanagertypes.DefaultParams()

			// Fund account to pay for the pool creation fee.
			suite.FundAcc(suite.TestAccs[0], poolmanagerParams.PoolCreationFee)

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
				SwapFee:                   DefaultZeroSwapFee,
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

// TestCollectFees_Events tests that events are correctly emitted
// when calling CollectFees.
func (suite *KeeperTestSuite) TestCollectFees_Events() {
	testcases := map[string]struct {
		upperTick                int64
		lowerTick                int64
		positionId               uint64
		expectedCollectFeesEvent int
		expectedMessageEvents    int
		expectedError            error
		errorFromValidateBasic   error
	}{
		"happy path": {
			upperTick:                DefaultUpperTick,
			lowerTick:                DefaultLowerTick,
			positionId:               DefaultPositionId,
			expectedCollectFeesEvent: 1,
			expectedMessageEvents:    2, // 1 for collect fees, 1 for message
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			suite.SetupDefaultPosition(pool.GetId())

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			msg := &cltypes.MsgCollectFees{
				Sender:     suite.TestAccs[0].String(),
				PositionId: tc.positionId,
			}

			response, err := msgServer.CollectFees(sdk.WrapSDKContext(ctx), msg)

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtCollectFees, tc.expectedCollectFeesEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
			}

			// Some validate basic checks are defense in depth so they would normally not be possible to reach
			// This check allows us to still test these cases
			if tc.errorFromValidateBasic != nil {
				suite.Require().Error(msg.ValidateBasic())
				suite.Require().ErrorAs(msg.ValidateBasic(), &tc.errorFromValidateBasic)
			}
		})
	}
}

// TestCollectIncentives_Events tests that events are correctly emitted
// when calling CollectIncentives.
func (suite *KeeperTestSuite) TestCollectIncentives_Events() {
	uptimeHelper := getExpectedUptimes()
	testcases := map[string]struct {
		upperTick                      int64
		lowerTick                      int64
		positionId                     uint64
		expectedCollectIncentivesEvent int
		expectedMessageEvents          int
		expectedError                  error
		errorFromValidateBasic         error
	}{
		"happy path": {
			upperTick:                      DefaultUpperTick,
			lowerTick:                      DefaultLowerTick,
			positionId:                     DefaultPositionId,
			expectedCollectIncentivesEvent: 1,
			expectedMessageEvents:          2, // 1 for collect incentives, 1 for message
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			suite.SetupDefaultPosition(pool.GetId())

			// Get the position and set the current time to be 7 days after the position was created
			position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(ctx, tc.positionId)
			suite.Require().NoError(err)
			ctx = ctx.WithBlockTime(position.JoinTime.Add(time.Hour * 24 * 7))
			positionAge := ctx.BlockTime().Sub(position.JoinTime)

			// Set up accrued incentives
			err = addToUptimeAccums(ctx, pool.GetId(), suite.App.ConcentratedLiquidityKeeper, uptimeHelper.hundredTokensMultiDenom)
			suite.Require().NoError(err)
			suite.FundAcc(pool.GetAddress(), expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, positionAge, sdk.OneInt()))

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			msg := &cltypes.MsgCollectIncentives{
				Sender:     suite.TestAccs[0].String(),
				PositionId: tc.positionId,
			}

			// System under test
			response, err := msgServer.CollectIncentives(sdk.WrapSDKContext(ctx), msg)

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtCollectIncentives, tc.expectedCollectIncentivesEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, 0)
			}

			// Some validate basic checks are defense in depth so they would normally not be possible to reach
			// This check allows us to still test these cases
			if tc.errorFromValidateBasic != nil {
				suite.Require().Error(msg.ValidateBasic())
				suite.Require().ErrorAs(msg.ValidateBasic(), &tc.errorFromValidateBasic)
			}
		})
	}
}
