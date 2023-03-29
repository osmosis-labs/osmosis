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
		sender                   string
		denom0                   string
		denom1                   string
		tickSpacing              uint64
		exponentAtPriceOne       sdk.Int
		expectedPoolCreatedEvent int
		expectedMessageEvents    int
		expectedError            error
	}{
		"happy path": {
			denom0:                   ETH,
			denom1:                   USDC,
			tickSpacing:              DefaultTickSpacing,
			exponentAtPriceOne:       DefaultExponentAtPriceOne,
			expectedPoolCreatedEvent: 1,
			expectedMessageEvents:    3, // 1 for pool created, 1 for coin spent, 1 for coin received
		},
		"error: missing denom0": {
			denom1:             USDC,
			tickSpacing:        DefaultTickSpacing,
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("received denom0 with invalid metadata: %s", ""),
		},
		"error: missing denom1": {
			denom0:             ETH,
			tickSpacing:        DefaultTickSpacing,
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("received denom1 with invalid metadata: %s", ""),
		},
		"error: missing tickSpacing": {
			denom0:             ETH,
			denom1:             USDC,
			exponentAtPriceOne: DefaultExponentAtPriceOne,
			expectedError:      fmt.Errorf("tick spacing must be positive"),
		},
		"error: precision value below minimum": {
			denom0:             ETH,
			denom1:             USDC,
			tickSpacing:        DefaultTickSpacing,
			exponentAtPriceOne: cltypes.ExponentAtPriceOneMin.Sub(sdk.OneInt()),
			expectedError:      cltypes.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: cltypes.ExponentAtPriceOneMin.Sub(sdk.OneInt()), PrecisionValueAtPriceOneMin: cltypes.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: cltypes.ExponentAtPriceOneMax},
		},
		"error: precision value above maximum": {
			denom0:             ETH,
			denom1:             USDC,
			tickSpacing:        DefaultTickSpacing,
			exponentAtPriceOne: cltypes.ExponentAtPriceOneMax.Add(sdk.OneInt()),
			expectedError:      cltypes.ExponentAtPriceOneError{ProvidedExponentAtPriceOne: cltypes.ExponentAtPriceOneMax.Add(sdk.OneInt()), PrecisionValueAtPriceOneMin: cltypes.ExponentAtPriceOneMin, PrecisionValueAtPriceOneMax: cltypes.ExponentAtPriceOneMax},
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
				Sender:             suite.TestAccs[0].String(),
				Denom0:             tc.denom0,
				Denom1:             tc.denom1,
				TickSpacing:        tc.tickSpacing,
				ExponentAtPriceOne: tc.exponentAtPriceOne,
				SwapFee:            DefaultZeroSwapFee,
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
		upperTick                     int64
		lowerTick                     int64
		positionIds                   []uint64
		numPositionsToCreate          int
		expectedTotalCollectFeesEvent int
		expectedCollectFeesEvent      int
		expectedMessageEvents         int
		expectedError                 error
		errorFromValidateBasic        error
	}{
		"single position ID": {
			upperTick:                     DefaultUpperTick,
			lowerTick:                     DefaultLowerTick,
			positionIds:                   []uint64{DefaultPositionId},
			numPositionsToCreate:          1,
			expectedTotalCollectFeesEvent: 1,
			expectedCollectFeesEvent:      1,
			expectedMessageEvents:         2, // 1 for collect fees, 1 for send message
		},
		"two position IDs": {
			upperTick:                     DefaultUpperTick,
			lowerTick:                     DefaultLowerTick,
			positionIds:                   []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate:          2,
			expectedTotalCollectFeesEvent: 1,
			expectedCollectFeesEvent:      2,
			expectedMessageEvents:         3, // 1 for collect fees, 2 for send messages
		},
		"three position IDs": {
			upperTick:                     DefaultUpperTick,
			lowerTick:                     DefaultLowerTick,
			positionIds:                   []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:          3,
			expectedTotalCollectFeesEvent: 1,
			expectedCollectFeesEvent:      3,
			expectedMessageEvents:         4, // 1 for collect fees, 3 for send messages
		},
		"error": {
			upperTick:                     DefaultUpperTick,
			lowerTick:                     DefaultLowerTick,
			positionIds:                   []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:          2,
			expectedTotalCollectFeesEvent: 0,
			expectedCollectFeesEvent:      0,
			expectedMessageEvents:         2, // 2 emitted for send messages
			expectedError:                 cltypes.PositionIdNotFoundError{PositionId: DefaultPositionId + 2},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				suite.SetupDefaultPosition(pool.GetId())
			}

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			msg := &cltypes.MsgCollectFees{
				Sender:      suite.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			response, err := msgServer.CollectFees(sdk.WrapSDKContext(ctx), msg)

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtTotalCollectFees, tc.expectedTotalCollectFeesEvent)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtCollectFees, tc.expectedCollectFeesEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
			}
		})
	}
}

// TestCollectIncentives_Events tests that events are correctly emitted
// when calling CollectIncentives.
func (suite *KeeperTestSuite) TestCollectIncentives_Events() {
	uptimeHelper := getExpectedUptimes()
	testcases := map[string]struct {
		upperTick                           int64
		lowerTick                           int64
		positionIds                         []uint64
		numPositionsToCreate                int
		expectedTotalCollectIncentivesEvent int
		expectedCollectIncentivesEvent      int
		expectedMessageEvents               int
		expectedError                       error
		errorFromValidateBasic              error
	}{
		"single position ID": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId},
			numPositionsToCreate:                1,
			expectedTotalCollectIncentivesEvent: 1,
			expectedCollectIncentivesEvent:      1,
			expectedMessageEvents:               2, // 1 for collect incentives, 1 for send message
		},
		"two position IDs": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate:                2,
			expectedTotalCollectIncentivesEvent: 1,
			expectedCollectIncentivesEvent:      2,
			expectedMessageEvents:               3, // 1 for collect incentives, 2 for send messages
		},
		"three position IDs": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:                3,
			expectedTotalCollectIncentivesEvent: 1,
			expectedCollectIncentivesEvent:      3,
			expectedMessageEvents:               4, // 1 for collect incentives, 3 for send messages
		},
		"error": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:                2,
			expectedTotalCollectIncentivesEvent: 0,
			expectedCollectIncentivesEvent:      0,
			expectedError:                       cltypes.PositionIdNotFoundError{PositionId: DefaultPositionId + 2},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				suite.SetupDefaultPosition(pool.GetId())
			}

			position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(ctx, tc.positionIds[0])
			suite.Require().NoError(err)
			ctx = ctx.WithBlockTime(position.JoinTime.Add(time.Hour * 24 * 7))
			positionAge := ctx.BlockTime().Sub(position.JoinTime)

			// Set up accrued incentives
			err = addToUptimeAccums(ctx, pool.GetId(), suite.App.ConcentratedLiquidityKeeper, uptimeHelper.hundredTokensMultiDenom)
			suite.Require().NoError(err)
			suite.FundAcc(pool.GetAddress(), expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, positionAge, sdk.NewInt(int64(len(tc.positionIds)))))

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			msg := &cltypes.MsgCollectIncentives{
				Sender:      suite.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			// System under test
			response, err := msgServer.CollectIncentives(sdk.WrapSDKContext(ctx), msg)

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtTotalCollectIncentives, tc.expectedTotalCollectIncentivesEvent)
				suite.AssertEventEmitted(ctx, cltypes.TypeEvtCollectIncentives, tc.expectedCollectIncentivesEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
			}
		})
	}
}
