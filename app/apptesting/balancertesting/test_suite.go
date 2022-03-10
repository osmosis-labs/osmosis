package balancertesting

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type BalancerTestHelper struct {
	*apptesting.KeeperTestHelper
}

func (balancerTestHelper *BalancerTestHelper) SetupGammPoolsWithBondDenomMultiplier(multipliers []sdk.Dec) []gammtypes.PoolI {
	balancerTestHelper.App.GAMMKeeper.SetParams(balancerTestHelper.Ctx, gammtypes.Params{
		PoolCreationFee: sdk.Coins{},
	})

	bondDenom := balancerTestHelper.App.StakingKeeper.BondDenom(balancerTestHelper.Ctx)
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	pools := []gammtypes.PoolI{}

	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)

		uosmoAmount := gammtypes.InitPoolSharesSupply.ToDec().Mul(multiplier).RoundInt()

		err := simapp.FundAccount(balancerTestHelper.App.BankKeeper, balancerTestHelper.Ctx, acc1, sdk.NewCoins(
			sdk.NewCoin(bondDenom, uosmoAmount.Mul(sdk.NewInt(10))),
			sdk.NewInt64Coin(token, 100000),
		))
		balancerTestHelper.NoError(err)

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(bondDenom, uosmoAmount),
			}
			defaultBarAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(token, sdk.NewInt(10000)),
			}
			poolAssets []gammtypes.PoolAsset = []gammtypes.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolId, err := balancerTestHelper.App.GAMMKeeper.CreateBalancerPool(balancerTestHelper.Ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, defaultFutureGovernor)
		balancerTestHelper.Require().NoError(err)

		pool, err := balancerTestHelper.App.GAMMKeeper.GetPool(balancerTestHelper.Ctx, poolId)
		balancerTestHelper.Require().NoError(err)

		pools = append(pools, pool)
	}
	return pools
}

// SwapAndSetSpotPrice runs a swap to set Spot price of a pool using arbitrary values
// returns spot price after the arbitrary swap
func (balancerTestHelper *BalancerTestHelper) SwapAndSetSpotPrice(poolId uint64, fromAsset gammtypes.PoolAsset, toAsset gammtypes.PoolAsset) sdk.Dec {
	// create a dummy account
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	// fund dummy account with tokens to swap
	coins := sdk.Coins{sdk.NewInt64Coin(fromAsset.Token.Denom, 100000000000000)}
	err := simapp.FundAccount(balancerTestHelper.App.BankKeeper, balancerTestHelper.Ctx, acc1, coins)
	balancerTestHelper.Require().NoError(err)

	_, _, err = balancerTestHelper.App.GAMMKeeper.SwapExactAmountOut(
		balancerTestHelper.Ctx, acc1,
		poolId, fromAsset.Token.Denom, fromAsset.Token.Amount,
		sdk.NewCoin(toAsset.Token.Denom, toAsset.Token.Amount.Quo(sdk.NewInt(4))))
	balancerTestHelper.Require().NoError(err)

	spotPrice, err := balancerTestHelper.App.GAMMKeeper.CalculateSpotPrice(balancerTestHelper.Ctx, poolId, toAsset.Token.Denom, fromAsset.Token.Denom)
	balancerTestHelper.Require().NoError(err)
	return spotPrice
}
