package poolmanager

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v17/app/params"

	"github.com/osmosis-labs/osmosis/osmoutils"
	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v17/x/txfees/types"
)

// 1 << 256 - 1 where 256 is the max bit length defined for sdk.Int
var intMaxValue = sdk.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))

// RouteExactAmountIn processes a swap along the given route using the swap function
// corresponding to poolID's pool type. It takes in the input denom and amount for
// the initial swap against the first pool and chains the output as the input for the
// next routed pool until the last pool is reached.
// Transaction succeeds if final amount out is greater than tokenOutMinAmount defined
// and no errors are encountered along the way.
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	poolManagerParams := k.GetParams(ctx)

	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	// Iterate through the route and execute a series of swaps through each pool.
	for i, routeStep := range route {
		// To prevent the multihop swap from being interrupted prematurely, we keep
		// the minimum expected output at a very low number until the last pool
		_outMinAmount := sdk.NewInt(1)
		if len(route)-1 == i {
			_outMinAmount = tokenOutMinAmount
		}

		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// Check if pool has swaps enabled.
		if !pool.IsActive(ctx) {
			return sdk.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)
		takerFee := pool.GetTakerFee(ctx)
		totalFees := spreadFactor.Add(takerFee)

		tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, routeStep.TokenOutDenom, _outMinAmount, totalFees)
		if err != nil {
			return sdk.Int{}, err
		}

		err = k.extractTakerFeeToFeePool(ctx, pool, tokenIn, takerFee, poolManagerParams)
		if err != nil {
			return sdk.Int{}, err
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, nil
}

// SplitRouteExactAmountIn routes the swap across multiple multihop paths
// to get the desired token out. This is useful for achieving the most optimal execution. However, note that the responsibility
// of determining the optimal split is left to the client. This method simply route the swap across the given route.
// The route must end with the same token out and begin with the same token in.
//
// It performs the price impact protection check on the combination of tokens out from all multihop paths. The given tokenOutMinAmount
// is used for comparison.
//
// Returns error if:
//   - route are empty
//   - route contain duplicate multihop paths
//   - last token out denom is not the same for all multihop paths in routeStep
//   - one of the multihop swaps fails for internal reasons
//   - final token out computed is not positive
//   - final token out computed is smaller than tokenOutMinAmount
func (k Keeper) SplitRouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInSplitRoute,
	tokenInDenom string,
	tokenOutMinAmount sdk.Int,
) (sdk.Int, error) {
	if err := types.ValidateSwapAmountInSplitRoute(routes); err != nil {
		return sdk.Int{}, err
	}

	var (
		// We start the multihop min amount as zero because we want
		// to perform a price impact protection check on the combination of tokens out
		// from all multihop paths.
		multihopStartTokenOutMinAmount = sdk.ZeroInt()
		totalOutAmount                 = sdk.ZeroInt()
	)

	for _, multihopRoute := range routes {
		tokenOutAmount, err := k.RouteExactAmountIn(
			ctx,
			sender,
			types.SwapAmountInRoutes(multihopRoute.Pools),
			sdk.NewCoin(tokenInDenom, multihopRoute.TokenInAmount),
			multihopStartTokenOutMinAmount)
		if err != nil {
			return sdk.Int{}, err
		}

		totalOutAmount = totalOutAmount.Add(tokenOutAmount)
	}

	if !totalOutAmount.IsPositive() {
		return sdk.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: true, Amount: totalOutAmount}
	}

	if totalOutAmount.LT(tokenOutMinAmount) {
		return sdk.Int{}, types.PriceImpactProtectionExactInError{Actual: totalOutAmount, MinAmount: tokenOutMinAmount}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSplitRouteSwapExactAmountIn,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalOutAmount.String()),
		),
	})

	return totalOutAmount, nil
}

