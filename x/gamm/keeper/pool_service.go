package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/osmosis/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func (k Keeper) CreateBalancerPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	BalancerPoolParams balancer.BalancerPoolParams,
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

func (k Keeper) updatePoolForJoin(
	ctx sdk.Context,
	pool types.PoolI,
	sender sdk.AccAddress,
	coins sdk.Coins,
	shareOutAmount sdk.Int,
) error {
	err := pool.AddPoolAssetBalance(coins...)
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

	coins, err := types.CalcJoin(
		pool.GetAllPoolAssets(),
		pool.GetTotalShares().Amount,
		shareOutAmount,
	)
	if err != nil {
		return err
	}

	// Assume that the tokenInMaxAmounts is validated.
	tokenInMaxMap := make(map[string]sdk.Int)
	for _, max := range tokenInMaxs {
		tokenInMaxMap[max.Denom] = max.Amount
	}

	for _, tokenIn := range coins {
		if tokenInMaxAmount, ok := tokenInMaxMap[tokenIn.Denom]; ok && tokenIn.Amount.GT(tokenInMaxAmount) {
			return sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", tokenIn.Denom)
		}
	}

	return k.updatePoolForJoin(ctx, pool, sender, coins, shareOutAmount)
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

	shareOutAmount = types.CalcPoolOutGivenSingleIn(
		pool.Swap(),
		PoolAsset.Normalize(pool.GetTotalWeight()),
		pool.GetTotalShares().Amount,
		tokenIn.Amount,
		pool.GetPoolSwapFee(),
	).TruncateInt()

	if shareOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share amount is zero or negative")
	}

	if shareOutAmount.LT(shareOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", PoolAsset.Token.Denom)
	}

	return shareOutAmount, k.updatePoolForJoin(ctx, pool, sender, sdk.Coins{tokenIn}, shareOutAmount)
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

	tokenInAmount = types.CalcSingleInGivenPoolOut(
		pool.Swap(),
		PoolAsset.Normalize(pool.GetTotalWeight()),
		pool.GetTotalShares().Amount,
		shareOutAmount,
		pool.GetPoolSwapFee(),
	).TruncateInt()

	if tokenInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", PoolAsset.Token.Denom)
	}

	tokenIn := sdk.NewCoin(tokenInDenom, tokenInAmount)

	return tokenInAmount, k.updatePoolForJoin(ctx, pool, sender, sdk.Coins{tokenIn}, shareOutAmount)
}

func (k Keeper) updatePoolForExit(
	ctx sdk.Context,
	pool types.PoolI,
	sender sdk.AccAddress,
	shareInAmount sdk.Int,
	coins sdk.Coins,
) error {
	err := pool.SubPoolAssetBalance(coins...)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoins(ctx, pool.GetAddress(), sender, coins)
	if err != nil {
		return err
	}

	exitFee := pool.GetPoolExitFee().MulInt(shareInAmount).TruncateInt()
	shareInAmountAfterExitFee := shareInAmount.Sub(exitFee)
	// Remove the exit fee shares from the pool.
	// This distributes the exit fee liquidity to every other LP remaining in the pool.

	// TODO: `balancer` contract sends the exit fee to the `factory` contract.
	//       But, it is unclear that how the exit fees in the `factory` contract are handled.
	//       And, it seems to be not good way to send the exit fee to the pool,
	//       because the pool doesn't have the PoolAsset about exit fee.
	//       So, temporarily, just burn the exit fee.

	if exitFee.IsPositive() {
		err = k.BurnPoolShareFromAccount(ctx, pool, sender, exitFee)
		if err != nil {
			return err
		}
	}

	err = k.BurnPoolShareFromAccount(ctx, pool, sender, shareInAmountAfterExitFee)
	if err != nil {
		return err
	}

	err = k.SetPool(ctx, pool)
	if err != nil {
		return err
	}

	k.createRemoveLiquidityEvent(ctx, sender, pool.GetId(), coins)
	k.hooks.AfterExitPool(ctx, sender, pool.GetId(), shareInAmount, coins)
	k.RecordTotalLiquidityDecrease(ctx, coins)

	return nil

}

func (k Keeper) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareInAmount sdk.Int,
	tokenOutMins sdk.Coins,
) (err error) {
	pool, err := k.GetPool(ctx, poolId)
	if err != nil {
		return err
	}

	coins, err := types.CalcExit(
		pool.GetAllPoolAssets(),
		pool.GetTotalShares().Amount,
		shareInAmount,
		pool.GetPoolExitFee(),
	)
	if err != nil {
		return err
	}

	// Assume that the tokenInMaxAmounts is validated.
	tokenOutMinMap := make(map[string]sdk.Int)
	for _, min := range tokenOutMins {
		tokenOutMinMap[min.Denom] = min.Amount
	}

	for _, tokenOut := range coins {
		// Check if a minimum token amount is specified for this token,
		// and if so ensure that the minimum is less than the amount returned.
		if tokenOutMinAmount, ok := tokenOutMinMap[tokenOut.Denom]; ok && tokenOut.Amount.LT(tokenOutMinAmount) {
			return sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", tokenOut.Denom)
		}
	}

	return k.updatePoolForExit(ctx, pool, sender, shareInAmount, coins)
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

	tokenOutAmount = types.CalcSingleOutGivenPoolIn(
		pool.Swap(),
		PoolAsset.Normalize(pool.GetTotalWeight()),
		pool.GetTotalShares().Amount,
		shareInAmount,
		pool.GetPoolSwapFee(),
		pool.GetPoolExitFee(),
	).TruncateInt()

	if tokenOutAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if tokenOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMinAmount, "%s token is lesser than min amount", PoolAsset.Token.Denom)
	}

	tokenOut := sdk.NewCoin(tokenOutDenom, tokenOutAmount)
	return tokenOutAmount, k.updatePoolForExit(ctx, pool, sender, shareInAmount, sdk.Coins{tokenOut})
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

	shareInAmount = types.CalcPoolInGivenSingleOut(
		pool.Swap(),
		PoolAsset.Normalize(pool.GetTotalWeight()),
		pool.GetTotalShares().Amount,
		tokenOut.Amount,
		pool.GetPoolSwapFee(),
		pool.GetPoolExitFee(),
	).TruncateInt()

	if shareInAmount.LTE(sdk.ZeroInt()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "token amount is zero or negative")
	}

	if shareInAmount.GT(shareInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%s token is larger than max amount", PoolAsset.Token.Denom)
	}

	return shareInAmount, k.updatePoolForExit(ctx, pool, sender, shareInAmount, sdk.Coins{tokenOut})
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
