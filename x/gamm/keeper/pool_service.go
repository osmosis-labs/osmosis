package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

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

	totalSharesAmount := pool.GetTotalShares().Amount
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

	totalSharesAmount := pool.GetTotalShares().Amount
	exitFee := pool.GetPoolExitFee().MulInt(shareInAmount).TruncateInt()
	shareInAmountAfterExitFee := shareInAmount.Sub(exitFee)
	shareRatio := shareInAmountAfterExitFee.ToDec().QuoInt(totalSharesAmount)

	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	// Assume that the tokenInMaxAmounts is validated.
	tokenOutMinMap := make(map[string]sdk.Int)
	for _, min := range tokenOutMins {
		tokenOutMinMap[min.Denom] = min.Amount
	}

	PoolAssets := pool.GetAllPoolAssets()
	newPoolCoins := make([]sdk.Coin, 0, len(PoolAssets))
	// Transfer the PoolAssets tokens to the user account from the pool's module account.
	var coins sdk.Coins
	for _, PoolAsset := range PoolAssets {
		tokenOutAmount := shareRatio.MulInt(PoolAsset.Token.Amount).TruncateInt()
		if tokenOutAmount.LTE(sdk.ZeroInt()) {
			return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
		}

		// Check if a minimum token amount is specified for this token,
		// and if so ensure that the minimum is less than the amount returned.
		if tokenOutMinAmount, ok := tokenOutMinMap[PoolAsset.Token.Denom]; ok && tokenOutAmount.LT(tokenOutMinAmount) {
			return sdk.Coins{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", PoolAsset.Token.Denom)
		}

		newPoolCoins = append(newPoolCoins,
			sdk.NewCoin(PoolAsset.Token.Denom, PoolAsset.Token.Amount.Sub(tokenOutAmount)))
		coins = append(coins, sdk.NewCoin(PoolAsset.Token.Denom, tokenOutAmount))
	}

	err = pool.UpdatePoolAssetBalances(newPoolCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, coins)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Remove the exit fee shares from the pool.
	// This distributes the exit fee liquidity to every other LP remaining in the pool.
	if exitFee.IsPositive() {
		err = k.BurnPoolShareFromAccount(ctx, pool, sender, exitFee)
		if err != nil {
			return sdk.Coins{}, err
		}
	}

	err = k.BurnPoolShareFromAccount(ctx, pool, sender, shareInAmountAfterExitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Coins{}, err
	}

	k.createRemoveLiquidityEvent(ctx, sender, pool.GetId(), coins)
	k.hooks.AfterExitPool(ctx, sender, pool.GetId(), shareInAmount, coins)
	k.RecordTotalLiquidityDecrease(ctx, coins)

	return coins, nil
}

func (k Keeper) ExitSwapShareAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOutDenom string,
	shareInAmount sdk.Int,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "exit swap on inactive pool")
	}

	PoolAsset, err := pool.GetPoolAsset(tokenOutDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := PoolAsset.Weight.ToDec().Quo(pool.GetTotalWeight().ToDec())
	tokenOutAmount = calcSingleOutGivenPoolIn(
		PoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		pool.GetTotalShares().Amount.ToDec(),
		shareInAmount.ToDec(),
		pool.GetPoolSwapFee(),
		pool.GetPoolExitFee(),
	).TruncateInt()

	if tokenOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", PoolAsset.Token.Denom)
	}

	PoolAsset.Token.Amount = PoolAsset.Token.Amount.Sub(tokenOutAmount)
	err = pool.UpdatePoolAssetBalance(PoolAsset.Token)
	if err != nil {
		return sdk.Int{}, err
	}

	exitFee := pool.GetPoolExitFee().MulInt(shareInAmount).TruncateInt()
	shareInAmountAfterExitFee := shareInAmount.Sub(exitFee)

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, sdk.Coins{
		sdk.NewCoin(tokenOutDenom, tokenOutAmount),
	})
	if err != nil {
		return sdk.Int{}, err
	}

	// TODO: `balancer` contract sends the exit fee to the `factory` contract.
	//       But, it is unclear that how the exit fees in the `factory` contract are handled.
	//       And, it seems to be not good way to send the exit fee to the pool,
	//       because the pool doesn't have the PoolAsset about exit fee.
	//       So, temporarily, just burn the exit fee.
	if exitFee.IsPositive() {
		err = k.BurnPoolShareFromAccount(ctx, pool, sender, exitFee)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	err = k.BurnPoolShareFromAccount(ctx, pool, sender, shareInAmountAfterExitFee)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Int{}, err
	}

	removedCoins := sdk.Coins{sdk.NewCoin(tokenOutDenom, tokenOutAmount)}
	k.createRemoveLiquidityEvent(ctx, sender, pool.GetId(), removedCoins)
	k.hooks.AfterExitPool(ctx, sender, pool.GetId(), shareInAmount, removedCoins)
	k.RecordTotalLiquidityDecrease(ctx, removedCoins)

	return tokenOutAmount, nil
}

