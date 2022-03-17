package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (k Keeper) CalculateSpotPrice(
	ctx sdk.Context,
	poolID uint64,
	quoteAssetDenom string,
	baseAssetDenom string) (sdk.Dec, error) {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return sdk.Dec{}, err
	}
	return pool.SpotPrice(ctx, quoteAssetDenom, baseAssetDenom)
}

func (k Keeper) CreateBalancerPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	BalancerPoolParams balancer.PoolParams,
	poolAssets []types.PoolAsset,
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

func (k Keeper) JoinPool(
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

	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	// (tokens per share) * number of shares out = # tokens * (# shares out / cur total shares)
	shareRatio := shareOutAmount.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	// Assume that the tokenInMaxAmounts is validated.
	tokenInMaxMap := make(map[string]sdk.Int)
	for _, max := range tokenInMaxs {
		tokenInMaxMap[max.Denom] = max.Amount
	}

	PoolAssets := pool.GetAllPoolAssets()
	newPoolCoins := make([]sdk.Coin, 0, len(PoolAssets))
	// Transfer the PoolAssets tokens to the pool's module account from the user account.
	var coins sdk.Coins
	for _, PoolAsset := range PoolAssets {
		tokenInAmount := shareRatio.MulInt(PoolAsset.Token.Amount).TruncateInt()
		if tokenInAmount.LTE(sdk.ZeroInt()) {
			return sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
		}

		if tokenInMaxAmount, ok := tokenInMaxMap[PoolAsset.Token.Denom]; ok && tokenInAmount.GT(tokenInMaxAmount) {
			return sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", PoolAsset.Token.Denom)
		}

		newPoolCoins = append(newPoolCoins,
			sdk.NewCoin(PoolAsset.Token.Denom, PoolAsset.Token.Amount.Add(tokenInAmount)))
		coins = append(coins, sdk.NewCoin(PoolAsset.Token.Denom, tokenInAmount))
	}

	err = pool.UpdatePoolAssetBalances(newPoolCoins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), coins)
	if err != nil {
		return err
	}

	err = k.MintPoolShareToAccount(ctx, pool, sender, shareOutAmount)
	if err != nil {
		return err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	k.createAddLiquidityEvent(ctx, sender, pool.GetId(), coins)
	k.hooks.AfterJoinPool(ctx, sender, pool.GetId(), coins, shareOutAmount)
	k.RecordTotalLiquidityIncrease(ctx, coins)

	return nil
}

func (k Keeper) JoinSwapExternAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	shareOutMinAmount sdk.Int,
) (shareOutAmount sdk.Int, err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "join swap on inactive pool")
	}

	PoolAsset, err := pool.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := PoolAsset.Weight.ToDec().Quo(pool.GetTotalWeight().ToDec())
	shareOutAmount = calcPoolOutGivenSingleIn(
		PoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		pool.GetTotalShares().Amount.ToDec(),
		tokenIn.Amount.ToDec(),
		pool.GetPoolSwapFee(),
	).TruncateInt()

	if shareOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share amount is zero or negative")
	}

	if shareOutAmount.LT(shareOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", PoolAsset.Token.Denom)
	}

	updatedTokenAmount := PoolAsset.Token.Add(tokenIn)
	err = pool.UpdatePoolAssetBalance(updatedTokenAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{tokenIn})
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.MintPoolShareToAccount(ctx, pool, sender, shareOutAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Int{}, err
	}

	addedCoins := sdk.Coins{tokenIn}
	k.createAddLiquidityEvent(ctx, sender, pool.GetId(), addedCoins)
	k.hooks.AfterJoinPool(ctx, sender, pool.GetId(), addedCoins, shareOutAmount)
	k.RecordTotalLiquidityIncrease(ctx, addedCoins)

	return shareOutAmount, nil
}

func (k Keeper) JoinSwapShareAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenInDenom string,
	shareOutAmount sdk.Int,
	tokenInMaxAmount sdk.Int,
) (tokenInAmount sdk.Int, err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "join swap on inactive pool")
	}

	PoolAsset, err := pool.GetPoolAsset(tokenInDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := PoolAsset.Weight.ToDec().Quo(pool.GetTotalWeight().ToDec())
	tokenInAmount = calcSingleInGivenPoolOut(
		PoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		pool.GetTotalShares().Amount.ToDec(),
		shareOutAmount.ToDec(),
		pool.GetPoolSwapFee(),
	).TruncateInt()

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", PoolAsset.Token.Denom)
	}

	PoolAsset.Token.Amount = PoolAsset.Token.Amount.Add(tokenInAmount)
	err = pool.UpdatePoolAssetBalance(PoolAsset.Token)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.Coins{sdk.NewCoin(tokenInDenom, tokenInAmount)})
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.MintPoolShareToAccount(ctx, pool, sender, shareOutAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Int{}, err
	}

	coinsAdded := sdk.Coins{sdk.NewCoin(tokenInDenom, tokenInAmount)}
	k.createAddLiquidityEvent(ctx, sender, pool.GetId(), coinsAdded)
	k.hooks.AfterJoinPool(ctx, sender, pool.GetId(), coinsAdded, shareOutAmount)
	k.RecordTotalLiquidityIncrease(ctx, coinsAdded)

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
	if shareInAmount.GTE(totalSharesAmount) {
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

	err = k.BurnPoolShareFromAccount(ctx, pool, sender, shareInAmount)
	if err != nil {
		return sdk.Coins{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Coins{}, err
	}

	k.createRemoveLiquidityEvent(ctx, sender, pool.GetId(), exitCoins)
	k.hooks.AfterExitPool(ctx, sender, pool.GetId(), shareInAmount, exitCoins)
	k.RecordTotalLiquidityDecrease(ctx, exitCoins)

	return exitCoins, nil
}

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
