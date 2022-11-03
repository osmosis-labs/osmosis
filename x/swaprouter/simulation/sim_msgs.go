package swaproutersimulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	balancertypes "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// simulationKeeper is a wrapper around swaprouter's keeper which makes it easy to wire new keepers
type simulationKeeper struct {
	keeper swaprouter.Keeper

	gammKeeper types.GammKeeper
}

// PoolCreationFee denotes how much it costs to create a pool.
var PoolCreationFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000)

// RandomSwapExactAmountIn utilizes a random pool and swaps and exact amount in for minimum of the secondary pool token
// TODO: Improve this to swap through multiple pools
func RandomSwapExactAmountIn(k simulationKeeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgSwapExactAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, coinOut, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []types.SwapAmountInRoute{{
		PoolId:        pool_id,
		TokenOutDenom: coinOut.Denom,
	}}

	// find an address that has a balance of the coinIn
	sender, accCoinIn, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", coinIn.Denom)
	}

	// select a random amount that is upper-bound by the address's balance of coinIn
	randomCoinSubset := sim.RandSubsetCoins(sdk.NewCoins(sdk.NewCoin(accCoinIn.Denom, accCoinIn.Amount)))

	// calculate the minimum number of tokens received from input of tokenIn
	tokenOutMin, err := pool.CalcOutAmtGivenIn(ctx, randomCoinSubset, coinOut.Denom, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapExactAmountIn{
		Sender:            sender.Address.String(),
		Routes:            route,
		TokenIn:           randomCoinSubset[0],
		TokenOutMinAmount: tokenOutMin.Amount,
	}, nil
}

// RandomSwapExactAmountOut utilizes a random pool and swaps a max amount amount in for an exact amount of the secondary pool token
// TODO: Improve this to swap through multiple pools
func RandomSwapExactAmountOut(k simulationKeeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgSwapExactAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool_id, pool, coinIn, coinOut, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []types.SwapAmountOutRoute{{
		PoolId:       pool_id,
		TokenInDenom: coinIn.Denom,
	}}

	// find an address that has a balance of the coinIn
	senderAcc, accCoin, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", coinIn.Denom)
	}

	// set the subset of coins to be upper-bound to the minimum between the address and the pool itself
	randomCoinInSubset := osmoutils.MinCoins(sdk.NewCoins(coinIn), sdk.NewCoins(accCoin))

	// utilize CalcOutAmtGivenIn to calculate tokenOut and use tokenOut to calculate tokenInMax
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, randomCoinInSubset, coinOut.Denom, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}
	tokenInMax, err := pool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), coinIn.Denom, pool.GetSwapFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapExactAmountOut{
		Sender:           senderAcc.Address.String(),
		Routes:           route,
		TokenInMaxAmount: tokenInMax.Amount,
		TokenOut:         tokenOut,
	}, nil
}

// RandomCreatePoolMsg attempts to find an account with two or more distinct denoms and attempts to send a
// create pool message composed of those denoms
func RandomCreateUniV2Msg(k swaprouter.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*balancertypes.MsgCreateBalancerPool, error) {
	var poolAssets []balancertypes.PoolAsset
	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(k, sim, ctx))
	if !senderExists {
		return nil, errors.New("no sender with two different denoms & pool creation fee exists")
	}
	poolCoins, ok := sim.GetRandSubsetOfKDenoms(ctx, sender, 2)
	if !ok {
		return nil, fmt.Errorf("provided sender with requested number of denoms does not exist")
	}
	if poolCoins.Add(PoolCreationFee).IsAnyGT(sim.BankKeeper().SpendableCoins(ctx, sender.Address)) {
		return nil, errors.New("chose an account / creation amount that didn't pass fee bar")
	}

	// TODO: pseudo-randomly generate swap and exit fees
	poolParams := &balancertypes.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}

	// from the above selected account, determine the token type and respective weight needed to make the pool
	for i := 0; i < len(poolCoins); i++ {
		poolAssets = append(poolAssets, balancertypes.PoolAsset{
			Weight: sdk.OneInt(),
			Token:  poolCoins[i],
		})
	}

	return &balancertypes.MsgCreateBalancerPool{
		Sender:     sender.Address.String(),
		PoolParams: poolParams,
		PoolAssets: poolAssets,
	}, nil
}

func createPoolRestriction(k swaprouter.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf("stake").GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}

func getRandPool(k simulationKeeper, sim *simtypes.SimCtx, ctx sdk.Context) (uint64, gammtypes.TraditionalAmmInterface, sdk.Coin, sdk.Coin, []string, string, error) {
	// select a pseudo-random pool ID, max bound by the upcoming pool ID
	pool_id := simtypes.RandLTBound(sim, k.keeper.GetNextPoolId(ctx))
	pool, err := k.gammKeeper.GetPoolAndPoke(ctx, pool_id)
	if err != nil {
		return 0, nil, sdk.NewCoin("denom", sdk.ZeroInt()), sdk.NewCoin("denom", sdk.ZeroInt()), []string{}, "", err
	}
	poolCoins := pool.GetTotalPoolLiquidity(ctx)

	// TODO: Improve this, don't just assume two asset pools
	// randomly select one of the pool denoms to be the coinIn and one to be the coinOut
	r := sim.GetSeededRand("select random seed")
	index := r.Intn(len(poolCoins) - 1)
	coinIn := poolCoins[index]
	poolCoins = simtypes.RemoveIndex(poolCoins, index)
	coinOut := poolCoins[0]
	poolDenoms := osmoutils.CoinsDenoms(pool.GetTotalPoolLiquidity(ctx))
	gammDenom := gammtypes.GetPoolShareDenom(pool_id)
	return pool_id, pool, coinIn, coinOut, poolDenoms, gammDenom, err
}
