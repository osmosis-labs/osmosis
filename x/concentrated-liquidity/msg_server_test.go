package concentrated_liquidity_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// TestCreateConcentratedPool_Events tests that events are correctly emitted
// when calling CreateConcentratedPool.
func (s *KeeperTestSuite) TestCreateConcentratedPool_Events() {
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
			expectedMessageEvents:    3, // 1 for coin spent, 1 for coin received, 1 for after pool create hook
		},
		"error: tickSpacing zero": {
			denom0:        ETH,
			denom1:        USDC,
			tickSpacing:   0,
			expectedError: fmt.Errorf("tick spacing must be positive"),
		},
		"error: tickSpacing not authorized": {
			denom0:        ETH,
			denom1:        USDC,
			tickSpacing:   DefaultTickSpacing + 1,
			expectedError: types.UnauthorizedTickSpacingError{ProvidedTickSpacing: DefaultTickSpacing + 1, AuthorizedTickSpacings: s.App.ConcentratedLiquidityKeeper.GetParams(s.Ctx).AuthorizedTickSpacing},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			ctx := s.Ctx

			// Retrieve the pool creation fee from poolmanager params.
			poolmanagerParams := poolmanagertypes.DefaultParams()

			// Fund account to pay for the pool creation fee.
			s.FundAcc(s.TestAccs[0], poolmanagerParams.PoolCreationFee)

			msgServer := cl.NewMsgCreatorServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.CreateConcentratedPool(ctx, &clmodel.MsgCreateConcentratedPool{
				Sender:       s.TestAccs[0].String(),
				Denom0:       tc.denom0,
				Denom1:       tc.denom1,
				TickSpacing:  tc.tickSpacing,
				SpreadFactor: DefaultZeroSpreadFactor,
			})

			if tc.expectedError == nil {
				s.NoError(err)
				s.NotNil(response)
				s.AssertEventEmitted(ctx, poolmanagertypes.TypeEvtPoolCreated, tc.expectedPoolCreatedEvent)
				s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Nil(response)
			}
		})
	}
}

// TestCreatePositionMsg tests that create position msg validate basic have been correctly implemented.
// Also checks correct assertion of events of CreatePosition.
func (s *KeeperTestSuite) TestCreatePositionMsg() {
	testcases := map[string]lpTest{
		"happy case": {},
		"error: lower tick is equal to upper tick": {
			lowerTick:     DefaultUpperTick,
			expectedError: types.InvalidLowerUpperTickError{LowerTick: DefaultUpperTick, UpperTick: DefaultUpperTick},
		},
		"error: tokens provided is three": {
			tokensProvided: DefaultCoins.Add(sdk.NewCoin("foo", osmomath.NewInt(10))),
			expectedError:  types.CoinLengthError{Length: 3, MaxLength: 2},
		},
		"error: token min amount 0 is negative": {
			amount0Minimum: osmomath.NewInt(-10),
			expectedError:  types.NotPositiveRequireAmountError{Amount: osmomath.NewInt(-10).String()},
		},
		"error: token min amount 1 is negative": {
			amount1Minimum: osmomath.NewInt(-10),
			expectedError:  types.NotPositiveRequireAmountError{Amount: osmomath.NewInt(-10).String()},
		},
	}
	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()
			ctx := s.Ctx

			baseConfigCopy := *baseCase
			mergeConfigs(&baseConfigCopy, &tc)
			tc = baseConfigCopy

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			s.PrepareConcentratedPool()
			msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// fund Sender to create position
			s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))

			msg := &types.MsgCreatePosition{
				PoolId:          tc.poolId,
				Sender:          s.TestAccs[0].String(),
				LowerTick:       tc.lowerTick,
				UpperTick:       tc.upperTick,
				TokensProvided:  tc.tokensProvided,
				TokenMinAmount0: tc.amount0Minimum,
				TokenMinAmount1: tc.amount1Minimum,
			}

			if tc.expectedError == nil {
				response, err := msgServer.CreatePosition(ctx, msg)
				s.NoError(err)
				s.NotNil(response)
				s.AssertEventEmitted(ctx, sdk.EventTypeMessage, 1)
			} else {
				s.Require().ErrorContains(msg.ValidateBasic(), tc.expectedError.Error())
			}
		})
	}
}

