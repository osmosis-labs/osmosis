package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

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
	pool, err := k.GetPoolAndPoke(ctx, poolID)
	if err != nil {
		return sdk.Dec{}, err
	}

	return pool.SpotPrice(ctx, baseAssetDenom, quoteAssetDenom)
}

func validateCreatePoolMsg(ctx sdk.Context, msg types.CreatePoolMsg) error {
	err := msg.Validate(ctx)
	if err != nil {
		return err
	}

	initialPoolLiquidity := msg.InitialLiquidity()
	numAssets := initialPoolLiquidity.Len()
	if numAssets < types.MinPoolAssets {
		return types.ErrTooFewPoolAssets
	}
	if numAssets > types.MaxPoolAssets {
		return sdkerrors.Wrapf(
			types.ErrTooManyPoolAssets,
			"pool has too many PoolAssets (%d)", numAssets,
		)
	}
	return nil
}

func (k Keeper) validateCreatedPool(
	ctx sdk.Context,
	initialPoolLiquidity sdk.Coins,
	poolId uint64,
	pool types.PoolI,
) error {
	if pool.GetId() != poolId {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool ID.")
	}
	if !pool.GetAddress().Equals(types.NewPoolAddress(poolId)) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool address.")
	}
	// Notably we use the initial pool liquidity at the start of the messages definition
	// just in case CreatePool was mutative.
	if !pool.GetTotalPoolLiquidity(ctx).IsEqual(initialPoolLiquidity) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created, with initial liquidity not equal to what was specified.")
	}
	// This check can be removed later, and replaced with a minimum.
	if !pool.GetTotalShares().Equal(types.InitPoolSharesSupply) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect number of initial shares.")
	}
	acc := k.accountKeeper.GetAccount(ctx, pool.GetAddress())
	if acc != nil {
		return sdkerrors.Wrapf(types.ErrPoolAlreadyExist, "pool %d already exist", poolId)
	}
	return nil
}

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, shares are minted
// and sent to the pool creator. The shares are created using a denomination in
// the form of gamm/pool/{poolID}. In addition, the x/bank metadata is updated
// to reflect the newly created GAMM share denomination.
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
	err := validateCreatePoolMsg(ctx, msg)
	if err != nil {
		return 0, err
	}

	sender := msg.PoolCreator()
	initialPoolLiquidity := msg.InitialLiquidity()

	// send pool creation fee to community pool
	params := k.GetParams(ctx)
	if err := k.distrKeeper.FundCommunityPool(ctx, params.PoolCreationFee, sender); err != nil {
		return 0, err
	}

	poolId := k.GetNextPoolNumberAndIncrement(ctx)
	pool, err := msg.CreatePool(ctx, poolId)
	if err != nil {
		return 0, err
	}

	if err := k.validateCreatedPool(ctx, initialPoolLiquidity, poolId, pool); err != nil {
		return 0, err
	}

	// create and save the pool's module account to the account keeper
	acc := k.accountKeeper.NewAccount(
		ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(pool.GetAddress()),
			pool.GetAddress().String(),
		),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	// send initial liquidity to the pool
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), initialPoolLiquidity)
	if err != nil {
		return 0, err
	}

	// Mint the initial pool shares share token to the sender
	err = k.MintPoolShareToAccount(ctx, pool, sender, pool.GetTotalShares())
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

	if err := k.SetPool(ctx, pool); err != nil {
		return 0, err
	}

	k.hooks.AfterPoolCreated(ctx, sender, pool.GetId())
	k.RecordTotalLiquidityIncrease(ctx, initialPoolLiquidity)

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
	// all pools handled within this method are pointer references, `JoinPool` directly updates the pools
	pool, err := k.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}

	// we do an abstract calculation on the lp liquidity coins needed to have
	// the designated amount of given shares of the pool without performing swap
	neededLpLiquidity, err := getMaximalNoSwapLPAmount(ctx, pool, shareOutAmount)
	if err != nil {
		return err
	}

	// check that needed lp liquidity does not exceed the given `tokenInMaxs` parameter. Return error if so.
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