func (k Keeper) ExitSwapExternAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	if !pool.IsActive(ctx.BlockTime()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrPoolLocked, "exit swap on inactive pool")
	}

	PoolAsset, err := pool.GetPoolAsset(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := PoolAsset.Weight.ToDec().Quo(pool.GetTotalWeight().ToDec())
	shareInAmount = calcPoolInGivenSingleOut(
		PoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		pool.GetTotalShares().Amount.ToDec(),
		tokenOut.Amount.ToDec(),
		pool.GetPoolSwapFee(),
		pool.GetPoolExitFee(),
	).TruncateInt()

	if shareInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if shareInAmount.GT(shareInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", PoolAsset.Token.Denom)
	}

	PoolAsset.Token.Amount = PoolAsset.Token.Amount.Sub(tokenOut.Amount)
	err = pool.UpdatePoolAssetBalance(PoolAsset.Token)
	if err != nil {
		return sdk.Int{}, err
	}

	exitFee := pool.GetPoolExitFee().MulInt(shareInAmount).TruncateInt()
	shareInAmountAfterExitFee := shareInAmount.Sub(exitFee)

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, sdk.Coins{
		tokenOut,
	})
	if err != nil {
		return sdk.Int{}, err
	}

	// TODO: `balancer` contract sends the exit fee to the `factory` contract.
	//       But, it is unclear that how the exit fees in the `factory` contract are handled.
	//       And, it seems to be not good way to send the exit fee to the pool,
	//       because the pool doesn't have the PoolAsset about exit fee.
	//       So, temporarily, just burn the exit fee.
	if exitFee.IsPositive() {
		err = k.BurnPoolShareFromAccount(ctx, pool, sender, exitFee)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	err = k.BurnPoolShareFromAccount(ctx, pool, sender, shareInAmountAfterExitFee)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return sdk.Int{}, err
	}

	removedCoins := sdk.Coins{tokenOut}
	k.createRemoveLiquidityEvent(ctx, sender, pool.GetId(), removedCoins)
	k.hooks.AfterExitPool(ctx, sender, pool.GetId(), shareInAmount, removedCoins)
	k.RecordTotalLiquidityDecrease(ctx, removedCoins)

	return shareInAmount, nil
}

func (k Keeper) GetTotalLiquidity(ctx sdk.Context) sdk.Coins {
	coins := sdk.Coins{}
	k.IterateDenomLiquidity(ctx, func(coin sdk.Coin) bool {
		coins = coins.Add(coin)
		return false
	})
	return coins
}

func (k Keeper) SetTotalLiquidity(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		k.SetDenomLiquidity(ctx, coin.Denom, coin.Amount)
	}
}

func (k Keeper) SetDenomLiquidity(ctx sdk.Context, denom string, amount sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz, err := amount.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(types.GetDenomPrefix(denom), bz)
}

func (k Keeper) GetDenomLiquidity(ctx sdk.Context, denom string) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDenomPrefix(denom))
	if bz == nil {
		return sdk.NewInt(0)
	}

	var amount sdk.Int
	if err := amount.Unmarshal(bz); err != nil {
		panic(err)
	}
	return amount
}

func (k Keeper) IterateDenomLiquidity(ctx sdk.Context, cb func(sdk.Coin) bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyTotalLiquidity)

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount sdk.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}

		if cb(sdk.NewCoin(string(iterator.Key()), amount)) {
			break
		}
	}
}

func (k Keeper) RecordTotalLiquidityIncrease(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		amount := k.GetDenomLiquidity(ctx, coin.Denom)
		amount = amount.Add(coin.Amount)
		k.SetDenomLiquidity(ctx, coin.Denom, amount)
	}
}

func (k Keeper) RecordTotalLiquidityDecrease(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		amount := k.GetDenomLiquidity(ctx, coin.Denom)
		amount = amount.Sub(coin.Amount)
		k.SetDenomLiquidity(ctx, coin.Denom, amount)
	}
}
