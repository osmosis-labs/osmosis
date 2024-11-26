package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cl "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestBeforeValidatorSlashed() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		slashedValIndexes     []int64
		expSlashedLockIndexes []int64
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]int64{0},
		},
		{
			"with an unbonding validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Unbonding},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]int64{0},
		},
		{
			"with an unbonded validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Unbonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{},
			[]int64{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			locks := []lockuptypes.PeriodLock{}
			slashFactor := osmomath.NewDecWithPrec(5, 2)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				delAddr := delAddrs[del.delIndex]
				lock := s.setupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)

				// save accounts and locks for future use
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.slashedValIndexes {
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrs[valIndex])
				s.Require().NoError(err)
				s.Ctx = s.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				s.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)

				// should not be slashing unbonded validator
				defer func() {
					if r := recover(); r != nil {
						s.Require().Equal(true, validator.IsUnbonded())
					}
				}()
				s.App.StakingKeeper.Slash(s.Ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeValidatorSlashed hook
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expSlashedLockIndexes {
				gotLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, locks[lockIndex].ID)
				s.Require().NoError(err)
				s.Require().Equal(
					osmomath.NewDec(1000000).Mul(osmomath.OneDec().Sub(slashFactor)).TruncateInt().String(),
					gotLock.Coins.AmountOf(denoms[0]).String(),
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSlashLockupsForUnbondingDelegationSlash() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
	}{
		{
			"happy path with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]uint64{1, 2},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)
				// superfluid undelegate
				err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.Owner, lockId)
				s.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			slashFactor := osmomath.NewDecWithPrec(5, 2)
			for i := 0; i < len(valAddrs); i++ {
				s.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
					s.Ctx,
					valAddrs[i],
					slashFactor)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check check unbonding lockup changes
			for _, lockId := range tc.superUnbondingLockIds {
				gotLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)
				s.Require().Equal(gotLock.Coins[0].Amount.String(), osmomath.NewInt(950000).String())
			}
		})
	}
}

