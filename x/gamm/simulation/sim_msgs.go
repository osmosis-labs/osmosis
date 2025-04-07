package gammsimulation

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	legacysimulationtype "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	balancertypes "github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var PoolCreationFee = sdk.NewInt64Coin("stake", 10_000_000)

// RandomJoinPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective underlying token denoms, and attempts to execute a join pool transaction
func RandomJoinPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinPool, error) {
	// get random pool
	pool, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}
	poolDenoms := pool.GetTotalPoolLiquidity(ctx).Denoms()

	// get address that has all denoms from the randomly selected pool
	sender, tokenIn, senderExists := sim.SelAddrWithDenoms(ctx, poolDenoms)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", poolDenoms)
	}

	// cap joining pool to the pool liquidity
	tokenIn = tokenIn.Min(pool.GetTotalPoolLiquidity(ctx))

	// TODO: Fix API so this is a one liner, pool.CalcJoinPoolNoSwapShares()
	minShareOutAmt, err := deriveRealMinShareOutAmt(ctx, tokenIn, pool)
	if err != nil {
		return nil, err
	}

	// TODO: Make FuzzTokenSubset API, token_in_maxs := sim.FuzzTokensSubset(sender, poolDenoms)
	// TODO: Add some slippage tolerance
	// TODO: Make MinShareOutAmt fuzz API: minShareOutAmt = sim.FuzzEqualInt(share_out_amount)
	return &types.MsgJoinPool{
		Sender:         sender.Address.String(),
		PoolId:         pool.GetId(),
		ShareOutAmount: minShareOutAmt,
		TokenInMaxs:    tokenIn,
	}, nil
}

// RandomExitPoolMsg pseudo-randomly selects an existing pool ID, attempts to find an account with the
// respective unbonded gamm shares, and attempts to execute an exit pool transaction
func RandomExitPoolMsg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitPool, error) {
	// get random pool
	pool, _, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}
	poolId := pool.GetId()
	gammDenom := types.GetPoolShareDenom(poolId)

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// calculate the minimum number of tokens received from input of gamm shares
	tokenOutMins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgExitPool{
		Sender:        sender.Address.String(),
		PoolId:        poolId,
		ShareInAmount: gammShares.Amount,
		TokenOutMins:  tokenOutMins,
	}, nil
}

// RandomCreatePoolMsg attempts to find an account with two or more distinct denoms and attempts to send a
// create pool message composed of those denoms
func RandomCreateUniV2Msg(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*balancertypes.MsgCreateBalancerPool, error) {
	var poolAssets []balancertypes.PoolAsset
	// find an address with two or more distinct denoms in their wallet
	sender, senderExists := sim.RandomSimAccountWithConstraint(createPoolRestriction(sim, ctx))
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
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	}

	// from the above selected account, determine the token type and respective weight needed to make the pool
	for i := 0; i < len(poolCoins); i++ {
		poolAssets = append(poolAssets, balancertypes.PoolAsset{
			Weight: osmomath.OneInt(),
			Token:  poolCoins[i],
		})
	}

	return &balancertypes.MsgCreateBalancerPool{
		Sender:     sender.Address.String(),
		PoolParams: poolParams,
		PoolAssets: poolAssets,
	}, nil
}

// RandomSwapExactAmountIn utilizes a random pool and swaps and exact amount in for minimum of the secondary pool token
// TODO: Improve this to swap through multiple pools
func RandomSwapExactAmountIn(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgSwapExactAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, coinOut, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountInRoute{{
		PoolId:        pool.GetId(),
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

	// N.B. Calling MsgSwapExactAmountIn executes the swap via the pool manager, which charges the taker fee.
	// We therefore need to remove the taker fee from the amountIn before calling the calc method.
	takerFee, err := k.GetTradingPairTakerFee(ctx, coinIn.Denom, coinOut.Denom)
	if err != nil {
		return nil, err
	}

	amountInAfterSubTakerFee := randomCoinSubset[0].Amount.ToLegacyDec().Mul(osmomath.OneDec().Sub(takerFee))
	tokenInAfterSubTakerFee := sdk.NewCoin(randomCoinSubset[0].Denom, amountInAfterSubTakerFee.TruncateInt())

	tokenOutMin, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(tokenInAfterSubTakerFee), coinOut.Denom, pool.GetSpreadFactor(ctx))
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
func RandomSwapExactAmountOut(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgSwapExactAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, coinOut, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// set the swap route to use this pool
	route := []poolmanagertypes.SwapAmountOutRoute{{
		PoolId:       pool.GetId(),
		TokenInDenom: coinIn.Denom,
	}}

	// find an address that has a balance of the coinIn
	senderAcc, accCoin, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", coinIn.Denom)
	}

	// set the subset of coins to be upper-bound to the minimum between the address and the pool itself
	randomCoinInSubset := sdk.NewCoins(coinIn).Min(sdk.NewCoins(accCoin))

	// utilize CalcOutAmtGivenIn to calculate tokenOut and use tokenOut to calculate tokenInMax
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, randomCoinInSubset, coinOut.Denom, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}
	tokenInMax, err := pool.CalcInAmtGivenOut(ctx, sdk.NewCoins(tokenOut), coinIn.Denom, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	// N.B. Calling MsgSwapExactAmountOut executes the swap via the pool manager, which charges the taker fee.
	// We therefore need to add the taker fee to the amountIn after calling the calc method.
	takerFee, err := k.GetTradingPairTakerFee(ctx, coinIn.Denom, coinOut.Denom)
	if err != nil {
		return nil, err
	}

	amountInAfterAddTakerFee := tokenInMax.Amount.ToLegacyDec().Quo(osmomath.OneDec().Sub(takerFee))
	tokenInMax = sdk.NewCoin(tokenInMax.Denom, amountInAfterAddTakerFee.TruncateInt())

	return &types.MsgSwapExactAmountOut{
		Sender:           senderAcc.Address.String(),
		Routes:           route,
		TokenInMaxAmount: tokenInMax.Amount,
		TokenOut:         tokenOut,
	}, nil
}

// RandomJoinSwapExternAmountIn utilizes a random pool and with a random account does a single asset join with an exact
// amount of an asset for a minimum number of LP shares
func RandomJoinSwapExternAmountIn(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinSwapExternAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}

	// find an address with the coinIn denom and randomly select a subset of the coin
	sender, tokenIn, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", coinIn.Denom)
	}

	// cap joining pool to the pool liquidity
	newTokenIn := sdk.NewCoins(coinIn).Min(sdk.NewCoins(tokenIn))

	// calc shares out with tokenIn
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, newTokenIn, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	return &types.MsgJoinSwapExternAmountIn{
		Sender:            sender.Address.String(),
		PoolId:            pool.GetId(),
		TokenIn:           newTokenIn[0],
		ShareOutMinAmount: minShareOutAmt,
	}, nil
}