// TestAddToPosition_Events tests that events are correctly emitted
// when calling AddToPosition.
func (s *KeeperTestSuite) TestAddToPosition_Events() {
	testcases := map[string]struct {
		lastPositionInPool           bool
		expectedAddedToPositionEvent int
		expectedMessageEvents        int
		expectedError                error
	}{
		"happy path": {
			expectedAddedToPositionEvent: 1,
			expectedMessageEvents:        3,
		},
		"error: last position in pool": {
			lastPositionInPool:           true,
			expectedAddedToPositionEvent: 0,
			expectedError:                types.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := s.PrepareConcentratedPool()

			// Position from current account.
			posId := s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[0])

			if !tc.lastPositionInPool {
				// Position from another account.
				s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[1])
			}

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			s.FundAcc(s.TestAccs[0], sdk.NewCoins(DefaultCoin0, DefaultCoin1))
			msg := &types.MsgAddToPosition{
				PositionId: posId,
				Sender:     s.TestAccs[0].String(),
				Amount0:    DefaultCoin0.Amount,
				Amount1:    DefaultCoin1.Amount,
			}

			response, err := msgServer.AddToPosition(s.Ctx, msg)

			if tc.expectedError == nil {
				s.NoError(err)
				s.NotNil(response)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtAddToPosition, tc.expectedAddedToPositionEvent)
				s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Nil(response)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtAddToPosition, tc.expectedAddedToPositionEvent)
			}
		})
	}
}

// TODO: Add test cases for withdraw position messages

// TestCollectSpreadRewards_Events tests that events are correctly emitted
// when calling CollectSpreadRewards.
func (s *KeeperTestSuite) TestCollectSpreadRewards_Events() {
	testcases := map[string]struct {
		upperTick                              int64
		lowerTick                              int64
		positionIds                            []uint64
		numPositionsToCreate                   int
		shouldSetupUnownedPosition             bool
		expectedTotalCollectSpreadRewardsEvent int
		expectedCollectSpreadRewardsEvent      int
		expectedMessageEvents                  int
		expectedError                          error
		errorFromValidateBasic                 error
	}{
		"single position ID": {
			upperTick:                              DefaultUpperTick,
			lowerTick:                              DefaultLowerTick,
			positionIds:                            []uint64{DefaultPositionId},
			numPositionsToCreate:                   1,
			expectedTotalCollectSpreadRewardsEvent: 1,
			expectedCollectSpreadRewardsEvent:      1,
			expectedMessageEvents:                  1, // 1 for send message
		},
		"two position IDs": {
			upperTick:                              DefaultUpperTick,
			lowerTick:                              DefaultLowerTick,
			positionIds:                            []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate:                   2,
			expectedTotalCollectSpreadRewardsEvent: 1,
			expectedCollectSpreadRewardsEvent:      2,
			expectedMessageEvents:                  2, // 2 for send messages
		},
		"three position IDs": {
			upperTick:                              DefaultUpperTick,
			lowerTick:                              DefaultLowerTick,
			positionIds:                            []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:                   3,
			expectedTotalCollectSpreadRewardsEvent: 1,
			expectedCollectSpreadRewardsEvent:      3,
			expectedMessageEvents:                  3, // 3 for send messages
		},
		"error: attempt to claim fees with different owner": {
			upperTick:                              DefaultUpperTick,
			lowerTick:                              DefaultLowerTick,
			positionIds:                            []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			shouldSetupUnownedPosition:             true,
			numPositionsToCreate:                   2,
			expectedTotalCollectSpreadRewardsEvent: 0,
			expectedError:                          types.NotPositionOwnerError{},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := s.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				s.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[1])
			}

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			msg := &types.MsgCollectSpreadRewards{
				Sender:      s.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			// Add spread rewards to the pool's accum so we aren't just claiming 0 rewards.
			// Claiming 0 rewards is still a valid message, but is not as valuable for testing.
			s.AddToSpreadRewardAccumulator(validPoolId, sdk.NewDecCoin(ETH, osmomath.NewInt(1)))

			// Determine expected rewards from all provided positions without modifying state.
			expectedTotalSpreadRewards := sdk.Coins(nil)
			cacheCtx, _ := s.Ctx.CacheContext()
			for _, positionId := range tc.positionIds {
				spreadRewardsClaimed, _ := s.App.ConcentratedLiquidityKeeper.PrepareClaimableSpreadRewards(cacheCtx, positionId)
				expectedTotalSpreadRewards = expectedTotalSpreadRewards.Add(spreadRewardsClaimed...)
			}

			// Fund the spread rewards account with the expected rewards (not testing the distribution algorithm here, just the events, so this is okay)
			s.FundAcc(pool.GetSpreadRewardsAddress(), sdk.NewCoins(sdk.NewCoin(ETH, expectedTotalSpreadRewards[0].Amount)))

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			// System under test.
			response, err := msgServer.CollectSpreadRewards(s.Ctx, msg)

			if tc.expectedError == nil {
				s.Require().NoError(err)
				s.Require().NotNil(response)
				s.Require().Equal(expectedTotalSpreadRewards, response.CollectedSpreadRewards)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtTotalCollectSpreadRewards, tc.expectedTotalCollectSpreadRewardsEvent)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtCollectSpreadRewards, tc.expectedCollectSpreadRewardsEvent)
				s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				s.Require().Nil(response)
			}
		})
	}
}

