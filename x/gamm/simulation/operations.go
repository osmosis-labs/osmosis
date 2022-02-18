package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/v7/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreatePool              = "op_weight_create_pool"
	OpWeightMsgSwapExactAmountIn       = "op_weight_swap_exact_amount_in"
	OpWeightMsgSwapExactAmountOut      = "op_weight_swap_exact_amount_out"
	OpWeightMsgJoinPool                = "op_weight_join_pool"
	OpWeightMsgExitPool                = "op_weight_exit_pool"
	OpWeightMsgJoinSwapExternAmountIn  = "op_weight_join_swap_extern_amount_in"
	OpWeightMsgJoinSwapShareAmountOut  = "op_weight_join_swap_share_amount_out"
	OpWeightMsgExitSwapExternAmountOut = "op_weight_exit_swap_extern_amount_out"
	OpWeightMsgExitSwapShareAmountIn   = "op_weight_exit_swap_share_amount_in"

	DefaultWeightMsgCreatePool              int = 10
	DefaultWeightMsgSwapExactAmountIn       int = 25
	DefaultWeightMsgSwapExactAmountOut      int = 10
	DefaultWeightMsgJoinPool                int = 10
	DefaultWeightMsgExitPool                int = 10
	DefaultWeightMsgJoinSwapExternAmountIn  int = 10
	DefaultWeightMsgJoinSwapShareAmountOut  int = 10
	DefaultWeightMsgExitSwapExternAmountOut int = 10
	DefaultWeightMsgExitSwapShareAmountIn   int = 10
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePool        int
		weightMsgSwapExactAmountIn int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePool, &weightMsgCreatePool, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePool = simappparams.DefaultWeightMsgCreateValidator
			weightMsgSwapExactAmountIn = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePool,
			SimulateMsgCreateBalancerPool(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSwapExactAmountIn,
			SimulateMsgSwapExactAmountIn(ak, bk, k),
		),
	}
}

func genFuturePoolGovernor(r *rand.Rand, addr sdk.Address, tokenList []string) string {
	choice := r.Int31n(4)
	if choice == 0 { // No governor
		return ""
	} else if choice == 1 { // Single address governor
		return addr.String()
	} else if choice == 2 { // LP token governor
		return "1d"
	} else { // Other token governor
		token := tokenList[r.Intn(len(tokenList))]
		return token + ",1d"
	}
}

func genPoolAssets(r *rand.Rand, acct simtypes.Account, coins sdk.Coins) []types.PoolAsset {
	// selecting random number between [2, Min(coins.Len, 6)]
	numCoins := 2 + r.Intn(Min(coins.Len(), 6)-1)
	denomIndices := r.Perm(coins.Len())
	assets := []types.PoolAsset{}
	for _, denomIndex := range denomIndices[:numCoins] {
		denom := coins[denomIndex].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[denomIndex].Amount.QuoRaw(100))
		reserveAmt := sdk.NewCoin(denom, amt)
		weight := sdk.NewInt(r.Int63n(9) + 1)
		assets = append(assets, types.PoolAsset{Token: reserveAmt, Weight: weight})
	}

	return assets
}

