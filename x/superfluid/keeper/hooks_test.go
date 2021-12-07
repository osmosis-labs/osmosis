package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func (suite *KeeperTestSuite) TestSuperfluidAfterEpochEnd() {
	valAddr, lock := suite.SetupSuperfluidDelegate()

	expAcc := types.SuperfluidIntermediaryAccount{
		Denom:   lock.Coins[0].Denom,
		ValAddr: valAddr.String(),
	}

	// check delegation from intermediary account to validator
	delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
	suite.Require().True(found)
	suite.Require().Equal(delegation.Shares, sdk.NewDec(1900000)) // 95% x 2 x 1000000

	// twap price change before refresh
	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 2, "gamm/pool/1", sdk.NewDec(10))
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1, sdk.Coins{
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewInt64Coin("foo", 100000),
		sdk.NewInt64Coin("bar", 100000),
	})
	suite.Require().NoError(err)
	poolId, err := suite.app.GAMMKeeper.CreateBalancerPool(
		suite.ctx, acc1, gammtypes.BalancerPoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, []gammtypes.PoolAsset{
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(10000)),
			},
		},
		"")
	suite.Require().NoError(err)
	suite.Require().Equal(poolId, uint64(1))

	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 3,
	})

	// run epoch actions
	suite.NotPanics(func() {
		params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
		suite.app.SuperfluidKeeper.AfterEpochEnd(suite.ctx, params.RefreshEpochIdentifier, 3)
	})

	// check delegation changes
	delegation, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, expAcc.GetAddress(), valAddr)
	suite.Require().True(found)
	suite.Require().Equal(delegation.Shares, sdk.NewDec(9500000)) // 95% x 10 x 1000000

	// check lptoken twap value set
	newEpochTwap := suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 3, "gamm/pool/1")
	suite.Require().Equal(newEpochTwap, sdk.NewDec(10000))
}
