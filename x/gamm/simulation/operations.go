package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	osmo_simulation "github.com/osmosis-labs/osmosis/x/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/osmosis/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
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
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak stakingTypes.AccountKeeper,
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
			SimulateMsgCreatePool(ak, bk, k),
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
	denomIndices := r.Perm(numCoins)
	assets := []types.PoolAsset{}
	for i := 0; i < numCoins; i++ {
		denom := coins[denomIndices[i]].Denom
		amt, _ := simtypes.RandPositiveInt(r, coins[i].Amount)
		reserveAmt := sdk.NewCoin(denom, amt)
		weight := sdk.OneInt()
		assets = append(assets, types.PoolAsset{Token: reserveAmt, Weight: weight})
	}

	return assets
}

func genPoolParams(r *rand.Rand, blockTime time.Time, assets []types.PoolAsset) types.PoolParams {
	swapFeeInt := int64(r.Intn(1e5))
	swapFee := sdk.NewDecWithPrec(swapFeeInt, 5)

	exitFeeInt := int64(r.Intn(1e5))
	exitFee := sdk.NewDecWithPrec(exitFeeInt, 5)

	// TODO: Randomly generate LBP params
	return types.PoolParams{
		SwapFee:                  swapFee,
		ExitFee:                  exitFee,
		SmoothWeightChangeParams: nil,
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

// SimulateMsgCreatePool generates a MsgCreatePool with random values
func SimulateMsgCreatePool(ak stakingTypes.AccountKeeper, bk stakingTypes.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		simCoins := bk.SpendableCoins(ctx, simAccount.Address)
		if simCoins.Len() <= 1 {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgCreatePool, "Account doesn't have 2 different coin types"), nil, nil
		}

		poolAssets := genPoolAssets(r, simAccount, simCoins)
		poolParams := genPoolParams(r, ctx.BlockTime(), poolAssets)

		// TODO: Replace []string{} with all token types on chain.
		futurePoolGovernor := genFuturePoolGovernor(r, simAccount.Address, []string{})
		msg := types.MsgCreatePool{
			Sender:             simAccount.Address.String(),
			FuturePoolGovernor: futurePoolGovernor,
			PoolAssets:         poolAssets,
			PoolParams:         poolParams,
		}

		spentCoins := types.PoolAssetsCoins(poolAssets)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, spentCoins, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

// SimulateMsgSwapExactAmountIn generates a MsgSwapExactAmountIn with random values
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
		amt, _ := simtypes.RandPositiveInt(r, coin.Amount)

		tokenIn := sdk.Coin{
			Denom:  coin.Denom,
			Amount: amt,
		}

		routes, tokenOut := RandomExactAmountInRoute(ctx, r, k, tokenIn)
		if routes == nil {
			return simtypes.NoOpMsg(
				types.ModuleName, types.TypeMsgSwapExactAmountIn, "No pool exist"), nil, nil
		}

		tokenOutMin, _ := simtypes.RandPositiveInt(r, tokenOut.Amount)

		msg := types.MsgSwapExactAmountIn{
			Sender:            simAccount.Address.String(),
			Routes:            routes,
			TokenIn:           tokenIn,
			TokenOutMinAmount: tokenOutMin,
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		return osmo_simulation.GenAndDeliverTxWithRandFees(
			r, app, txGen, &msg, sdk.Coins{tokenIn}, ctx, simAccount, ak, bk, types.ModuleName)
	}
}

func RandomExactAmountInRoute(ctx sdk.Context, r *rand.Rand, k keeper.Keeper, tokenIn sdk.Coin) (res []types.SwapAmountInRoute, tokenOut sdk.Coin) {
	routeLen := r.Intn(1) + 1

	pools, err := k.GetPools(ctx)
	if err != nil {
		panic(err)
	}
	if len(pools) == 0 {
		return
	}

	res = make([]types.SwapAmountInRoute, routeLen)
	for i := range res {
		for {
			pool := pools[r.Intn(len(pools))]
			inAsset, err := pool.GetPoolAsset(tokenIn.Denom)
			if err != nil {
				continue
			}
			if inAsset.Token.Amount.LT(tokenIn.Amount) {
				continue
			}
			for _, asset := range pool.GetAllPoolAssets() {
				if asset.Token.Denom != tokenIn.Denom {
					res[i] = types.SwapAmountInRoute{
						PoolId:        pool.GetId(),
						TokenOutDenom: asset.Token.Denom,
					}
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
			}
			break
		}
	}

	tokenOut = tokenIn
	return
}
