package apptesting

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

var (
	defaultFutureGovernor = ""
	dummyAcc              = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func (suite *KeeperTestHelper) PrepareBalancerPoolWithPoolParams(PoolParams balancer.PoolParams) uint64 {
	// Mint some assets to the accounts.
	err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, dummyAcc, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	))
	if err != nil {
		panic(err)
	}

	poolAssets := []balancertypes.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(300),
			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
		},
	}
	msg := balancer.NewMsgCreateBalancerPool(dummyAcc, PoolParams, poolAssets, defaultFutureGovernor)
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.NoError(err)
	return poolId
}

func (suite *KeeperTestHelper) PrepareBalancerPool() uint64 {
	poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	})

	spotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "foo", "bar")
	suite.NoError(err)
	suite.Equal(sdk.NewDec(2).String(), spotPrice.String())
	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "bar", "baz")
	suite.NoError(err)
	suite.Equal(sdk.NewDecWithPrec(15, 1).String(), spotPrice.String())
	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "baz", "foo")
	suite.NoError(err)
	suite.Equal(sdk.NewDec(1).Quo(sdk.NewDec(3)).String(), spotPrice.String())

	return poolId
}
