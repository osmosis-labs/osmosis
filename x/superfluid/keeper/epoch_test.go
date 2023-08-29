package keeper_test

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cltypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v19/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
)

func (s *KeeperTestSuite) TestUpdateOsmoEquivalentMultipliers() {
	testCases := []struct {
		name                  string
		asset                 types.SuperfluidAsset
		expectedMultiplier    sdk.Dec
		removeStakingAsset    bool
		poolDoesNotExist      bool
		expectedError         error
		expectedZeroMultipler bool
	}{
		{
			name:               "update LP token Osmo equivalent successfully",
			asset:              types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedMultiplier: sdk.MustNewDecFromStr("0.01"),
		},
		{
			name:             "update LP token Osmo equivalent with pool unexpectedly deleted",
			asset:            types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			poolDoesNotExist: true,
			expectedError:    gammtypes.PoolDoesNotExistError{PoolId: 1},
		},
		{
			name:               "update LP token Osmo equivalent with pool unexpectedly removed Osmo",
			asset:              types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			removeStakingAsset: true,
			expectedError:      errors.New("pool 1 has zero OSMO amount"),
		},
		{
			name:               "update concentrated share Osmo equivalent successfully",
			asset:              types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			expectedMultiplier: sdk.MustNewDecFromStr("1"),
		},
		{
			name:             "update concentrated share Osmo equivalent with pool unexpectedly deleted",
			asset:            types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			poolDoesNotExist: true,
			// Note: this does not error since CL errors are surrounded in `ApplyFuncIfNoError`
			expectedZeroMultipler: true,
		},
		{
			name:               "update concentrated share Osmo equivalent with pool unexpectedly removed Osmo",
			asset:              types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			removeStakingAsset: true,
			// Note: this does not error since CL errors are surrounded in `ApplyFuncIfNoError`
			expectedZeroMultipler: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.Ctx
			superfluidKeeper := s.App.SuperfluidKeeper

			// Switch the default staking denom to something else if the test case requires it
			stakeDenom := s.App.StakingKeeper.BondDenom(ctx)
			if tc.removeStakingAsset {
				stakeDenom = "bar"
			}
			poolCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(1000000000000000000)), sdk.NewCoin("foo", sdk.NewInt(1000000000000000000)))

			// Ensure that the multiplier is zero before the test
			multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)
			s.Require().Equal(multiplier, sdk.ZeroDec())

			// Create the respective pool if the test case requires it
			if !tc.poolDoesNotExist {
				if tc.asset.AssetType == types.SuperfluidAssetTypeLPShare {
					s.PrepareBalancerPoolWithCoins(poolCoins...)
				} else if tc.asset.AssetType == types.SuperfluidAssetTypeConcentratedShare {
					s.PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(stakeDenom, "foo")
				}
			}

			// System under test
			err := superfluidKeeper.UpdateOsmoEquivalentMultipliers(ctx, tc.asset, 1)

			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())

				// Ensure unwind superfluid asset is called
				// Check that multiplier was not set
				multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)
				s.Require().Equal(multiplier, sdk.ZeroDec())
				// Check that the asset was deleted
				_, err := superfluidKeeper.GetSuperfluidAsset(ctx, tc.asset.Denom)
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				// Check that multiplier was set correctly
				multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)

				if !tc.expectedZeroMultipler {
					s.Require().NotEqual(multiplier, sdk.ZeroDec())
				} else {
					// Zero on success is expected on CL errors since those are surrounded with `ApplyFuncIfNoError`
					s.Require().Equal(multiplier, sdk.ZeroDec())
				}
			}
		})
	}
}

type gaugeChecker struct {
	intermediaryAccIndex     uint64
	valIndex                 int64
	lockIndexes              []int64
	lpIndex                  int64
	rewarded                 bool
	expectedDistributedCoins sdk.Coins
}
type distributionTestCase struct {
	name             string
	validatorStats   []stakingtypes.BondStatus
	superDelegations []superfluidDelegation
	rewardedVals     []int64
	gaugeChecks      []gaugeChecker
}