// RandomJoinSwapShareAmountOut utilizes a random pool and with a random account and swaps a maximum of a specified token
// for an exact amount of LP shares
func RandomJoinSwapShareAmountOut(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgJoinSwapShareAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, _, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}
	poolId := pool.GetId()

	// find an address with the coinIn denom and randomly select a subset of the coin
	sender, tokenIn, senderExists := sim.SelAddrWithDenom(ctx, coinIn.Denom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denoms %s exists", coinIn.Denom)
	}

	// cap joining pool to the pool liquidity
	newTokenIn := sdk.NewCoins(coinIn).Min(sdk.NewCoins(tokenIn))

	// calc shares out with tokenIn
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, newTokenIn, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	// use PoolAmountOutExtension to calculate correct tokenInMaxAmount
	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return nil, fmt.Errorf("pool with id %d does not support this kind of join", poolId)
	}
	tokenInAmount, err := extendedPool.CalcTokenInShareAmountOut(ctx, tokenIn.Denom, minShareOutAmt, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	accountBalance := sim.BankKeeper().GetBalance(ctx, sender.Address, tokenIn.Denom)
	if accountBalance.Amount.LT(tokenInAmount) {
		return nil, fmt.Errorf("insufficient funds")
	}

	return &types.MsgJoinSwapShareAmountOut{
		Sender:           sender.Address.String(),
		PoolId:           poolId,
		TokenInDenom:     tokenIn.Denom,
		ShareOutAmount:   minShareOutAmt,
		TokenInMaxAmount: tokenInAmount,
	}, nil
}

// RandomExitSwapExternAmountOut utilizes a random pool and with a random account and swaps a maximum number of LP shares
// for an exact amount of one of the token pairs
func RandomExitSwapExternAmountOut(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitSwapExternAmountOut, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, coinOut, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}
	poolId := pool.GetId()
	gammDenom := types.GetPoolShareDenom(poolId)

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// select a random amount of the account's gamm shares
	gammShares.Amount = sim.RandomAmount(gammShares.Amount)

	// calc exitedCoins from gammShares in
	exitedCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	// get amount of coinIn from exitedCoins and calculate how much of tokenOut you should get from that
	exitedCoinsIn := exitedCoins.AmountOf(coinIn.Denom)
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewCoin(coinIn.Denom, exitedCoinsIn)), coinOut.Denom, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to calculate the swap for the entire amount
	// I felt I was doing it correct but it was always off
	// since we are only doing half the swap out, we only use half the share in
	return &types.MsgExitSwapExternAmountOut{
		Sender:           sender.Address.String(),
		PoolId:           poolId,
		TokenOut:         tokenOut,
		ShareInMaxAmount: gammShares.Amount.Quo(osmomath.NewInt(2)),
	}, nil
}