// TestCollectIncentives_Events tests that events are correctly emitted
// when calling CollectIncentives.
func (s *KeeperTestSuite) TestCollectIncentives_Events() {
	uptimeHelper := getExpectedUptimes()
	twoWeeks := time.Hour * 24 * 14
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
			expectedMessageEvents:               1, // 1 for collect send
		},
		"two position IDs": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate:                2,
			expectedTotalCollectIncentivesEvent: 1,
			expectedCollectIncentivesEvent:      2,
			expectedMessageEvents:               2, // 2 for collect send
		},
		"three position IDs": {
			upperTick:                           DefaultUpperTick,
			lowerTick:                           DefaultLowerTick,
			positionIds:                         []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:                3,
			expectedTotalCollectIncentivesEvent: 1,
			expectedCollectIncentivesEvent:      3,
			expectedMessageEvents:               3, // 3 for collect send
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
		s.Run(name, func() {
			s.SetupTest()
			ctx := s.Ctx

			// Create a cl pool with a default position
			pool := s.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				s.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[1])
			}

			position, err := s.App.ConcentratedLiquidityKeeper.GetPosition(ctx, tc.positionIds[0])
			s.Require().NoError(err)
			ctx = ctx.WithBlockTime(position.JoinTime.Add(time.Hour * 24 * 7))
			positionAge := ctx.BlockTime().Sub(position.JoinTime)

			// Set up accrued incentives
			err = addToUptimeAccums(ctx, pool.GetId(), s.App.ConcentratedLiquidityKeeper, uptimeHelper.hundredTokensMultiDenom)
			s.Require().NoError(err)

			numPositions := osmomath.NewInt(int64(len(tc.positionIds)))
			// Fund the incentives address with the amount of incentives we expect the positions to both claim and forfeit.
			// The claim amount must be funded to the incentives address in order for the rewards to be sent to the user.
			// The forfeited about must be funded to the incentives address in order for the forfeited rewards to be sent to the community pool.
			incentivesToBeSentToUsers := expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, positionAge, numPositions)
			incentivesToBeSentToCommunityPool := expectedIncentivesFromUptimeGrowth(uptimeHelper.hundredTokensMultiDenom, DefaultLiquidityAmt, twoWeeks, numPositions).Sub(incentivesToBeSentToUsers...)
			totalAmountToFund := incentivesToBeSentToUsers.Add(incentivesToBeSentToCommunityPool...)
			s.FundAcc(pool.GetIncentivesAddress(), totalAmountToFund)

			msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			msg := &types.MsgCollectIncentives{
				Sender:      s.TestAccs[0].String(),
				PositionIds: tc.positionIds,
			}

			// System under test
			response, err := msgServer.CollectIncentives(ctx, msg)

			if tc.expectedError == nil {
				s.Require().NoError(err)
				s.Require().NotNil(response)
				s.AssertEventEmitted(ctx, types.TypeEvtTotalCollectIncentives, tc.expectedTotalCollectIncentivesEvent)
				s.AssertEventEmitted(ctx, types.TypeEvtCollectIncentives, tc.expectedCollectIncentivesEvent)
				s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				s.Require().Nil(response)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFungify_Events() {

	s.T().Skip("TODO: re-enable fungify test if message is restored")

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
		s.Run(name, func() {
			s.SetupTest()

			// msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Create a cl pool with a default position
			pool := s.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				s.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[1])
			}

			fullChargeDuration := s.App.ConcentratedLiquidityKeeper.GetLargestAuthorizedUptimeDuration(s.Ctx)
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(fullChargeDuration))

			if tc.shouldSetupUncharged {
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(-time.Millisecond))
			}

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			// msg := &types.MsgFungifyChargedPositions{
			// 	Sender:      s.TestAccs[0].String(),
			// 	PositionIds: tc.positionIdsToFungify,
			// }

			// response, err := msgServer.FungifyChargedPositions(s.Ctx, msg)

			// if tc.expectedError == nil {
			// 	s.Require().NoError(err)
			// 	s.Require().NotNil(response)
			// 	s.AssertEventEmitted(s.Ctx, types.TypeEvtFungifyChargedPosition, tc.expectedFungifyEvents)
			// 	s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			// } else {
			// 	s.Require().Error(err)
			// 	s.Require().ErrorAs(err, &tc.expectedError)
			// 	s.Require().Nil(response)
			// }
		})
	}
}