var (
	// distributed coin when there is one account receiving from one gauge
	defaultSingleLockDistributedCoins = sdk.NewCoins(sdk.NewInt64Coin("stake", 19999))
	// distributed coins when there is two account receiving from one gauge
	defaultTwoLockDistributedCoins = sdk.NewCoins(sdk.NewInt64Coin("stake", 9999))
	// distributed coins when there is one account receiving from two gauge
	defaultTwoGaugeDistributedCoins = sdk.NewCoins(sdk.NewInt64Coin("stake", 19998))
	distributionTestCases           = []distributionTestCase{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, []int64{0}, 0, true, defaultSingleLockDistributedCoins}},
		},
		{
			"two LP tokens delegation to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, []int64{0}, 0, true, defaultTwoLockDistributedCoins}, {1, 0, []int64{0}, 1, true, defaultTwoLockDistributedCoins}},
		},
		{
			"one LP token with two locks to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, []int64{0, 1}, 0, true, defaultTwoGaugeDistributedCoins}},
		},
		// In this case, allocate reward to validators with different stat.
		// There is no difference between Bonded, Unbonding, Unbonded
		{
			"add unbonded, unbonding validator case",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}, {2, 2, 0, 1000000}},
			[]int64{0, 1, 2},
			[]gaugeChecker{
				{0, 0, []int64{0}, 0, true, defaultSingleLockDistributedCoins},
				{1, 1, []int64{1}, 0, true, defaultSingleLockDistributedCoins},
				{2, 2, []int64{2}, 0, true, defaultSingleLockDistributedCoins}},
		},
		// Do not allocate rewards to the Unbonded validator. Therefore gauges are not distributed
		{
			"Unallocate to Unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, []int64{0}, 0, true, defaultSingleLockDistributedCoins}, {1, 1, []int64{1}, 0, false, defaultSingleLockDistributedCoins}},
		},
	}
)

func (s *KeeperTestSuite) TestMoveSuperfluidDelegationRewardToGauges() {
	for _, tc := range distributionTestCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

			// allocate rewards to designated validators
			for _, valIndex := range tc.rewardedVals {
				s.AllocateRewardsToValidator(valAddrs[valIndex], sdk.NewInt(20000))
			}

			// move intermediary account delegation rewards to gauges
			s.App.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(s.Ctx)

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check gauge balance
			for _, gaugeCheck := range tc.gaugeChecks {
				gaugeId := intermediaryAccs[gaugeCheck.intermediaryAccIndex].GaugeId
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
				s.Require().NoError(err)
				s.Require().Equal(gauge.Id, gaugeId)
				s.Require().Equal(gauge.IsPerpetual, true)
				s.Require().Equal(lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(denoms[gaugeCheck.lpIndex], valAddrs[gaugeCheck.valIndex].String()),
					Duration:      unbondingDuration,
				}, gauge.DistributeTo)
				if gaugeCheck.rewarded {
					s.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsPositive())
				} else {
					s.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsZero())
				}
				s.Require().Equal(gauge.StartTime, s.Ctx.BlockTime())
				s.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				s.Require().Equal(gauge.FilledEpochs, uint64(0))
				s.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))
			}
		})
	}
}

