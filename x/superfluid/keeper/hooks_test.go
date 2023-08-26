package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	lockupkeeper "github.com/osmosis-labs/osmosis/v19/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v19/x/superfluid/types"
)

func (s *KeeperTestSuite) TestSuperfluidAfterEpochEnd() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		delegatorNumber  int
		superDelegations []superfluidDelegation
		expRewards       []sdk.Coins
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			// bond denom staked in pool = 15_000_000
			// with risk adjustment, the actual bond denom staked via superfluid would be 15_000_000 * (1 - 0.5) = 7_500_000
			// we do an arbitrary swap to set spot price, which adjusts superfluid staked equivalent base denom 20_000_000 * (1 - 0.5) = 10_000_000 during begin block
			// delegation rewards are calculated using the equation (current period cumulative reward ratio - last period cumulative reward ratio) * asset amount
			// in this test case, the calculation for expected reward would be the following (0.99999 - 0) * 10_000_000
			// thus we expect 999_990 stake as rewards
			[]sdk.Coins{{sdk.NewCoin("stake", sdk.NewInt(999990))}},
		},
		{
			"happy path with two validator and delegator each",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			// reward for the first block propser / lock 0 that has been superfluid staked would be equivalent to calculations done above
			// 999_990 stake as rewards.
			// reward for the second delegation is expected to be different. Amount superfluid staked would be equivalently 7_500_000 stake.
			// This would be the first block propsed by the second validator, current period cumulative reward ratio being 999_86.66684,
			// last period cumulative reward ratio being 0
			// Thus as rewards, we expect 999986stake, calculted using the following equation: (999_86.66684 - 0) * 7_500_000
			[]sdk.Coins{{sdk.NewCoin("stake", sdk.NewInt(999990))}, {sdk.NewCoin("stake", sdk.NewInt(999986))}},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, poolIds := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20)})

			// Generate delegator addresses
			delAddrs, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			// run swap and set spot price
			pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolIds[0])
			s.Require().NoError(err)
			coins := pool.GetTotalPoolLiquidity(s.Ctx)
			s.SwapAndSetSpotPrice(poolIds[0], coins[1], coins[0])

			// run epoch actions
			// run begin block for each validator so that both validator gets block rewards
			for _, valAddr := range valAddrs {
				s.BeginNewBlockWithProposer(true, valAddr)
			}

			// check lptoken twap value set
			newEpochMultiplier := s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, denoms[0])
			s.Require().Equal(newEpochMultiplier, sdk.NewDec(15))

			for index, lock := range locks {
				// check gauge creation in new block
				intermediaryAccAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
				intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, intermediaryAccAddr)
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, intermediaryAcc.GaugeId)
				s.Require().NoError(err)
				s.Require().Equal(gauge.Id, intermediaryAcc.GaugeId)
				s.Require().Equal(gauge.IsPerpetual, true)
				s.Require().Equal(gauge.Coins, tc.expRewards[index])
				s.Require().Equal(gauge.DistributedCoins.String(), tc.expRewards[index].String())
			}

			// check delegation changes
			for _, acc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				s.Require().NoError(err)
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
				s.Require().True(found)
				s.Require().Equal(sdk.NewDec(7500000), delegation.Shares)
			}

			for index, delAddr := range delAddrs {
				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, delAddr)
				s.Require().Equal(tc.expRewards[index], balance)
			}
		})
	}
}

// func (s *KeeperTestSuite) TestOnStartUnlock() {
// 	testCases := []struct {
// 		name             string
// 		validatorStats   []stakingtypes.BondStatus
// 		superDelegations []superfluidDelegation
// 		unbondingLockIds []uint64
// 		expUnbondingErr  []bool
// 	}{
// 		{
// 			"with single validator and single superfluid delegation and single lockup unlock",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and single undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}, {0, 0, DefaultGammAsset, 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and multiple undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}, {0, 0, DefaultGammAsset, 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"with multiple validators and multiple superfluid delegations and multiple undelegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}, {0, 1, DefaultGammAsset, 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"undelegating not available lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}},
// 			[]uint64{2},
// 			[]bool{true},
// 		},
// 		{
// 			"try undelegating twice for same lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, DefaultGammAsset, 1000000}},
// 			[]uint64{1, 1},
// 			[]bool{false, true},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		s.Run(tc.name, func() {
// 			s.SetupTest()

