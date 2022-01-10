package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/keeper"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func (suite *KeeperTestSuite) createGammPool(denoms []string) uint64 {
	coins := sdk.Coins{}
	poolAssets := []gammtypes.PoolAsset{}
	for _, denom := range denoms {
		coins = coins.Add(sdk.NewInt64Coin(denom, 10000000000))
		poolAssets = append(poolAssets, gammtypes.PoolAsset{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(denom, sdk.NewInt(10000)),
		})
	}

	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1, coins)
	suite.Require().NoError(err)

	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(
		suite.ctx, acc1, gammtypes.BalancerPoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, "")
	suite.Require().NoError(err)

	return poolId
}

func (suite *KeeperTestSuite) TestSuperfluidAfterEpochEnd() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			// gamm swap operation before refresh
			suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 2, "gamm/pool/1", sdk.NewDec(10))
			acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

			err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1, sdk.Coins{sdk.NewInt64Coin("foo", 1000000)})
			suite.Require().NoError(err)
			_, _, err = suite.app.GAMMKeeper.SwapExactAmountOut(suite.ctx, acc1, 1, "foo", sdk.NewInt(1000000), sdk.NewInt64Coin(appparams.BaseCoinUnit, 2500))
			suite.Require().NoError(err)

			// run epoch actions
			suite.NotPanics(func() {
				params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
				suite.app.SuperfluidKeeper.AfterEpochEnd(suite.ctx, params.RefreshEpochIdentifier, 2)
			})

			// check lptoken twap value set
			newEpochTwap := suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 2, "gamm/pool/1")
			suite.Require().Equal(newEpochTwap, sdk.NewDec(7500))

			// check delegation changes
			for _, acc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, acc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(7125000000)) // 95% x 7500 x 1000000
			}
		})
	}
}

func (suite *KeeperTestSuite) TestOnStartUnlock() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		unbondingLockIds []uint64
		expUnbondingErr  []bool
	}{
		{
			"with single validator and single superfluid delegation and single lockup unlock",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1},
			[]bool{false},
		},
		{
			"with single validator and multiple superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]uint64{1},
			[]bool{false},
		},
		{
			"with single validator and multiple superfluid delegations and multiple undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]uint64{1, 2},
			[]bool{false, false},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]uint64{1, 2},
			[]bool{false, false},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{2},
			[]bool{true},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1, 1},
			[]bool{false, true},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for index, lockId := range tc.unbondingLockIds {
				// get intermediary account
				accAddr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				intermediaryAcc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				// unlock native lockup
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				if err == nil {
					err = suite.app.LockupKeeper.BeginUnlock(suite.ctx, *lock)
				}

				if tc.expUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// check lockId and intermediary account connection deletion
				addr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.StakingSuffix(valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, lockId, keeper.UnstakingSuffix(valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.ctx.BlockTime().Add(keeper.SuperfluidUnbondDuration))
			}
		})
	}
}