// SwapExactAmountIn is an API for swapping an exact amount of tokens
// as input to a pool to get a minimum amount of the desired token out.
// The method succeeds when tokenOutAmount is greater than tokenOutMinAmount defined.
// Errors otherwise. Also, errors if the pool id is invalid, if tokens do not belong to the pool with given
// id or if sender does not have the swapped-in tokenIn.
func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return sdk.Int{}, poolErr
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	poolManagerParams := k.GetParams(ctx)

	spreadFactor := pool.GetSpreadFactor(ctx)
	takerFee := pool.GetTakerFee(ctx)
	totalFees := spreadFactor.Add(takerFee)

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, totalFees)
	if err != nil {
		return sdk.Int{}, err
	}

	err = k.extractTakerFeeToFeePool(ctx, pool, tokenIn, takerFee, poolManagerParams)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) SwapExactAmountInNoTakerFee(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolId uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
) (tokenOutAmount sdk.Int, err error) {
	// Get the pool-specific module implementation to ensure that
	// swaps are routed to the pool type corresponding to pool ID's pool.
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Int{}, err
	}

	// Get pool as a general pool type. Note that the underlying function used
	// still varies with the pool type.
	pool, poolErr := swapModule.GetPool(ctx, poolId)
	if poolErr != nil {
		return sdk.Int{}, poolErr
	}

	// Check if pool has swaps enabled.
	if !pool.IsActive(ctx) {
		return sdk.Int{}, fmt.Errorf("pool %d is not active", pool.GetId())
	}

	spreadFactor := pool.GetSpreadFactor(ctx)

	// routeStep to the pool-specific SwapExactAmountIn implementation.
	tokenOutAmount, err = swapModule.SwapExactAmountIn(ctx, sender, pool, tokenIn, tokenOutDenom, tokenOutMinAmount, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenOutAmount, nil
}

func (k Keeper) MultihopEstimateOutGivenExactAmountIn(
	ctx sdk.Context,
	route []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
) (tokenOutAmount sdk.Int, err error) {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			tokenOutAmount = sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateOutGivenExactAmountIn failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountInRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	for _, routeStep := range route {
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		// Execute the expected swap on the current routed pool
		poolI, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)
		takerFee := poolI.GetTakerFee(ctx)
		totalFees := spreadFactor.Add(takerFee)

		tokenOut, err := swapModule.CalcOutAmtGivenIn(ctx, poolI, tokenIn, routeStep.TokenOutDenom, totalFees)
		if err != nil {
			return sdk.Int{}, err
		}

		tokenOutAmount = tokenOut.Amount
		if !tokenOutAmount.IsPositive() {
			return sdk.Int{}, errors.New("token amount must be positive")
		}

		// Chain output of current pool as the input for the next routed pool
		tokenIn = sdk.NewCoin(routeStep.TokenOutDenom, tokenOutAmount)
	}
	return tokenOutAmount, err
}

// RouteExactAmountOut processes a swap along the given route using the swap function corresponding
// to poolID's pool type. This function is responsible for computing the optimal output amount
// for a given input amount when swapping tokens, taking into account the current price of the
// tokens in the pool and any slippage.
// Transaction succeeds if the calculated tokenInAmount of the first pool is less than the defined
// tokenInMaxAmount defined.
func (k Keeper) RouteExactAmountOut(ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountOutRoute,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	poolManagerParams := k.GetParams(ctx)

	// Ensure that provided route is not empty and has valid denom format.
	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	defer func() {
		if r := recover(); r != nil {
			tokenInAmount = sdk.Int{}
			err = fmt.Errorf("function RouteExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	var insExpected []sdk.Int
	insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)

	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}
	insExpected[0] = tokenInMaxAmount

	// Iterates through each routed pool and executes their respective swaps. Note that all of the work to get the return
	// value of this method is done when we calculate insExpected – this for loop primarily serves to execute the actual
	// swaps on each pool.
	for i, routeStep := range route {
		// Get underlying pool type corresponding to the pool ID at the current routeStep.
		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return sdk.Int{}, err
		}

		_tokenOut := tokenOut

		// If there is one pool left in the routeStep, set the expected output of the current swap
		// to the estimated input of the final pool.
		if i != len(route)-1 {
			_tokenOut = sdk.NewCoin(route[i+1].TokenInDenom, insExpected[i+1])
		}

		// Execute the expected swap on the current routed pool
		pool, poolErr := swapModule.GetPool(ctx, routeStep.PoolId)
		if poolErr != nil {
			return sdk.Int{}, poolErr
		}

		// check if pool is active, if not error
		if !pool.IsActive(ctx) {
			return sdk.Int{}, types.InactivePoolError{PoolId: pool.GetId()}
		}

		spreadFactor := pool.GetSpreadFactor(ctx)
		takerFee := pool.GetTakerFee(ctx)
		totalFees := spreadFactor.Add(takerFee)

		_tokenInAmount, swapErr := swapModule.SwapExactAmountOut(ctx, sender, pool, routeStep.TokenInDenom, insExpected[i], _tokenOut, totalFees)
		if swapErr != nil {
			return sdk.Int{}, swapErr
		}

		tokenIn := sdk.NewCoin(routeStep.TokenInDenom, _tokenInAmount)
		err = k.extractTakerFeeToFeePool(ctx, pool, tokenIn, takerFee, poolManagerParams)
		if err != nil {
			return sdk.Int{}, err
		}

		// Sets the final amount of tokens that need to be input into the first pool. Even though this is the final return value for the
		// whole method and will not change after the first iteration, we still iterate through the rest of the pools to execute their respective
		// swaps.
		if i == 0 {
			tokenInAmount = _tokenInAmount
		}
	}

	return tokenInAmount, nil
}

