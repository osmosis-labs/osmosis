package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// CalculateSpotPrice returns the spot price of the quote asset in terms of the base asset,
// using the specified pool.
// E.g. if pool 1 traded 2 atom for 3 osmo, the quote asset was atom, and the base asset was osmo,
// this would return 1.5. (Meaning that 1 atom costs 1.5 osmo)
func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolID uint64,
	baseAssetDenom string,
	quoteAssetDenom string,
) (sdk.Dec, error) {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return sdk.Dec{}, err
	}
	return pool.SpotPrice(ctx, baseAssetDenom, quoteAssetDenom)
}

func (k Keeper) CreateBalancerPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	BalancerPoolParams balancer.PoolParams,
	poolAssets []balancer.PoolAsset,
	futurePoolGovernor string,
) (uint64, error) {
	if len(poolAssets) < types.MinPoolAssets {
		return 0, types.ErrTooFewPoolAssets
	}
	// TODO: Add the limit of binding token to the pool params?
	if len(poolAssets) > types.MaxPoolAssets {
		return 0, sdkerrors.Wrapf(
			types.ErrTooManyPoolAssets,
			"pool has too many PoolAssets (%d)", len(poolAssets),
		)
	}

	// send pool creation fee to community pool
	params := k.GetParams(ctx)
	err := k.distrKeeper.FundCommunityPool(ctx, params.PoolCreationFee, sender)
	if err != nil {
		return 0, err
	}

	pool, err := k.newBalancerPool(ctx, BalancerPoolParams, poolAssets, futurePoolGovernor)
	if err != nil {
		return 0, err
	}

	// Transfer the PoolAssets tokens to the pool's module account from the user account.
	var coins sdk.Coins
	for _, asset := range poolAssets {
		coins = append(coins, asset.Token)
	}
	if coins == nil {
		return 0, types.ErrTooFewPoolAssets
	}

	coins = coins.Sort()
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), coins)
	if err != nil {
		return 0, err
	}

	// Mint the initial 100.000000000000000000 share token to the sender
	// TODO: read number from what pool says is the number of shares
	err = k.MintPoolShareToAccount(ctx, pool, sender, types.InitPoolSharesSupply)
	if err != nil {
		return 0, err
	}

	// Finally, add the share token's meta data to the bank keeper.
	poolShareBaseDenom := types.GetPoolShareDenom(pool.GetId())
	poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", pool.GetId())
	k.bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: fmt.Sprintf("The share token of the gamm pool %d", pool.GetId()),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    poolShareBaseDenom,
				Exponent: 0,
				Aliases: []string{
					"attopoolshare",
				},
			},
			{
				Denom:    poolShareDisplayDenom,
				Exponent: types.OneShareExponent,
				Aliases:  nil,
			},
		},
		Base:    poolShareBaseDenom,
		Display: poolShareDisplayDenom,
	})

	err = k.SetPool(ctx, pool)
	if err != nil {
		return 0, err
	}

	k.hooks.AfterPoolCreated(ctx, sender, pool.GetId())
	k.RecordTotalLiquidityIncrease(ctx, coins)

	return pool.GetId(), nil
}

// JoinPoolNoSwap aims to LP exactly enough to pool #{poolId} to get shareOutAmount number of LP shares.
// If the required tokens is greater than tokenInMaxs, returns an error & the message reverts.
// Leftover tokens that weren't LP'd (due to being at inexact ratios) remain in the sender account.
//
// JoinPoolNoSwap determines the maximum amount that can be LP'd without any swap,
// by looking at the ratio of the total LP'd assets. (e.g. 2 osmo : 1 atom)
// It then finds the maximal amount that can be LP'd.
func (k Keeper) JoinPoolNoSwap(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareOutAmount sdk.Int,
	tokenInMaxs sdk.Coins,
) (err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return err
	}

	neededLpLiquidity, err := getMaximalNoSwapLPAmount(ctx, pool, shareOutAmount)
	if err != nil {
		return err
	}

	// if neededLPLiquidity >= tokenInMaxs, return err
	// if tokenInMaxs == 0, don't do this check.
	if tokenInMaxs.Len() != 0 {
		if !(neededLpLiquidity.DenomsSubsetOf(tokenInMaxs) && tokenInMaxs.IsAllGTE(neededLpLiquidity)) {
			return sdkerrors.Wrapf(types.ErrLimitMaxAmount, "TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap,"+
				" upperbound: %v, needed %v", tokenInMaxs, neededLpLiquidity)
		}
	}

	sharesOut, err := pool.JoinPool(ctx, neededLpLiquidity, pool.GetSwapFee(ctx))
	if err != nil {
		return err
	}
	// sanity check, don't return error as not worth halting the LP. We know its not too much.
	if sharesOut.LT(shareOutAmount) {
		ctx.Logger().Error(fmt.Sprintf("Expected to JoinPoolNoSwap >= %s shares, actually did %s shares",
			shareOutAmount, sharesOut))
	}

	err = k.applyJoinPoolStateChange(ctx, pool, sender, sharesOut, neededLpLiquidity)
	return err
}

