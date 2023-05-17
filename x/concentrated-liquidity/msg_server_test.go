package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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
		expectedPoolCreatedEvent int
		expectedMessageEvents    int
		expectedError            error
	}{
		"happy path": {
			denom0:                   ETH,
			denom1:                   USDC,
			tickSpacing:              DefaultTickSpacing,
			expectedPoolCreatedEvent: 1,
			expectedMessageEvents:    4, // 1 for pool created, 1 for coin spent, 1 for coin received, 1 for after pool create hook
		},
		"error: tickSpacing not positive": {
			denom0:        ETH,
			denom1:        USDC,
			tickSpacing:   0,
			expectedError: fmt.Errorf("tick spacing must be positive"),
		},
		"error: tickSpacing not authorized": {
			denom0:        ETH,
			denom1:        USDC,
			tickSpacing:   DefaultTickSpacing + 1,
			expectedError: types.UnauthorizedTickSpacingError{ProvidedTickSpacing: DefaultTickSpacing + 1, AuthorizedTickSpacings: suite.App.ConcentratedLiquidityKeeper.GetParams(suite.Ctx).AuthorizedTickSpacing},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx

			// Retrieve the pool creation fee from poolmanager params.
			poolmanagerParams := poolmanagertypes.DefaultParams()

			// Fund account to pay for the pool creation fee.
			suite.FundAcc(suite.TestAccs[0], poolmanagerParams.PoolCreationFee)

			msgServer := cl.NewMsgCreatorServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.CreateConcentratedPool(sdk.WrapSDKContext(ctx), &clmodel.MsgCreateConcentratedPool{
				Sender:      suite.TestAccs[0].String(),
				Denom0:      tc.denom0,
				Denom1:      tc.denom1,
				TickSpacing: tc.tickSpacing,
				SwapFee:     DefaultZeroSwapFee,
			})

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, poolmanagertypes.TypeEvtPoolCreated, tc.expectedPoolCreatedEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
			}
		})
	}
}

// TestCreatePositionMsg tests that create position msg validate basic have been correctly implemented.
// Also checks correct assertion of events of CreatePosition.
func (suite *KeeperTestSuite) TestCreatePositionMsg() {
	testcases := map[string]lpTest{
		"happy case": {},
		"error: lower tick is equal to upper tick": {
			lowerTick:     DefaultUpperTick,
			expectedError: types.InvalidLowerUpperTickError{LowerTick: DefaultUpperTick, UpperTick: DefaultUpperTick},
		},
		"error: tokens provided is three": {
			tokensProvided: DefaultCoins.Add(sdk.NewCoin("foo", sdk.NewInt(10))),
			expectedError:  types.CoinLengthError{Length: 3, MaxLength: 2},
		},
		"error: token min amount 0 is negative": {
			amount0Minimum: sdk.NewInt(-10),
			expectedError:  types.NotPositiveRequireAmountError{Amount: sdk.NewInt(-10).String()},
		},
		"error: token min amount 1 is negative": {
			amount1Minimum: sdk.NewInt(-10),
			expectedError:  types.NotPositiveRequireAmountError{Amount: sdk.NewInt(-10).String()},
		},
	}
	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx

			baseConfigCopy := *baseCase
			fmt.Println(baseConfigCopy.tokensProvided)
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			suite.PrepareConcentratedPool()
			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// fund sender to create position
			suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))

			msg := &types.MsgCreatePosition{
				PoolId:          tc.poolId,
				Sender:          suite.TestAccs[0].String(),
				LowerTick:       tc.lowerTick,
				UpperTick:       tc.upperTick,
				TokensProvided:  tc.tokensProvided,
				TokenMinAmount0: tc.amount0Minimum,
				TokenMinAmount1: tc.amount1Minimum,
			}

			if tc.expectedError == nil {
				response, err := msgServer.CreatePosition(sdk.WrapSDKContext(ctx), msg)
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, 2)
			} else {
				suite.Require().ErrorContains(msg.ValidateBasic(), tc.expectedError.Error())
			}
		})
	}
}