// SplitRouteExactAmountOut route the swap across multiple multihop paths
// to get the desired token in. This is useful for achieving the most optimal execution. However, note that the responsibility
// of determining the optimal split is left to the client. This method simply route the swap across the given route.
// The route must end with the same token out and begin with the same token in.
//
// It performs the price impact protection check on the combination of tokens in from all multihop paths. The given tokenInMaxAmount
// is used for comparison.
//
// Returns error if:
//   - route are empty
//   - route contain duplicate multihop paths
//   - last token out denom is not the same for all multihop paths in routeStep
//   - one of the multihop swaps fails for internal reasons
//   - final token out computed is not positive
//   - final token out computed is smaller than tokenInMaxAmount
func (k Keeper) SplitRouteExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	route []types.SwapAmountOutSplitRoute,
	tokenOutDenom string,
	tokenInMaxAmount sdk.Int,
) (sdk.Int, error) {
	if err := types.ValidateSwapAmountOutSplitRoute(route); err != nil {
		return sdk.Int{}, err
	}

	var (
		// We start the multihop min amount as int max value
		// that is defined as one under the max bit length of sdk.Int
		// which is 256. This is to ensure that we utilize price impact protection
		// on the total of in amount from all multihop paths.
		multihopStartTokenInMaxAmount = intMaxValue
		totalInAmount                 = sdk.ZeroInt()
	)

	for _, multihopRoute := range route {
		tokenOutAmount, err := k.RouteExactAmountOut(
			ctx,
			sender,
			types.SwapAmountOutRoutes(multihopRoute.Pools),
			multihopStartTokenInMaxAmount,
			sdk.NewCoin(tokenOutDenom, multihopRoute.TokenOutAmount))
		if err != nil {
			return sdk.Int{}, err
		}

		totalInAmount = totalInAmount.Add(tokenOutAmount)
	}

	if !totalInAmount.IsPositive() {
		return sdk.Int{}, types.FinalAmountIsNotPositiveError{IsAmountOut: false, Amount: totalInAmount}
	}

	if totalInAmount.GT(tokenInMaxAmount) {
		return sdk.Int{}, types.PriceImpactProtectionExactOutError{Actual: totalInAmount, MaxAmount: tokenInMaxAmount}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSplitRouteSwapExactAmountOut,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyTokensOut, totalInAmount.String()),
		),
	})

	return totalInAmount, nil
}

func (k Keeper) RouteGetPoolDenoms(
	ctx sdk.Context,
	poolId uint64,
) (denoms []string, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return []string{}, err
	}

	denoms, err = swapModule.GetPoolDenoms(ctx, poolId)
	if err != nil {
		return []string{}, err
	}

	return denoms, nil
}

func (k Keeper) RouteCalculateSpotPrice(
	ctx sdk.Context,
	poolId uint64,
	quoteAssetDenom string,
	baseAssetDenom string,
) (price sdk.Dec, err error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	price, err = swapModule.CalculateSpotPrice(ctx, poolId, quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	return price, nil
}

func (k Keeper) MultihopEstimateInGivenExactAmountOut(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) (tokenInAmount sdk.Int, err error) {
	var insExpected []sdk.Int

	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			insExpected = []sdk.Int{}
			err = fmt.Errorf("function MultihopEstimateInGivenExactAmountOut failed due to internal reason: %v", r)
		}
	}()

	routeStep := types.SwapAmountOutRoutes(route)
	if err := routeStep.Validate(); err != nil {
		return sdk.Int{}, err
	}

	// Determine what the estimated input would be for each pool along the multi-hop route
	insExpected, err = k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
	if err != nil {
		return sdk.Int{}, err
	}
	if len(insExpected) == 0 {
		return sdk.Int{}, nil
	}

	return insExpected[0], nil
}

func (k Keeper) GetPool(
	ctx sdk.Context,
	poolId uint64,
) (types.PoolI, error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return nil, err
	}

	return swapModule.GetPool(ctx, poolId)
}