func (s *KeeperTestSuite) TestDistributeSuperfluidGauges() {
	changeRewardReceiverTestCases := []bool{true, false}
	for _, tc := range distributionTestCases {
		// run distributionTestCases two times.
		// Once with lock reward reciver as owner,
		// Second time with lock reward receiver as a different account.
		for _, changeRewardReceiver := range changeRewardReceiverTestCases {
			tc := tc

			s.Run(tc.name, func() {
				s.SetupTest()
				// create one more account to set reward receiver as arbitrary account
				thirdTestAcc := CreateRandomAccounts(1)
				s.TestAccs = append(s.TestAccs, thirdTestAcc...)
				// setup validators
				valAddrs := s.SetupValidators(tc.validatorStats)

				denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

				// setup superfluid delegations
				delAddresses, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)

				// if test setting is changing reward receiver,
				// we iterate over all locks and change the reward receiver to owner with index of + 1
				if changeRewardReceiver {
					for _, lock := range locks {
						var newRewardReceiver string
						if lock.Owner == s.TestAccs[0].String() || lock.Owner == delAddresses[0].String() {
							newRewardReceiver = s.TestAccs[1].String()
						} else if lock.Owner == s.TestAccs[1].String() || lock.Owner == delAddresses[1].String() {
							newRewardReceiver = s.TestAccs[2].String()
						} else {
							newRewardReceiver = s.TestAccs[3].String()
						}
						err := s.App.LockupKeeper.SetLockRewardReceiverAddress(s.Ctx, lock.ID, lock.OwnerAddress(), newRewardReceiver)
						s.Require().NoError(err)
					}
				}

				// allocate rewards to designated validators
				for _, valIndex := range tc.rewardedVals {
					s.AllocateRewardsToValidator(valAddrs[valIndex], sdk.NewInt(20000))
				}

				// move intermediary account delegation rewards to gauges
				s.App.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(s.Ctx)

				// move gauges to active gauge by declaring epoch end
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))
				epochId := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Identifier
				err := s.App.IncentivesKeeper.AfterEpochEnd(s.Ctx, epochId, 1)
				s.Require().NoError(err)

				// create a map of delegator index -> number of locks eligible for distribution.
				// This is used to check amount of coins distributed for the delegator in future check.
				lockIndexDistributionNumMap := make(map[int64]int64)
				for _, gaugeCheck := range tc.gaugeChecks {
					for _, lockIndex := range gaugeCheck.lockIndexes {
						lockIndexDistributionNumMap[lockIndex]++
					}
				}

				// system under test
				s.App.SuperfluidKeeper.DistributeSuperfluidGauges(s.Ctx)

				for _, gaugeCheck := range tc.gaugeChecks {
					gaugeId := intermediaryAccs[gaugeCheck.intermediaryAccIndex].GaugeId
					gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
					s.Require().NoError(err)
					s.Require().Equal(gauge.Id, gaugeId)
					s.Require().Equal(gauge.IsPerpetual, true)
					s.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))

					bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

					moduleAddress := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
					moduleBalanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddress, bondDenom)

					if gaugeCheck.rewarded {
						s.Require().Equal(gauge.FilledEpochs, uint64(1))
						s.Require().Equal(gaugeCheck.expectedDistributedCoins, gauge.DistributedCoins)
						s.Require().Equal(gauge.Coins.Sub(gauge.DistributedCoins).AmountOf(bondDenom), moduleBalanceAfter.Amount)

						// iterate over delegator index that received incentive from this gauge and check balance
						for _, lockIndex := range gaugeCheck.lockIndexes {
							lock := locks[lockIndex]

							// get updated lock from state
							updatedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
							s.Require().NoError(err)

							// reward receiver is lock owner if lock owner is an empty string literal
							rewardReceiver := updatedLock.RewardReceiverAddress
							if rewardReceiver == "" {
								rewardReceiver = lock.Owner
							}

							// check balance of the reward receiver
							rewardReceiverBalance := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(rewardReceiver), bondDenom)

							// reward receiver balance should be
							// gauge distributed coin / amount of locks for that gauge * num of gauges delegator is getting rewards from
							numOfDistribution := lockIndexDistributionNumMap[lockIndex]
							expectedDelegatorBalance := gauge.DistributedCoins.AmountOf(bondDenom).Int64() / int64(len(gaugeCheck.lockIndexes)) * numOfDistribution
							s.Require().Equal(expectedDelegatorBalance, rewardReceiverBalance.Amount.Int64())
						}
					} else {
						s.Require().Equal(gauge.FilledEpochs, uint64(0))
						s.Require().Equal(sdk.Coins(nil), gauge.DistributedCoins)
						for _, lockIndex := range gaugeCheck.lockIndexes {
							lock := locks[lockIndex]

							// get updated lock from state
							updatedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
							s.Require().NoError(err)

							rewardReceiver := updatedLock.RewardReceiverAddress
							if rewardReceiver == "" {
								rewardReceiver = lock.Owner
							}
							delegatorBalance := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(rewardReceiver), bondDenom)
							s.Require().Equal(sdk.ZeroInt(), delegatorBalance.Amount)
						}
					}
				}
			})
		}
	}
}