// getMaximalNoSwapLPAmount returns the coins(lp liquidity) needed to get the specified amount of share of the pool.
// Steps to getting the needed lp liquidity coins needed for the share of the pools are
// 		1. calculate how much percent of the pool does given share account for(# of input shares / # of current total shares)
// 		2. since we know how much % of the pool we want, iterate through all pool liquidity to calculate how much coins we need for
// 	  	   each pool asset.
func getMaximalNoSwapLPAmount(ctx sdk.Context, pool types.PoolI, shareOutAmount sdk.Int) (neededLpLiquidity sdk.Coins, err error) {
	totalSharesAmount := pool.GetTotalShares()
	// shareRatio is the desired number of shares, divided by the total number of
	// shares currently in the pool. It is intended to be used in scenarios where you want
	shareRatio := shareOutAmount.ToDec().QuoInt(totalSharesAmount)
	if shareRatio.LTE(sdk.ZeroDec()) {
		return sdk.Coins{}, sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative")
	}

	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)
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

// JoinSwapExactAmountIn is an LP transaction, that will LP all of the provided
// tokensIn coins. The underlying pool is responsible for swapping any non-even
// LP proportions to the correct ratios. An error is returned if the amount of
// LP shares obtained at the end is less than shareOutMinAmount. Otherwise, we
// return the total amount of shares outgoing from joining the pool.
func (k Keeper) JoinSwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokensIn sdk.Coins,
	shareOutMinAmount sdk.Int,
) (sdk.Int, error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesOut, err := pool.JoinPool(ctx, tokensIn, pool.GetSwapFee(ctx))
	switch {
	case err != nil:
		return sdk.ZeroInt(), err

	case sharesOut.LT(shareOutMinAmount):
		return sdk.ZeroInt(), sdkerrors.Wrapf(
			types.ErrLimitMinAmount,
			"too much slippage; needed a minimum of %s shares to pass, got %s",
			shareOutMinAmount, sharesOut,
		)

	case sharesOut.LTE(sdk.ZeroInt()):
		return sdk.ZeroInt(), sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share amount is zero or negative")
	}

	if err := k.applyJoinPoolStateChange(ctx, pool, sender, sharesOut, tokensIn); err != nil {
		return sdk.ZeroInt(), err
	}

	return sharesOut, nil
}

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

	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return sdk.Int{}, fmt.Errorf("pool with id %d does not support this kind of join", poolId)
	}

	tokenInAmount, err = extendedPool.CalcTokenInShareAmountOut(ctx, tokenInDenom, shareOutAmount, pool.GetSwapFee(ctx))
	if err != nil {
		return sdk.Int{}, err
	}

	if tokenInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrLimitMaxAmount, "%d resulted tokens is larger than the max amount of %d", tokenInAmount.Int64(), tokenInMaxAmount.Int64())
	}

	tokenIn := sdk.NewCoins(sdk.NewCoin(tokenInDenom, tokenInAmount))
	extendedPool.IncreaseLiquidity(shareOutAmount, tokenIn)

	err = k.applyJoinPoolStateChange(ctx, pool, sender, shareOutAmount, tokenIn)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	return tokenInAmount, nil
}

func (k Keeper) ExitPool(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	shareInAmount sdk.Int,
	tokenOutMins sdk.Coins,
) (exitCoins sdk.Coins, err error) {
	pool, err := k.GetPoolAndPoke(ctx, poolId)
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

func (k Keeper) ExitSwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	pool, err := k.getPoolForSwap(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	extendedPool, ok := pool.(types.PoolAmountOutExtension)
	if !ok {
		return sdk.Int{}, fmt.Errorf("pool with id %d does not support this kind of exit", poolId)
	}

	shareInAmount, err = extendedPool.ExitSwapExactAmountOut(ctx, tokenOut, shareInMaxAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	if err := k.applyExitPoolStateChange(ctx, pool, sender, shareInAmount, sdk.Coins{tokenOut}); err != nil {
		return sdk.Int{}, err
	}

	return shareInAmount, nil
}