// RandomExitSwapShareAmountIn utilizes a random pool and with a random account and swaps an number of LP shares
// for a minimum amount of one of the token pairs
func RandomExitSwapShareAmountIn(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (*types.MsgExitSwapShareAmountIn, error) {
	// get random pool, randomly select one of the pool denoms to be the coinIn, other is coinOut
	pool, coinIn, coinOut, err := getRandPool(k, sim, ctx)
	if err != nil {
		return nil, err
	}
	poolId := pool.GetId()
	gammDenom := types.GetPoolShareDenom(poolId)

	// select an address that has gamm shares of the selected pool
	sender, gammShares, senderExists := sim.SelAddrWithDenom(ctx, gammDenom)
	if !senderExists {
		return nil, fmt.Errorf("no sender with denom %s exists", gammDenom)
	}

	// select a random amount of the account's gamm shares
	gammShares.Amount = sim.RandomAmount(gammShares.Amount)

	// calc exitedCoins from gammShares in
	exitedCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, gammShares.Amount, pool.GetExitFee(ctx))
	if err != nil {
		return nil, err
	}

	// get amount of coinIn from exitedCoins and calculate how much of tokenOut you should get from that
	exitedCoinsIn := exitedCoins.AmountOf(coinIn.Denom)
	tokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(sdk.NewCoin(coinIn.Denom, exitedCoinsIn)), coinOut.Denom, pool.GetSpreadFactor(ctx))
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to calculate the swap for the entire amount
	// I felt I was doing it correct but it was always off
	// since we are only doing half the swap out, we only use half the share in
	return &types.MsgExitSwapShareAmountIn{
		Sender:            sender.Address.String(),
		PoolId:            pool.GetId(),
		TokenOutDenom:     tokenOut.Denom,
		ShareInAmount:     gammShares.Amount.Quo(osmomath.NewInt(2)),
		TokenOutMinAmount: tokenOut.Amount.Quo(osmomath.NewInt(2)),
	}, nil
}

// TODO: Fix CalcJoinPoolShares API so we don't have to do this
func deriveRealMinShareOutAmt(ctx sdk.Context, tokenIn sdk.Coins, pool types.CFMMPoolI) (osmomath.Int, error) {
	minShareOutAmt, _, err := pool.CalcJoinPoolShares(ctx, tokenIn, pool.GetSpreadFactor(ctx))
	if err != nil {
		return osmomath.Int{}, err
	}
	totalSharesAmount := pool.GetTotalShares()

	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	shareRatio := minShareOutAmt.ToLegacyDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(osmomath.ZeroDec()) {
		return osmomath.Int{}, fmt.Errorf("share ratio is zero or negative")
	}

	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
	neededLpLiquidity := sdk.Coins{}

	for _, coin := range poolLiquidity {
		neededAmt := coin.Amount.ToLegacyDec().Mul(shareRatio).Ceil().RoundInt()
		if neededAmt.LTE(osmomath.ZeroInt()) {
			return osmomath.Int{}, fmt.Errorf("Too few shares out wanted")
		}
		neededCoin := sdk.Coin{Denom: coin.Denom, Amount: neededAmt}
		neededLpLiquidity = neededLpLiquidity.Add(neededCoin)
	}

	if tokenIn.Len() != 0 {
		if !(neededLpLiquidity.DenomsSubsetOf(tokenIn) && tokenIn.IsAllGTE(neededLpLiquidity)) {
			return osmomath.Int{}, fmt.Errorf("TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap,"+
				" upperbound: %v, needed %v", tokenIn, neededLpLiquidity)
		} else if !(tokenIn.DenomsSubsetOf(neededLpLiquidity)) {
			return osmomath.Int{}, fmt.Errorf("TokenInMaxs includes tokens that are not part of the target pool,"+
				" input tokens: %v, pool tokens %v", tokenIn, neededLpLiquidity)
		}
	}

	return minShareOutAmt, nil
}

func createPoolRestriction(sim *simtypes.SimCtx, ctx sdk.Context) simtypes.SimAccountConstraint {
	return func(acc legacysimulationtype.Account) bool {
		accCoins := sim.BankKeeper().SpendableCoins(ctx, acc.Address)
		hasTwoCoins := len(accCoins) >= 2
		hasPoolCreationFee := accCoins.AmountOf(PoolCreationFee.Denom).GT(PoolCreationFee.Amount)
		return hasTwoCoins && hasPoolCreationFee
	}
}

func getRandPool(k keeper.Keeper, sim *simtypes.SimCtx, ctx sdk.Context) (types.CFMMPoolI, sdk.Coin, sdk.Coin, error) {
	// select a pseudo-random pool ID, max bound by the upcoming pool ID
	pools, err := k.GetPoolsAndPoke(ctx)
	if err != nil {
		return nil, sdk.NewCoin("denom", osmomath.ZeroInt()), sdk.NewCoin("denom", osmomath.ZeroInt()), err
	}

	numPools := len(pools)
	if numPools == 0 {
		return nil, sdk.NewCoin("denom", osmomath.ZeroInt()), sdk.NewCoin("denom", osmomath.ZeroInt()), fmt.Errorf("no pools exist")
	}

	rand := sim.GetRand()

	pool := pools[rand.Intn(numPools)]

	poolCoins := pool.GetTotalPoolLiquidity(ctx)

	// TODO: Improve this, don't just assume two asset pools
	// randomly select one of the pool denoms to be the coinIn and one to be the coinOut
	r := sim.GetSeededRand("select random seed")
	index := r.Intn(len(poolCoins) - 1)
	coinIn := poolCoins[index]
	poolCoins = simtypes.RemoveIndex(poolCoins, index)
	coinOut := poolCoins[0]

	return pool, coinIn, coinOut, err
}