// AllPools returns all pools sorted by their ids
// from every pool module registered in the
// pool manager keeper.
func (k Keeper) AllPools(
	ctx sdk.Context,
) ([]types.PoolI, error) {
	less := func(i, j types.PoolI) bool {
		return i.GetId() < j.GetId()
	}

	//	Allocate the slice with the exact capacity to avoid reallocations.
	poolCount := k.GetNextPoolId(ctx)
	sortedPools := make([]types.PoolI, 0, poolCount)
	for _, poolModule := range k.poolModules {
		currentModulePools, err := poolModule.GetPools(ctx)
		if err != nil {
			return nil, err
		}

		sortedPools = osmoutils.MergeSlices(sortedPools, currentModulePools, less)
	}

	return sortedPools, nil
}

// createMultihopExpectedSwapOuts defines the output denom and output amount for the last pool in
// the routeStep of pools the caller is intending to hop through in a fixed-output multihop tx. It estimates the input
// amount for this last pool and then chains that input as the output of the previous pool in the routeStep, repeating
// until the first pool is reached. It returns an array of inputs, each of which correspond to a pool ID in the
// routeStep of pools for the original multihop transaction.
func (k Keeper) createMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]sdk.Int, error) {
	insExpected := make([]sdk.Int, len(route))
	for i := len(route) - 1; i >= 0; i-- {
		routeStep := route[i]

		swapModule, err := k.GetPoolModule(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		poolI, err := swapModule.GetPool(ctx, routeStep.PoolId)
		if err != nil {
			return nil, err
		}

		spreadFactor := poolI.GetSpreadFactor(ctx)
		takeFee := poolI.GetTakerFee(ctx)
		totalFees := spreadFactor.Add(takeFee)

		tokenIn, err := swapModule.CalcInAmtGivenOut(ctx, poolI, tokenOut, routeStep.TokenInDenom, totalFees)
		if err != nil {
			return nil, err
		}

		insExpected[i] = tokenIn.Amount
		tokenOut = tokenIn
	}

	return insExpected, nil
}

// GetTotalPoolLiquidity gets the total liquidity for a given poolId.
func (k Keeper) GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error) {
	swapModule, err := k.GetPoolModule(ctx, poolId)
	if err != nil {
		return nil, err
	}

	coins, err := swapModule.GetTotalPoolLiquidity(ctx, poolId)
	if err != nil {
		return coins, err
	}

	return coins, nil
}

// TotalLiquidity gets the total liquidity across all pools.
func (k Keeper) TotalLiquidity(ctx sdk.Context) (sdk.Coins, error) {
	totalGammLiquidity, err := k.gammKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalConcentratedLiquidity, err := k.concentratedKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalCosmwasmLiquidity, err := k.cosmwasmpoolKeeper.GetTotalLiquidity(ctx)
	if err != nil {
		return nil, err
	}
	totalLiquidity := totalGammLiquidity.Add(totalConcentratedLiquidity...).Add(totalCosmwasmLiquidity...)
	return totalLiquidity, nil
}

