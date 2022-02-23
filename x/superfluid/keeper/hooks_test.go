package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestSuperfluidAfterEpochEnd() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		expRewards       sdk.Coins
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			sdk.Coins{sdk.NewCoin("uion", sdk.OneInt())},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			valAddrs := suite.SetupValidators(tc.validatorStats)
			bondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(1)
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			// gamm swap operation before refresh
			suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 2, "gamm/pool/1", sdk.NewDec(10))
			acc1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address().Bytes())

			coins := sdk.Coins{sdk.NewInt64Coin("foo", 100000000000000)}
			err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, acc1, coins)
			suite.Require().NoError(err)
			_, _, err = suite.app.GAMMKeeper.SwapExactAmountOut(suite.ctx, acc1, 1, "foo", sdk.NewInt(100000000000000), sdk.NewInt64Coin(bondDenom, 250000000000))
			suite.Require().NoError(err)

			// run epoch actions
			suite.BeginNewBlock(true)

			// check lptoken twap value set
			newEpochTwap := suite.app.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")
			suite.Require().Equal(newEpochTwap.String(), "0.009999997500000000")

			// check delegation changes
			for _, acc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, acc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(9500))
				// TODO: Check reward distribution
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, acc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRewards, gauge.Coins)
				suite.Require().Equal(tc.expRewards, suite.app.BankKeeper.GetAllBalances(suite.ctx, delAddrs[0]))
				suite.Require().Equal(tc.expRewards, suite.app.BankKeeper.GetAllBalances(suite.ctx, acc.GetAccAddress()))
			}
		})
	}
}

// func (suite *KeeperTestSuite) TestOnStartUnlock() {
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
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and single undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and multiple undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"with multiple validators and multiple superfluid delegations and multiple undelegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 1, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"undelegating not available lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{2},
// 			[]bool{true},
// 		},
// 		{
// 			"try undelegating twice for same lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 1},
// 			[]bool{false, true},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			suite.Require().Equal(poolId, uint64(1))

// 			// Generate delegator addresses
// 			delAddrs := CreateRandomAccounts(1)

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			for index, lockId := range tc.unbondingLockIds {
// 				// get intermediary account
// 				accAddr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
// 				intermediaryAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, accAddr)
// 				valAddr := intermediaryAcc.ValAddr

// 				// unlock native lockup
// 				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
// 				if err == nil {
// 					err = suite.app.LockupKeeper.BeginUnlock(suite.ctx, *lock, nil)
// 				}

// 				if tc.expUnbondingErr[index] {
// 					suite.Require().Error(err)
// 					continue
// 				}
// 				suite.Require().NoError(err)

// 				// check lockId and intermediary account connection deletion
// 				addr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
// 				suite.Require().Equal(addr.String(), "")

// 				// check bonding synthetic lockup deletion
// 				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				unbondingDuration := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
// 				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
// 				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(unbondingDuration))
// 			}
// 		})
// 	}
// }

func (suite *KeeperTestSuite) TestBeforeSlashingUnbondingDelegationHook() {
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
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]uint64{1},
			[]int64{0},
			[]uint64{1},
			[]uint64{},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 0, "gamm/pool/1", 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1, 2},
			[]uint64{},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 1, "gamm/pool/1", 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1},
			[]uint64{2},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			slashFactor := sdk.NewDecWithPrec(5, 2)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)
			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)

				// superfluid undelegate
				err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.ctx = suite.ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)
				suite.app.StakingKeeper.Slash(suite.ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeSlashingUnbondingDelegation hook
			}

			// check slashed lockups
			for _, lockId := range tc.expSlashedLockIds {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewInt(950000).String(), gotLock.Coins.AmountOf("gamm/pool/1").String())
			}

			// check unslashed lockups
			for _, lockId := range tc.expUnslashedLockIds {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewInt(1000000).String(), gotLock.Coins.AmountOf("gamm/pool/1").String())
			}
		})
	}
}