func (s *KeeperTestSuite) TestTransferPositions_Events() {
	// expectedUptimes are used for claimable incentives tests
	expectedUptimes := getExpectedUptimes()

	testcases := map[string]struct {
		positionIds                    []uint64
		numPositionsToCreate           int
		shouldSetupUnownedPosition     bool
		hasIncentivesToClaim           bool
		hasSpreadRewardsToClaim        bool
		expectedTransferPositionsEvent int
		expectedMessageEvents          int // We expect these to always be 0 because no additional events are being triggered
		isLastPositionInPool           bool
		expectedError                  error
	}{
		"single position ID": {
			positionIds:                    []uint64{DefaultPositionId},
			numPositionsToCreate:           1,
			expectedTransferPositionsEvent: 1,
		},
		"single position ID with claimable incentives": {
			positionIds:                    []uint64{DefaultPositionId},
			hasIncentivesToClaim:           true,
			numPositionsToCreate:           1,
			expectedTransferPositionsEvent: 1,
			expectedMessageEvents:          0,
		},
		"single position ID with claimable spread rewards": {
			positionIds:                    []uint64{DefaultPositionId},
			hasSpreadRewardsToClaim:        true,
			numPositionsToCreate:           1,
			expectedTransferPositionsEvent: 1,
			expectedMessageEvents:          0,
		},
		"single position ID with claimable incentives and spread rewards": {
			positionIds:                    []uint64{DefaultPositionId},
			hasIncentivesToClaim:           true,
			hasSpreadRewardsToClaim:        true,
			numPositionsToCreate:           1,
			expectedTransferPositionsEvent: 1,
			expectedMessageEvents:          0,
		},
		"two position IDs": {
			positionIds:                    []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate:           2,
			expectedTransferPositionsEvent: 1,
		},
		"three position IDs": {
			positionIds:                    []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:           3,
			expectedTransferPositionsEvent: 1,
		},
		"three position IDs with claimable incentives and spread rewards": {
			positionIds:                    []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			hasIncentivesToClaim:           true,
			hasSpreadRewardsToClaim:        true,
			numPositionsToCreate:           3,
			expectedTransferPositionsEvent: 1,
			expectedMessageEvents:          0,
		},
		"two position IDs, second ID does not exist": {
			positionIds:          []uint64{DefaultPositionId, DefaultPositionId + 1},
			numPositionsToCreate: 1,
			expectedError:        types.PositionIdNotFoundError{PositionId: DefaultPositionId + 1},
		},
		"three position IDs, not an owner of one of them": {
			positionIds:                []uint64{DefaultPositionId, DefaultPositionId + 1, DefaultPositionId + 2},
			numPositionsToCreate:       2,
			shouldSetupUnownedPosition: true,
			expectedError:              types.NotPositionOwnerError{},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			// Create a cl pool with a default position
			pool := s.PrepareConcentratedPool()
			for i := 0; i < tc.numPositionsToCreate; i++ {
				s.SetupDefaultPosition(pool.GetId())
			}

			if tc.shouldSetupUnownedPosition {
				// Position from another account.
				s.SetupDefaultPositionAcc(pool.GetId(), s.TestAccs[1])
			}

			if tc.hasIncentivesToClaim {
				s.fundIncentiveAddr(pool.GetIncentivesAddress(), tc.positionIds)
				s.addUptimeGrowthInsideRange(s.Ctx, pool.GetId(), apptesting.DefaultLowerTick+1, DefaultLowerTick, DefaultUpperTick, expectedUptimes.hundredTokensMultiDenom)
			}

			if tc.hasSpreadRewardsToClaim {
				s.fundSpreadRewardsAddr(s.Ctx, pool.GetSpreadRewardsAddress(), tc.positionIds)
				s.AddToSpreadRewardAccumulator(pool.GetId(), sdk.NewDecCoin(ETH, osmomath.NewInt(10)))
			}

			// Move block time forward one day to claim and forfeit part of the incentives.
			s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Hour * 24))

			if !tc.isLastPositionInPool {
				// Setup a far out of range position that we do not touch, so when we transfer positions we do not transfer the last position in the pool.
				// This is because we special case this logic in the keeper to not allow the last position in the pool to be transferred.
				s.SetupPosition(pool.GetId(), s.TestAccs[2], sdk.NewCoins(DefaultCoin1), DefaultMinTick, DefaultMinTick+100, true)
			}

			msgServer := cl.NewMsgServerImpl(s.App.ConcentratedLiquidityKeeper)

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			msg := &types.MsgTransferPositions{
				PositionIds: tc.positionIds,
				Sender:      s.TestAccs[0].String(),
				NewOwner:    s.TestAccs[1].String(),
			}

			// System under test
			response, err := msgServer.TransferPositions(s.Ctx, msg)

			if tc.expectedError == nil {
				s.Require().NoError(err)
				s.Require().NotNil(response)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtTransferPositions, tc.expectedTransferPositionsEvent)
				s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			} else {
				s.Require().Error(err)
				s.Require().ErrorAs(err, &tc.expectedError)
				s.Require().Nil(response)
			}
		})
	}
}