// extractTakerFeeToFeePool takes in a pool and extracts the taker fee from the pool and sends it to the non native fee pool module account.
// Its important to note here that in the original swap, the taker fee + spread fee is sent to the pool's address, so this is why we
// pull directly from the pool and not the user's account.
func (k Keeper) extractTakerFeeToFeePool(ctx sdk.Context, pool types.PoolI, tokenIn sdk.Coin, takerFee sdk.Dec, poolManagerParams types.Params) error {
	var relevantPoolAddress sdk.AccAddress
	nonNativeFeeCollectorForStakingRewardsName := txfeestypes.NonNativeFeeCollectorForStakingRewardsName
	nonNativeFeeCollectorForCommunityPoolName := txfeestypes.NonNativeFeeCollectorForCommunityPoolName
	baseDenom := appparams.BaseCoinUnit

	if pool.GetType() == types.Concentrated {
		// If the pool being swapped from is a concentrated pool, the fee is sent from the spread rewards address
		concentratedTypePool, ok := pool.(cltypes.ConcentratedPoolExtension)
		if !ok {
			return fmt.Errorf("pool %d is not a concentrated pool", pool.GetId())
		}
		relevantPoolAddress = concentratedTypePool.GetSpreadRewardsAddress()
	} else {
		// If the pool being swapped from is not a concentrated pool, the fee is sent from the pool address
		relevantPoolAddress = pool.GetAddress()
	}

	// Take the taker fee from the input token and send it to the fee pool module account
	takerFeeDec := tokenIn.Amount.ToDec().Mul(takerFee)
	takerFeeCoin := sdk.NewCoin(tokenIn.Denom, takerFeeDec.TruncateInt())

	// We determine the distributution of the taker fee based on its denom
	// If the denom is the base denom:
	if takerFeeCoin.Denom == baseDenom {
		// Community Pool:
		if poolManagerParams.OsmoTakerFeeDistribution.CommunityPool.GT(sdk.ZeroDec()) {
			// Osmo community pool funds is a direct send
			osmoTakerFeeToCommunityPoolDec := takerFeeCoin.Amount.ToDec().Mul(poolManagerParams.OsmoTakerFeeDistribution.CommunityPool)
			osmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(baseDenom, osmoTakerFeeToCommunityPoolDec.TruncateInt()))
			err := k.communityPoolKeeper.FundCommunityPool(ctx, osmoTakerFeeToCommunityPoolCoins, relevantPoolAddress)
			if err != nil {
				return err
			}
		}
		// Staking Rewards:
		if poolManagerParams.OsmoTakerFeeDistribution.StakingRewards.GT(sdk.ZeroDec()) {
			// Osmo staking rewards funds are sent to the non native fee pool module account (even though its native, we want to distribute at the same time as the non native fee tokens)
			// We could stream these rewards via the fee collector account, but this is decision to be made by governance.
			osmoTakerFeeToStakingRewardsDec := takerFeeCoin.Amount.ToDec().Mul(poolManagerParams.OsmoTakerFeeDistribution.StakingRewards)
			osmoTakerFeeToStakingRewardsCoins := sdk.NewCoins(sdk.NewCoin(baseDenom, osmoTakerFeeToStakingRewardsDec.TruncateInt()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, relevantPoolAddress, nonNativeFeeCollectorForStakingRewardsName, osmoTakerFeeToStakingRewardsCoins)
			if err != nil {
				return err
			}
		}

		// If the denom is not the base denom:
	} else {
		// Community Pool:
		if poolManagerParams.NonOsmoTakerFeeDistribution.CommunityPool.GT(sdk.ZeroDec()) {
			denomIsWhitelisted := isDenomWhitelisted(takerFeeCoin.Denom, poolManagerParams.AuthorizedQuoteDenoms)
			// If the non osmo denom is a whitelisted quote asset, we send to the community pool
			if denomIsWhitelisted {
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeCoin.Amount.ToDec().Mul(poolManagerParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt()))
				err := k.communityPoolKeeper.FundCommunityPool(ctx, nonOsmoTakerFeeToCommunityPoolCoins, relevantPoolAddress)
				if err != nil {
					return err
				}
			} else {
				// If the non osmo denom is not a whitelisted asset, we send to the non native fee pool for community pool module account.
				// At epoch, this account swaps the non native, non whitelisted assets for XXX and sends to the community pool.
				nonOsmoTakerFeeToCommunityPoolDec := takerFeeCoin.Amount.ToDec().Mul(poolManagerParams.NonOsmoTakerFeeDistribution.CommunityPool)
				nonOsmoTakerFeeToCommunityPoolCoins := sdk.NewCoins(sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToCommunityPoolDec.TruncateInt()))
				err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, relevantPoolAddress, nonNativeFeeCollectorForCommunityPoolName, nonOsmoTakerFeeToCommunityPoolCoins)
				if err != nil {
					return err
				}
			}
		}
		// Staking Rewards:
		if poolManagerParams.NonOsmoTakerFeeDistribution.StakingRewards.GT(sdk.ZeroDec()) {
			// Non Osmo staking rewards are sent to the non native fee pool module account
			nonOsmoTakerFeeToStakingRewardsDec := takerFeeCoin.Amount.ToDec().Mul(poolManagerParams.NonOsmoTakerFeeDistribution.StakingRewards)
			nonOsmoTakerFeeToStakingRewardsCoins := sdk.NewCoins(sdk.NewCoin(tokenIn.Denom, nonOsmoTakerFeeToStakingRewardsDec.TruncateInt()))
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, relevantPoolAddress, nonNativeFeeCollectorForStakingRewardsName, nonOsmoTakerFeeToStakingRewardsCoins)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isDenomWhitelisted(denom string, authorizedQuoteDenoms []string) bool {
	for _, authorizedQuoteDenom := range authorizedQuoteDenoms {
		if denom == authorizedQuoteDenom {
			return true
		}
	}
	return false
}
