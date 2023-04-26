package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestBeforeValidatorSlashed() {
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
			[]int64{0},
			[]int64{0},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			locks := []lockuptypes.PeriodLock{}
			slashFactor := sdk.NewDecWithPrec(5, 2)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				delAddr := delAddrs[del.delIndex]
				lock := suite.setupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)

				// save accounts and locks for future use
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.Ctx = suite.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)

				// should not be slashing unbonded validator
				defer func() {
					if r := recover(); r != nil {
						suite.Require().Equal(true, validator.IsUnbonded())
					}
				}()
				suite.App.StakingKeeper.Slash(suite.Ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeValidatorSlashed hook
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expSlashedLockIndexes {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, locks[lockIndex].ID)
				suite.Require().NoError(err)
				suite.Require().Equal(
					gotLock.Coins.AmountOf(denoms[0]).String(),
					sdk.NewDec(1000000).Mul(sdk.OneDec().Sub(slashFactor)).TruncateInt().String(),
				)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSlashLockupsForUnbondingDelegationSlash() {
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

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				// superfluid undelegate
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			slashFactor := sdk.NewDecWithPrec(5, 2)
			for i := 0; i < len(valAddrs); i++ {
				suite.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
					suite.Ctx,
					valAddrs[i],
					suite.Ctx.BlockHeight(),
					slashFactor)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check check unbonding lockup changes
			for _, lockId := range tc.superUnbondingLockIds {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(gotLock.Coins[0].Amount.String(), sdk.NewInt(950000).String())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestPrepareConcentratedLockForSlash() {
	type prepareConcentratedLockTestCase struct {
		name         string
		slashPercent sdk.Dec
		expectedErr  bool
	}

	testCases := []prepareConcentratedLockTestCase{
		{
			name:         "SmallSlash",
			slashPercent: sdk.MustNewDecFromStr("0.001"),
			expectedErr:  false,
		},
		{
			name:         "FullSlash",
			slashPercent: sdk.MustNewDecFromStr("1"),
			expectedErr:  false,
		},
		{
			name:         "HalfSlash",
			slashPercent: sdk.MustNewDecFromStr("0.5"),
			expectedErr:  false,
		},
		{
			name:         "OverSlash",
			slashPercent: sdk.MustNewDecFromStr("1.001"),
			expectedErr:  true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			clPool, concentratedLockId, positionId := suite.PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition("uosmo", apptesting.USDC)
			clPoolId := clPool.GetId()

			lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, concentratedLockId)
			suite.Require().NoError(err)

			// Get position state entry
			positionPreSlash, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(suite.Ctx, positionId)
			suite.Require().NoError(err)

			// Get tick info for lower and upper tick
			lowerTickInfoPreSlash, err := suite.App.ConcentratedLiquidityKeeper.GetTickInfo(suite.Ctx, clPoolId, positionPreSlash.LowerTick)
			suite.Require().NoError(err)
			upperTickInfoPreSlash, err := suite.App.ConcentratedLiquidityKeeper.GetTickInfo(suite.Ctx, clPoolId, positionPreSlash.UpperTick)
			suite.Require().NoError(err)
			liquidityPreSlash := clPool.GetLiquidity()

			// Calculate underlying assets from liquidity getting slashed
			asset0PreSlash, asset1PreSlash, err := cl.CalculateUnderlyingAssetsFromPosition(suite.Ctx, positionPreSlash, clPool)
			suite.Require().NoError(err)

			slashAmt := positionPreSlash.Liquidity.Mul(tc.slashPercent)

			// Note the value of the fee accumulator before the slash for the position.
			feeAccumPreSlash, err := suite.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(suite.Ctx, clPoolId)
			suite.Require().NoError(err)
			feePositionKey := cltypes.KeyFeePositionAccumulator(positionId)
			positionSizePreSlash, err := feeAccumPreSlash.GetPositionSize(feePositionKey)
			suite.Require().NoError(err)

			// Note the numShares value of the position at each of the uptime accumulators.
			uptimeAccums, err := suite.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(suite.Ctx, clPoolId)
			uptimePositionKey := string(cltypes.KeyPositionId(positionId))
			suite.Require().NoError(err)
			numShares := make([]sdk.Dec, len(uptimeAccums))
			for i, uptimeAccum := range uptimeAccums {
				position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
				suite.Require().NoError(err)
				numShares[i] = position.NumShares
			}

			// System under test
			clPoolAddress, underlyingAssetsToSlash, err := suite.App.SuperfluidKeeper.PrepareConcentratedLockForSlash(suite.Ctx, lock, slashAmt)

			lowerTickInfoPostSlash, getTickErr := suite.App.ConcentratedLiquidityKeeper.GetTickInfo(suite.Ctx, clPoolId, positionPreSlash.LowerTick)
			suite.Require().NoError(getTickErr)
			upperTickInfoPostSlash, getTickErr := suite.App.ConcentratedLiquidityKeeper.GetTickInfo(suite.Ctx, clPoolId, positionPreSlash.UpperTick)
			suite.Require().NoError(getTickErr)

			clPool, getPoolErr := suite.App.ConcentratedLiquidityKeeper.GetPoolFromPoolIdAndConvertToConcentrated(suite.Ctx, clPoolId)
			suite.Require().NoError(getPoolErr)
			liquidityPostSlash := clPool.GetLiquidity()
			if tc.expectedErr {
				suite.Require().Error(err)

				// Check that the position's fee accumulator has not changed.
				feeAccumPostSlash, err := suite.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(suite.Ctx, clPoolId)
				suite.Require().NoError(err)
				positionSizePostSlash, err := feeAccumPostSlash.GetPositionSize(feePositionKey)
				suite.Require().NoError(err)
				suite.Require().Equal(positionSizePreSlash.String(), positionSizePostSlash.String())

				// Check that the numShares value of the position has not been updated in each of the uptime accumulators.
				uptimeAccums, err := suite.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(suite.Ctx, clPoolId)
				suite.Require().NoError(err)
				for i, uptimeAccum := range uptimeAccums {
					position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
					suite.Require().NoError(err)
					suite.Require().Equal(numShares[i].String(), position.NumShares.String())
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(clPool.GetAddress(), clPoolAddress)

				suite.Require().Equal(lowerTickInfoPreSlash.LiquidityGross.Sub(slashAmt).String(), lowerTickInfoPostSlash.LiquidityGross.String())
				suite.Require().Equal(upperTickInfoPreSlash.LiquidityGross.Sub(slashAmt).String(), upperTickInfoPostSlash.LiquidityGross.String())

				suite.Require().Equal(lowerTickInfoPreSlash.LiquidityNet.Sub(slashAmt).String(), lowerTickInfoPostSlash.LiquidityNet.String())
				suite.Require().Equal(upperTickInfoPreSlash.LiquidityNet.Add(slashAmt).String(), upperTickInfoPostSlash.LiquidityNet.String())

				suite.Require().Equal(liquidityPreSlash.Sub(slashAmt).String(), liquidityPostSlash.String())

				positionPostSlash, err := suite.App.ConcentratedLiquidityKeeper.GetPosition(suite.Ctx, positionId)
				suite.Require().NoError(err)
				suite.Require().Equal(positionPreSlash.Liquidity.Sub(slashAmt).String(), positionPostSlash.Liquidity.String())

				asset0PostSlash, asset1PostSlash, err := cl.CalculateUnderlyingAssetsFromPosition(suite.Ctx, positionPostSlash, clPool)
				suite.Require().NoError(err)

				errTolerance := osmomath.ErrTolerance{
					AdditiveTolerance: sdk.NewDec(1),
					// Actual should be greater than expected, so we round up
					RoundingDir: osmomath.RoundUp,
				}

				suite.Require().Equal(0, errTolerance.Compare(asset0PreSlash.Sub(asset0PostSlash).Amount, underlyingAssetsToSlash[0].Amount))
				suite.Require().Equal(0, errTolerance.Compare(asset1PreSlash.Sub(asset1PostSlash).Amount, underlyingAssetsToSlash[1].Amount))

				// Check that the fee accumulator has been updated by the amount it was slashed.
				feeAccumPostSlash, err := suite.App.ConcentratedLiquidityKeeper.GetFeeAccumulator(suite.Ctx, clPoolId)
				suite.Require().NoError(err)
				positionSizePostSlash, err := feeAccumPostSlash.GetPositionSize(feePositionKey)
				suite.Require().NoError(err)
				suite.Require().Equal(positionSizePreSlash.Sub(slashAmt).String(), positionSizePostSlash.String())

				// Check that the numShares value of the position has been updated in each of the uptime accumulators.
				uptimeAccums, err := suite.App.ConcentratedLiquidityKeeper.GetUptimeAccumulators(suite.Ctx, clPoolId)
				suite.Require().NoError(err)
				for i, uptimeAccum := range uptimeAccums {
					position, err := accum.GetPosition(uptimeAccum, uptimePositionKey)
					suite.Require().NoError(err)
					suite.Require().Equal(numShares[i].Sub(slashAmt).String(), position.NumShares.String())
				}
			}
		})
	}
}