// 			poolId := s.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			s.Require().Equal(poolId, uint64(1))

// 			// Generate delegator addresses
// 			delAddrs := CreateRandomAccounts(1)

// 			// setup validators
// 			valAddrs := s.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := s.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
// 			s.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			for index, lockId := range tc.unbondingLockIds {
// 				// get intermediary account
// 				accAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
// 				intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, accAddr)
// 				valAddr := intermediaryAcc.ValAddr

// 				// unlock native lockup
// 				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
// 				if err == nil {
// 					err = s.App.LockupKeeper.BeginUnlock(s.Ctx, *lock, nil)
// 				}

// 				if tc.expUnbondingErr[index] {
// 					s.Require().Error(err)
// 					continue
// 				}
// 				s.Require().NoError(err)

// 				// check lockId and intermediary account connection deletion
// 				addr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
// 				s.Require().Equal(addr.String(), "")

// 				// check bonding synthetic lockup deletion
// 				_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				s.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
// 				synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				s.Require().NoError(err)
// 				s.Require().Equal(synthLock.UnderlyingLockId, lockId)
// 				s.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				s.Require().Equal(synthLock.EndTime, s.Ctx.BlockTime().Add(unbondingDuration))
// 			}
// 		})
// 	}
// }

func (s *KeeperTestSuite) TestBeforeSlashingUnbondingDelegationHook() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		slashedValIndexes     []int64
		expSlashedLockIds     []uint64
		expUnslashedLockIds   []uint64
	}{
		{
			"happy path with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]int64{0},
			[]uint64{1},
			[]uint64{},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1, 2},
			[]uint64{},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1},
			[]uint64{2},
		},
		{
			"add unbonding validator case",
			[]stakingtypes.BondStatus{stakingtypes.Unbonding, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1},
			[]uint64{2},
		},
		{
			"add unbonded validator case",
			[]stakingtypes.BondStatus{stakingtypes.Unbonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1},
			[]uint64{2},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			slashFactor := sdk.NewDecWithPrec(5, 2)

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

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
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrs[valIndex])
				s.Require().True(found)
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
				// Note: this calls BeforeSlashingUnbondingDelegation hook
			}

			// check slashed lockups
			for _, lockId := range tc.expSlashedLockIds {
				gotLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)
				s.Require().Equal(sdk.NewInt(950000).String(), gotLock.Coins.AmountOf(denoms[0]).String())
			}

			// check unslashed lockups
			for _, lockId := range tc.expUnslashedLockIds {
				gotLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)
				s.Require().Equal(sdk.NewInt(1000000).String(), gotLock.Coins.AmountOf(denoms[0]).String())
			}
		})
	}
}

// TestAfterAddTokensToLock_Event tests that events are correctly emitted
// when calling AfterAddTokensToLock.
func (s *KeeperTestSuite) TestAfterAddTokensToLock_Event() {
	s.SetupTest()

	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20)})

	// setup superfluid delegations
	_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)
	s.checkIntermediaryAccountDelegations(intermediaryAccs)

	for index, lock := range locks {
		lockupMsgServer := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := sdk.WrapSDKContext(s.Ctx)
		coinsToLock := sdk.NewCoins(sdk.NewCoin(denoms[index], sdk.NewInt(100)))
		sender, _ := sdk.AccAddressFromBech32(lock.Owner)
		s.FundAcc(sender, coinsToLock)

		_, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(sender, time.Hour*504, coinsToLock))
		s.Require().NoError(err)

		// should call AfterAddTokensToLock hook and emit event here
		s.AssertEventEmitted(s.Ctx, types.TypeEvtSuperfluidIncreaseDelegation, 1)
	}
}
