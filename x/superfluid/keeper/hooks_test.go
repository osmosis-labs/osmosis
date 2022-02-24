package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
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
			sdk.Coins{},
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
			acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

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
				suite.Require().Equal(sdk.NewDec(5000), delegation.Shares)
				// TODO: Check reward distribution
			}
		})
	}
}

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