func (s *KeeperTestSuite) TestPrepareConcentratedLockForSlash() {
	type prepareConcentratedLockTestCase struct {
		name         string
		slashPercent osmomath.Dec
		expectedErr  bool
	}

	testCases := []prepareConcentratedLockTestCase{
		{
			name:         "SmallSlash",
			slashPercent: osmomath.MustNewDecFromStr("0.001"),
			expectedErr:  false,
		},
		{
			name:         "FullSlash",
			slashPercent: osmomath.MustNewDecFromStr("1"),
			expectedErr:  false,
		},
		{
			name:         "HalfSlash",
			slashPercent: osmomath.MustNewDecFromStr("0.5"),
			expectedErr:  false,
		},
		{
			name:         "OverSlash",
			slashPercent: osmomath.MustNewDecFromStr("1.001"),
			expectedErr:  true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			clPool, concentratedLockId, positionId := s.PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(appparams.BaseCoinUnit, apptesting.USDC)
			clPoolId := clPool.GetId()

			lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, concentratedLockId)
			s.Require().NoError(err)

			// Get position state entry
			positionPreSlash, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
			s.Require().NoError(err)

			// Get tick info for lower and upper tick
			lowerTickInfoPreSlash, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, clPoolId, positionPreSlash.LowerTick)
			s.Require().NoError(err)
			upperTickInfoPreSlash, err := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, clPoolId, positionPreSlash.UpperTick)
			s.Require().NoError(err)
			liquidityPreSlash := clPool.GetLiquidity()

			// Calculate underlying assets from liquidity getting slashed
			asset0PreSlash, asset1PreSlash, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, positionPreSlash, clPool)
			s.Require().NoError(err)

			slashAmt := positionPreSlash.Liquidity.Mul(tc.slashPercent)

			// Note the value of the fee accumulator before the slash for the position.
			feeAccumPreSlash, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, clPoolId)
			s.Require().NoError(err)
			feePositionKey := cltypes.KeySpreadRewardPositionAccumulator(positionId)
			positionSizePreSlash, err := feeAccumPreSlash.GetPositionSize(feePositionKey)
			s.Require().NoError(err)

			// Note the numShares value of the position at each of the uptime accumulators.
			uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPoolId)
			uptimePositionKey := string(cltypes.KeyPositionId(positionId))
			s.Require().NoError(err)
			numShares := make([]osmomath.Dec, len(uptimeAccums))
			for i, uptimeAccum := range uptimeAccums {
				position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
				s.Require().NoError(err)
				numShares[i] = position.NumShares
			}

			// System under test
			clPoolAddress, underlyingAssetsToSlash, err := s.App.SuperfluidKeeper.PrepareConcentratedLockForSlash(s.Ctx, lock, slashAmt)

			lowerTickInfoPostSlash, getTickErr := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, clPoolId, positionPreSlash.LowerTick)
			s.Require().NoError(getTickErr)
			upperTickInfoPostSlash, getTickErr := s.App.ConcentratedLiquidityKeeper.GetTickInfo(s.Ctx, clPoolId, positionPreSlash.UpperTick)
			s.Require().NoError(getTickErr)

			clPool, getPoolErr := s.App.ConcentratedLiquidityKeeper.GetConcentratedPoolById(s.Ctx, clPoolId)
			s.Require().NoError(getPoolErr)
			liquidityPostSlash := clPool.GetLiquidity()
			if tc.expectedErr {
				s.Require().Error(err)

				// Check that the position's fee accumulator has not changed.
				feeAccumPostSlash, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, clPoolId)
				s.Require().NoError(err)
				positionSizePostSlash, err := feeAccumPostSlash.GetPositionSize(feePositionKey)
				s.Require().NoError(err)
				s.Require().Equal(positionSizePreSlash.String(), positionSizePostSlash.String())

				// Check that the numShares value of the position has not been updated in each of the uptime accumulators.
				uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPoolId)
				s.Require().NoError(err)
				for i, uptimeAccum := range uptimeAccums {
					position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
					s.Require().NoError(err)
					s.Require().Equal(numShares[i].String(), position.NumShares.String())
				}
			} else {
				s.Require().NoError(err)
				s.Require().Equal(clPool.GetAddress(), clPoolAddress)

				s.Require().Equal(lowerTickInfoPreSlash.LiquidityGross.Sub(slashAmt).String(), lowerTickInfoPostSlash.LiquidityGross.String())
				s.Require().Equal(upperTickInfoPreSlash.LiquidityGross.Sub(slashAmt).String(), upperTickInfoPostSlash.LiquidityGross.String())

				s.Require().Equal(lowerTickInfoPreSlash.LiquidityNet.Sub(slashAmt).String(), lowerTickInfoPostSlash.LiquidityNet.String())
				s.Require().Equal(upperTickInfoPreSlash.LiquidityNet.Add(slashAmt).String(), upperTickInfoPostSlash.LiquidityNet.String())

				s.Require().Equal(liquidityPreSlash.Sub(slashAmt).String(), liquidityPostSlash.String())

				positionPostSlash, err := s.App.ConcentratedLiquidityKeeper.GetPosition(s.Ctx, positionId)
				s.Require().NoError(err)
				s.Require().Equal(positionPreSlash.Liquidity.Sub(slashAmt).String(), positionPostSlash.Liquidity.String())

				asset0PostSlash, asset1PostSlash, err := cl.CalculateUnderlyingAssetsFromPosition(s.Ctx, positionPostSlash, clPool)
				s.Require().NoError(err)

				errTolerance := osmomath.ErrTolerance{
					AdditiveTolerance: osmomath.NewDec(1),
					// Actual should be greater than expected, so we round up
					RoundingDir: osmomath.RoundUp,
				}

				osmoassert.Equal(s.T(), errTolerance, asset0PreSlash.Sub(asset0PostSlash).Amount, underlyingAssetsToSlash[0].Amount)
				osmoassert.Equal(s.T(), errTolerance, asset1PreSlash.Sub(asset1PostSlash).Amount, underlyingAssetsToSlash[1].Amount)

				// Check that the fee accumulator has been updated by the amount it was slashed.
				feeAccumPostSlash, err := s.App.ConcentratedLiquidityKeeper.GetSpreadRewardAccumulator(s.Ctx, clPoolId)
				s.Require().NoError(err)
				positionSizePostSlash, err := feeAccumPostSlash.GetPositionSize(feePositionKey)
				s.Require().NoError(err)
				s.Require().Equal(positionSizePreSlash.Sub(slashAmt).String(), positionSizePostSlash.String())

				// Check that the numShares value of the position has been updated in each of the uptime accumulators.
				uptimeAccums, err := s.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(s.Ctx, clPoolId)
				s.Require().NoError(err)
				for i, uptimeAccum := range uptimeAccums {
					position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
					s.Require().NoError(err)
					s.Require().Equal(numShares[i].Sub(slashAmt).String(), position.NumShares.String())
				}
			}
		})
	}
}