// TestAddToPosition_Events tests that events are correctly emitted
// when calling AddToPosition.
func (suite *KeeperTestSuite) TestAddToPosition_Events() {
	testcases := map[string]struct {
		lastPositionInPool           bool
		expectedAddedToPositionEvent int
		expectedMessageEvents        int
		expectedError                error
	}{
		"happy path": {
			expectedAddedToPositionEvent: 1,
			expectedMessageEvents:        5,
		},
		"error: last position in pool": {
			lastPositionInPool:           true,
			expectedAddedToPositionEvent: 0,
			expectedError:                types.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()

			// Position from current account.
			posId := suite.SetupDefaultPositionAcc(pool.GetId(), suite.TestAccs[0])

			if !tc.lastPositionInPool {
				// Position from another account.
				suite.SetupDefaultPositionAcc(pool.GetId(), suite.TestAccs[1])
			}

			// Reset event counts to 0 by creating a new manager.
			suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(suite.Ctx.EventManager().Events()))

			suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
			msg := &types.MsgAddToPosition{
				PositionId:    posId,
				Sender:        suite.TestAccs[0].String(),
				TokenDesired0: DefaultCoin0,
				TokenDesired1: DefaultCoin1,
			}

			response, err := msgServer.AddToPosition(sdk.WrapSDKContext(suite.Ctx), msg)

			if tc.expectedError == nil {
				suite.NoError(err)
				suite.NotNil(response)
				suite.AssertEventEmitted(suite.Ctx, types.TypeEvtAddToPosition, tc.expectedAddedToPositionEvent)
				suite.AssertEventEmitted(suite.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())
				suite.Require().Nil(response)
				suite.AssertEventEmitted(suite.Ctx, types.TypeEvtAddToPosition, tc.expectedAddedToPositionEvent)
			}
		})
	}
}

// TODO: Add test cases for withdraw position messages

