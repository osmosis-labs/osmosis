package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var DefaultAcctFunds sdk.Coins = sdk.NewCoins(
	sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
	sdk.NewCoin("foo", sdk.NewInt(10000000)),
	sdk.NewCoin("bar", sdk.NewInt(10000000)),
	sdk.NewCoin("baz", sdk.NewInt(10000000)),
)

// Returns a Univ2 pool with the initial liquidity being the provided balances
func (suite *KeeperTestHelper) PrepareUni2PoolWithAssets(asset1, asset2 sdk.Coin) uint64 {
	return suite.PrepareBalancerPoolWithPoolAsset(
		[]balancer.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  asset1,
			},
			{
				Weight: sdk.NewInt(1),
				Token:  asset2,
			},
		},
	)
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
	s := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := s.MulInt(gammtypes.SigFigs).RoundInt().ToDec().QuoInt(gammtypes.SigFigs)
	suite.Equal(sp.String(), spotPrice.String())

	return poolId
}

func (suite *KeeperTestHelper) PrepareBalancerPoolWithPoolParams(poolParams balancer.PoolParams) uint64 {
	// Mint some assets to the account.
	suite.FundAcc(suite.TestAccs[0], DefaultAcctFunds)

	poolAssets := []balancer.PoolAsset{
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
	msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], poolParams, poolAssets, "")
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.NoError(err)
	return poolId
}

func (suite *KeeperTestHelper) PrepareBalancerPoolWithPoolAsset(assets []balancer.PoolAsset) uint64 {
	suite.Require().Len(assets, 2)

	// Add coins for pool creation fee + coins needed to mint balances
	fundCoins := sdk.Coins{sdk.NewCoin("uosmo", sdk.NewInt(10000000000))}
	for _, a := range assets {
		fundCoins = fundCoins.Add(a.Token)
	}
	suite.FundAcc(suite.TestAccs[0], fundCoins)

	msg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], balancer.PoolParams{
		SwapFee: sdk.ZeroDec(),
		ExitFee: sdk.ZeroDec(),
	}, assets, "")
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.NoError(err)
	return poolId
}

func (suite *KeeperTestHelper) PrepareBalancerPoolWithSwapFee() uint64 {
	poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.MustNewDecFromStr("0.001"),
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
	s := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := s.MulInt(gammtypes.SigFigs).RoundInt().ToDec().QuoInt(gammtypes.SigFigs)
	suite.Equal(sp.String(), spotPrice.String())

	return poolId
}
