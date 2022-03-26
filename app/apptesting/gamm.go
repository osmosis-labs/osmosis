package apptesting

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
)

var gammPoolMakerAcc = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address().Bytes())

// Returns a Univ2 pool with the initial liquidity being the provided balances
func (suite *KeeperTestHelper) PrepareUni2PoolWithAssets(asset1, asset2 sdk.Coin) uint64 {
	return suite.preparePool(
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

func (suite *KeeperTestHelper) preparePool(assets []balancer.PoolAsset) uint64 {
	suite.Require().Len(assets, 2)

	// Add coins for pool creation fee + coins needed to mint balances
	fundCoins := sdk.Coins{sdk.NewCoin("uosmo", sdk.NewInt(10000000000))}
	for _, a := range assets {
		fundCoins = fundCoins.Add(a.Token)
	}
	simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, gammPoolMakerAcc, fundCoins)

	msg := balancer.NewMsgCreateBalancerPool(gammPoolMakerAcc, balancer.PoolParams{
		SwapFee: sdk.ZeroDec(),
		ExitFee: sdk.ZeroDec(),
	}, assets, "")
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.NoError(err)
	return poolId
}