func genBalancerPoolParams(r *rand.Rand, blockTime time.Time, assets []types.PoolAsset) balancer.PoolParams {
	// swapFeeInt := int64(r.Intn(1e5))
	// swapFee := sdk.NewDecWithPrec(swapFeeInt, 6)

	exitFeeInt := int64(r.Intn(1e5))
	exitFee := sdk.NewDecWithPrec(exitFeeInt, 6)

	// TODO: Randomly generate LBP params
	return balancer.PoolParams{
		// SwapFee:                  swapFee,
		SwapFee: sdk.ZeroDec(),
		ExitFee: exitFee,
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// SimulateMsgCreateBalancerPool generates a MsgCreatePool with random values
func SimulateMsgCreateBalancerPool(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 1 {
			return simtypes.NoOpMsg(
				types.ModuleName, balancer.TypeMsgCreateBalancerPool, "Account doesn't have 2 different coin types"), nil, nil
		}

		poolAssets := genPoolAssets(r, simAccount, simCoins)
		poolParams := genBalancerPoolParams(r, ctx.BlockTime(), poolAssets)

		// Commented out as genFuturePoolGovernor() panics on empty denom slice.
		// TODO: fix and provide proper denom types.
		// TODO: Replace []string{} with all token types on chain.
		// futurePoolGovernor := genFuturePoolGovernor(r, simAccount.Address, []string{})

		balances := bk.GetAllBalances(ctx, simAccount.Address)
		denoms := make([]string, len(balances))
		for i := range balances {
			denoms[i] = balances[i].Denom
		}

		// set the pool params to set the pool creation fee to dust amount of denom
		k.SetParams(ctx, types.Params{
			PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(denoms[0], 1)},
		})

		msg := &balancer.MsgCreateBalancerPool{
			Sender:             simAccount.Address.String(),
			PoolParams:         &poolParams,
			PoolAssets:         poolAssets,
			FuturePoolGovernor: "",
		}

		spentCoins := types.PoolAssetsCoins(poolAssets)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, msg, spentCoins, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

// SimulateMsgSwapExactAmountIn generates a MsgSwapExactAmountIn with random values
// TODO: Change to use expected keepers
func SimulateMsgSwapExactAmountIn(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSwapExactAmountIn, "Account have no coin"), nil, nil
		}

		coin := simCoins[r.Intn(len(simCoins))]
		// Use under 0.5% of the account balance
		// TODO: Make like a 33% probability of using a ton of balance
		amt, _ := simtypes.RandPositiveInt(r, coin.Amount.QuoRaw(200))

		tokenIn := sdk.Coin{
			Denom:  coin.Denom,
			Amount: amt,
		}

		routes, _ := RandomExactAmountInRoute(ctx, r, k, tokenIn)
		if len(routes) == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSwapExactAmountIn, "No pool exist"), nil, nil
		}

		msg := types.MsgSwapExactAmountIn{
			Sender:            simAccount.Address.String(),
			Routes:            routes,
			TokenIn:           tokenIn,
			TokenOutMinAmount: sdk.OneInt(),
			// TokenOutMinAmount: tokenOutMin.QuoRaw(2),
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, sdk.Coins{tokenIn}, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func RandomExactAmountInRoute(ctx sdk.Context, r *rand.Rand, k keeper.Keeper, tokenIn sdk.Coin) (res []types.SwapAmountInRoute, tokenOut sdk.Coin) {
	routeLen := r.Intn(1) + 1

	allpools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}

	pools := []types.PoolI{}
	for _, pool := range allpools {
		if pool.IsActive(ctx.BlockTime()) {
			pools = append(pools, pool)
		}
	}

	if len(pools) == 0 {
		return
	}

	res = []types.SwapAmountInRoute{}
	for i := 0; i < routeLen; i++ {
		// randomly selected pool might not include the source token, retry
		for retry := 0; retry < 10; retry++ {
			pool := pools[r.Intn(len(pools))]
			inAsset, err := pool.GetPoolAsset(tokenIn.Denom)
			if err != nil {
				continue
			}
			if inAsset.Token.Amount.LT(tokenIn.Amount) {
				continue
			}
			for _, asset := range pool.GetAllPoolAssets() {
				if asset.Token.Denom == tokenIn.Denom {
					continue
				}
				res = append(res, types.SwapAmountInRoute{
					PoolId:        pool.GetId(),
					TokenOutDenom: asset.Token.Denom,
				})
				sp, err := k.CalculateSpotPriceWithSwapFee(ctx, pool.GetId(), tokenIn.Denom, asset.Token.Denom)
				if err != nil {
					panic(err)
				}
				amt := tokenIn.Amount.ToDec().Quo(sp).RoundInt()
				tokenIn = sdk.Coin{
					Denom:  asset.Token.Denom,
					Amount: amt,
				}
				break
			}
			break
		}
	}

	tokenOut = tokenIn
	return
}
