package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
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
	type superfluidDelegation struct {
		valIndex int64
		lpDenom  string
	}
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
			valAddrs := []sdk.ValAddress{}
			for _, status := range tc.validatorStats {
				valAddr := suite.SetupValidator(status)
				valAddrs = append(valAddrs, valAddr)
			}

			intermediaryAccs := []types.SuperfluidIntermediaryAccount{}

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
				expAcc := types.SuperfluidIntermediaryAccount{
					Denom:   lock.Coins[0].Denom,
					ValAddr: valAddr.String(),
				}

				// check delegation from intermediary account to validator
				delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(delegation.Shares, sdk.NewDec(19000000)) // 95% x 20 x 1000000

				intermediaryAccs = append(intermediaryAccs, expAcc)
			}

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