// TestCollectFees_Events tests that events are correctly emitted
// when calling CollectFees.
func (suite *KeeperTestSuite) TestCollectFees_Events() {
	testcases := map[string]struct {
		upperTick                     int64
		lowerTick                     int64
		positionIds                   []uint64
		numPositionsToCreate          int
		shouldSetupUnownedPosition    bool
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
		"error: attempt to claim fees with different owner": {
			upperTick:                     DefaultUpperTick,
			lowerTick:                     DefaultLowerTick,
			positionIds:                   []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			shouldSetupUnownedPosition:    true,
			numPositionsToCreate:          2,
			expectedTotalCollectFeesEvent: 0,
			expectedError:                 types.NotPositionOwnerError{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				suite.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				suite.SetupDefaultPositionAcc(pool.GetId(), suite.TestAccs[1])
			}

			// Reset event counts to 0 by creating a new manager.
			suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(suite.Ctx.EventManager().Events()))

			msg := &types.MsgCollectFees{
				Sender:      suite.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			response, err := msgServer.CollectFees(sdk.WrapSDKContext(suite.Ctx), msg)

			if tc.expectedError == nil {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
				suite.AssertEventEmitted(suite.Ctx, types.TypeEvtTotalCollectFees, tc.expectedTotalCollectFeesEvent)
				suite.AssertEventEmitted(suite.Ctx, types.TypeEvtCollectFees, tc.expectedCollectFeesEvent)
				suite.AssertEventEmitted(suite.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorAs(err, &tc.expectedError)
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
		shouldSetupUnownedPosition          bool
		expectedTotalCollectIncentivesEvent int
		expectedCollectIncentivesEvent      int
		expectedMessageEvents               int
		expectedError                       error
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
		"error: three position IDs - not an owner": {
			upperTick:                  DefaultUpperTick,
			lowerTick:                  DefaultLowerTick,
			positionIds:                []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:       2,
			shouldSetupUnownedPosition: true,
			expectedError:              types.NotPositionOwnerError{},
		},
		"error": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:                2,
			expectedTotalCollectIncentivesEvent: 0,
			expectedCollectIncentivesEvent:      0,
			expectedError:                       types.PositionIdNotFoundError{PositionId: DefaultPositionId + 2},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()
			ctx := suite.Ctx

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				suite.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				suite.SetupDefaultPositionAcc(pool.GetId(), suite.TestAccs[1])
			}

			position, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(ctx, tc.positionIds[0])
			suite.Require().NoError(err)
			ctx = ctx.WithBlockTime(position.JoinTime.Add(time.Hour * 24 * 7))
			positionAge := ctx.BlockTime().Sub(position.JoinTime)

			// Set up accrued incentives
			err = addToUptimeAccums(ctx, pool.GetId(), suite.App.ConcentratedLiquidityKeeper, uptimeHelper.hundredTokensMultiDenom)
			suite.Require().NoError(err)
			suite.FundAcc(pool.GetIncentivesAddress(), expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, positionAge, sdk.NewInt(int64(len(tc.positionIds)))))

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			msg := &types.MsgCollectIncentives{
				Sender:      suite.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			// System under test
			response, err := msgServer.CollectIncentives(sdk.WrapSDKContext(ctx), msg)

			if tc.expectedError == nil {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
				suite.AssertEventEmitted(ctx, types.TypeEvtTotalCollectIncentives, tc.expectedTotalCollectIncentivesEvent)
				suite.AssertEventEmitted(ctx, types.TypeEvtCollectIncentives, tc.expectedCollectIncentivesEvent)
				suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorAs(err, &tc.expectedError)
				suite.Require().Nil(response)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestFungify_Events() {
	testcases := map[string]struct {
		positionIdsToFungify       []uint64
		numPositionsToCreate       int
		shouldSetupUnownedPosition bool
		shouldSetupUncharged       bool
		expectedFungifyEvents      int
		expectedMessageEvents      int
		expectedError              error
	}{
		"three position IDs": {
			positionIdsToFungify:  []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:  3,
			expectedFungifyEvents: 1,
			expectedMessageEvents: 1, // 1 for fungify
		},
		"error: single position ID": {
			positionIdsToFungify: []uint64{DefaultPositionId},
			numPositionsToCreate: 1,

			expectedError: types.PositionQuantityTooLowError{},
		},
		"error: attempt to fungify with different owner": {
			positionIdsToFungify:       []uint64{DefaultPositionId, DefaultPositionId + 1},
			shouldSetupUnownedPosition: true,
			numPositionsToCreate:       1,
			expectedError:              types.NotPositionOwnerError{},
		},
		"error: not fully charged": {
			positionIdsToFungify: []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate: 2,
			shouldSetupUncharged: true,
			expectedError:        types.PositionNotFullyChargedError{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()

			msgServer := cl.NewMsgServerImpl(suite.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := suite.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				suite.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				suite.SetupDefaultPositionAcc(pool.GetId(), suite.TestAccs[1])
			}

			fullChargeDuration := suite.App.ConcentratedLiquidityKeeper.GetLargestAuthorizedUptimeDuration(suite.Ctx)
			suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(fullChargeDuration))

			if tc.shouldSetupUncharged {
				suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(-time.Millisecond))
			}

			// Reset event counts to 0 by creating a new manager.
			suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(suite.Ctx.EventManager().Events()))

			msg := &types.MsgFungifyChargedPositions{
				Sender:      suite.TestAccs[0].String(),
				PositionIds: tc.positionIdsToFungify,
			}

			response, err := msgServer.FungifyChargedPositions(sdk.WrapSDKContext(suite.Ctx), msg)

			if tc.expectedError == nil {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
				suite.AssertEventEmitted(suite.Ctx, types.TypeEvtFungifyChargedPosition, tc.expectedFungifyEvents)
				suite.AssertEventEmitted(suite.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorAs(err, &tc.expectedError)
				suite.Require().Nil(response)
			}
		})
	}
}