func getMaximalNoSwapLPAmount(ctx sdk.Context, pool types.PoolI, shareOutAmount sdk.Int) (neededLpLiquidity sdk.Coins, err error) {
	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	// (tokens per share) * number of shares out = # tokens * (# shares out / cur total shares)
	shareRatio := shareOutAmount.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	poolLiquidity := pool.GetTotalLpBalances(ctx)
	neededLpLiquidity = sdk.Coins{}

	for _, coin := range poolLiquidity {
		// (coin.Amt * shareRatio).Ceil()
		neededAmt := coin.Amount.ToDec().Mul(shareRatio).Ceil().RoundInt()
		if neededAmt.LTE(sdk.ZeroInt()) {
			return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "Too few shares out wanted")
		}
		neededCoin := sdk.Coin{Denom: coin.Denom, Amount: neededAmt}
		neededLpLiquidity = neededLpLiquidity.Add(neededCoin)
	}
	return neededLpLiquidity, nil
}

// JoinSwapExactAmountIn is an LP transaction, that will LP all of the provided tokensIn coins.
// The underlying pool is responsible for swapping any non-even LP proportions to the correct ratios.
// If the amount of LP shares obtained at the end is less than shareOutMinAmount,
// then return an error and revert the message.
func (k Keeper) JoinSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokensIn sdk.Coins,
	shareOutMinAmount sdk.Int,
) (shareOutAmount sdk.Int, err error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesOut, err := pool.JoinPool(ctx, tokensIn, pool.GetSwapFee(ctx))
	if err != nil {
		return sdk.ZeroInt(), err
	} else if sharesOut.LT(shareOutMinAmount) {
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrLimitMinAmount,
			"Too much slippage, needed a minimum of %s shares to pass, got %s",
			shareOutMinAmount, sharesOut)
	} else if sharesOut.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share amount is zero or negative")
	}

	err = k.applyJoinPoolStateChange(ctx, pool, sender, sharesOut, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	return shareOutAmount, nil
}

//nolint:deadcode,govet // looks like we have known dead code beneath "panic"
func (k Keeper) JoinSwapShareAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	shareOutAmount sdk.Int,
	tokenInMaxAmount sdk.Int,
) (tokenInAmount sdk.Int, err error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	panic("implement") // I moved this past return, it caused everything beneath it to be dead code

	tokenInAmount = sdk.ZeroInt()

	// normalizedWeight := PoolAsset.Weight.ToDec().Quo(pool.GetTotalWeight().ToDec())
	// tokenInAmount = calcSingleInGivenPoolOut(
	// 	PoolAsset.Token.Amount.ToDec(),
	// 	normalizedWeight,
	// 	pool.GetTotalShares().Amount.ToDec(),
	// 	shareOutAmount.ToDec(),
	// 	pool.GetPoolSwapFee(ctx),
	// ).TruncateInt()

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", tokenInDenom)
	}

	tokenIn := sdk.Coins{sdk.NewCoin(tokenInDenom, tokenInAmount)}
	err = k.applyJoinPoolStateChange(ctx, pool, sender, shareOutAmount, tokenIn)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	return shareOutAmount, nil
}

func (k Keeper) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareInAmount sdk.Int,
	tokenOutMins sdk.Coins,
) (exitCoins sdk.Coins, err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	totalSharesAmount := pool.GetTotalShares()
	if shareInAmount.GTE(totalSharesAmount) || shareInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}
	exitFee := pool.GetExitFee(ctx)
	exitCoins, err = pool.ExitPool(ctx, shareInAmount, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}
	if !tokenOutMins.DenomsSubsetOf(exitCoins) || tokenOutMins.IsAnyGT(exitCoins) {
		return sdk.Coins{}, sdkerrors.Wrapf(types.ErrLimitMinAmount,
			"Exit pool returned %s , minimum tokens out specified as %s",
			exitCoins, tokenOutMins)
	}

	err = k.applyExitPoolStateChange(ctx, pool, sender, shareInAmount, exitCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	return exitCoins, nil
}

// ExitSwapShareAmountIn is an Exit Pool transaction, that will exit all of the provided LP shares,
// and then swap it all against the pool into tokenOutDenom.
// If the amount of tokens gotten out after the swap is less than tokenOutMinAmount, return an error.
func (k Keeper) ExitSwapShareAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOutDenom string,
	shareInAmount sdk.Int,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	exitCoins, err := k.ExitPool(ctx, sender, poolId, shareInAmount, sdk.Coins{})
	if err != nil {
		return sdk.Int{}, err
	}
	tokenOutAmount = exitCoins.AmountOf(tokenOutDenom)
	for _, coin := range exitCoins {
		if coin.Denom == tokenOutDenom {
			continue
		}
		swapOut, err := k.SwapExactAmountIn(ctx, sender, poolId, coin, tokenOutDenom, sdk.ZeroInt())
		if err != nil {
			return sdk.Int{}, err
		}
		tokenOutAmount = tokenOutAmount.Add(swapOut)
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount,
			"Provided LP shares yield %s tokens out, wanted a minimum of %s for it to work",
			tokenOutAmount, tokenOutMinAmount)
	}
	return tokenOutAmount, nil
}

func (k Keeper) ExitSwapExternAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	// Basically what we have to do is:
	// estimate how many LP shares this would take to do.
	// We do so by calculating how much a swap of half of tokenOut to TokenIn would be.
	// Then we calculate how many LP shares that'd be worth.
	// We should have code for that once we implement JoinPoolNoSwap.
	// Then check if the number of shares is LTE to shareInMaxAmount.
	// if so, use the needed number of shares, do exit pool, and the swap.

	panic("To implement later")
}
