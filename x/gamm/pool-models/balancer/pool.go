package balancer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

//nolint:deadcode
const (
	nonPostiveSharesAmountErrFormat = "shares amount must be positive, was %s"
	nonPostiveTokenAmountErrFormat  = "token amount must be positive, was %s"
	sharesLargerThanMaxErrFormat    = "%s resulted shares is larger than the max amount of %s"
	invalidInputDenomsErrFormat     = "input denoms must already exist in the pool (%s)"

	failedInterimLiquidityUpdateErrFormat        = "failed to update interim liquidity - pool asset %s does not exist"
	formatRepeatingPoolAssetsNotAllowedErrFormat = "repeating pool assets not allowed, found %s"
	formatNoPoolAssetFoundErrFormat              = "can't find the PoolAsset (%s)"
)

var (
	_ poolmanagertypes.PoolI       = &Pool{}
	_ types.PoolAmountOutExtension = &Pool{}
	_ types.WeightedPoolExtension  = &Pool{}
	_ types.CFMMPoolI              = &Pool{}
)

// NewPool returns a weighted CPMM pool with the provided parameters, and initial assets.
// Invariants that are assumed to be satisfied and not checked:
// (This is handled in ValidateBasic)
// * 2 <= len(assets) <= 8
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewBalancerPool(poolId uint64, balancerPoolParams PoolParams, assets []PoolAsset, futureGovernor string, blockTime time.Time) (Pool, error) {
	poolAddr := poolmanagertypes.NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address:            poolAddr.String(),
		Id:                 poolId,
		PoolParams:         PoolParams{},
		TotalWeight:        sdk.ZeroInt(),
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply),
		PoolAssets:         nil,
		FuturePoolGovernor: futureGovernor,
	}

	err := pool.SetInitialPoolAssets(assets)
	if err != nil {
		return Pool{}, err
	}

	sortedPoolAssets := pool.GetAllPoolAssets()
	err = balancerPoolParams.Validate(sortedPoolAssets)
	if err != nil {
		return Pool{}, err
	}

	err = pool.setInitialPoolParams(balancerPoolParams, sortedPoolAssets, blockTime)
	if err != nil {
		return Pool{}, err
	}

	return *pool, nil
}

// GetAddress returns the address of a pool.
// If the pool address is not bech32 valid, it returns an empty address.
func (p Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", p.GetId()))
	}
	return addr
}

func (p Pool) GetId() uint64 {
	return p.Id
}

func (p Pool) GetSpreadFactor(_ sdk.Context) sdk.Dec {
	return p.PoolParams.SwapFee
}

func (p Pool) GetTotalPoolLiquidity(_ sdk.Context) sdk.Coins {
	return poolAssetsCoins(p.PoolAssets)
}

func (p Pool) GetExitFee(_ sdk.Context) sdk.Dec {
	return p.PoolParams.ExitFee
}

func (p Pool) GetPoolParams() PoolParams {
	return p.PoolParams
}

func (p Pool) GetTotalWeight() sdk.Int {
	return p.TotalWeight
}

func (p Pool) GetTotalShares() sdk.Int {
	return p.TotalShares.Amount
}

func (p *Pool) AddTotalShares(amt sdk.Int) {
	p.TotalShares.Amount = p.TotalShares.Amount.Add(amt)
}

func (p *Pool) SubTotalShares(amt sdk.Int) {
	p.TotalShares.Amount = p.TotalShares.Amount.Sub(amt)
}

// SetInitialPoolAssets sets the PoolAssets in the pool. It is only designed to
// be called at the pool's creation. If the same denom's PoolAsset exists, will
// return error.
//
// The list of PoolAssets must be sorted. This is done to enable fast searching
// for a PoolAsset by denomination.
// TODO: Unify story for validation of []PoolAsset, some is here, some is in
// CreatePool.ValidateBasic()
func (p *Pool) SetInitialPoolAssets(PoolAssets []PoolAsset) error {
	exists := make(map[string]bool)
	for _, asset := range p.PoolAssets {
		exists[asset.Token.Denom] = true
	}

	newTotalWeight := p.TotalWeight
	scaledPoolAssets := make([]PoolAsset, 0, len(PoolAssets))

	// TODO: Refactor this into PoolAsset.validate()
	for _, asset := range PoolAssets {
		if asset.Token.Amount.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't add the zero or negative balance of token")
		}

		err := asset.validateWeight()
		if err != nil {
			return err
		}

		if exists[asset.Token.Denom] {
			return fmt.Errorf("same PoolAsset already exists")
		}
		exists[asset.Token.Denom] = true

		// Scale weight from the user provided weight to the correct internal weight
		asset.Weight = asset.Weight.MulRaw(GuaranteedWeightPrecision)
		scaledPoolAssets = append(scaledPoolAssets, asset)
		newTotalWeight = newTotalWeight.Add(asset.Weight)
	}

	// TODO: Change this to a more efficient sorted insert algorithm.
	// Furthermore, consider changing the underlying data type to allow in-place modification if the
	// number of PoolAssets is expected to be large.
	p.PoolAssets = append(p.PoolAssets, scaledPoolAssets...)
	sortPoolAssetsByDenom(p.PoolAssets)

	p.TotalWeight = newTotalWeight

	return nil
}

// setInitialPoolParams
func (p *Pool) setInitialPoolParams(params PoolParams, sortedAssets []PoolAsset, curBlockTime time.Time) error {
	p.PoolParams = params
	if params.SmoothWeightChangeParams != nil {
		// set initial assets
		initialWeights := make([]PoolAsset, len(sortedAssets))
		for i, v := range sortedAssets {
			initialWeights[i] = PoolAsset{
				Weight: v.Weight,
				Token:  sdk.Coin{Denom: v.Token.Denom, Amount: sdk.ZeroInt()},
			}
		}
		params.SmoothWeightChangeParams.InitialPoolWeights = initialWeights

		// sort target weights by denom
		targetPoolWeights := params.SmoothWeightChangeParams.TargetPoolWeights
		sortPoolAssetsByDenom(targetPoolWeights)

		// scale target pool weights by GuaranteedWeightPrecision
		for i, v := range targetPoolWeights {
			err := ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
			p.PoolParams.SmoothWeightChangeParams.TargetPoolWeights[i] = PoolAsset{
				Weight: v.Weight.MulRaw(GuaranteedWeightPrecision),
				Token:  v.Token,
			}
		}

		// Set start time if not present.
		if params.SmoothWeightChangeParams.StartTime.Unix() <= 0 {
			// Per https://golang.org/pkg/time/#Time.Unix, should be timezone independent
			params.SmoothWeightChangeParams.StartTime = time.Unix(curBlockTime.Unix(), 0)
		}
	}

	return nil
}

// GetPoolAssets returns the denom's PoolAsset, If the PoolAsset doesn't exist, will return error.
// As above, it will search the denom's PoolAsset by using binary search.
// So, it is important to make sure that the PoolAssets are sorted.
func (p Pool) GetPoolAsset(denom string) (PoolAsset, error) {
	_, asset, err := p.getPoolAssetAndIndex(denom)
	return asset, err
}

// Returns a pool asset, and its index. If err != nil, then the index will be valid.
func (p Pool) getPoolAssetAndIndex(denom string) (int, PoolAsset, error) {
	if denom == "" {
		return -1, PoolAsset{}, fmt.Errorf("you tried to find the PoolAsset with empty denom")
	}

	if len(p.PoolAssets) == 0 {
		return -1, PoolAsset{}, errorsmod.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(formatNoPoolAssetFoundErrFormat, denom))
	}

	i := sort.Search(len(p.PoolAssets), func(i int) bool {
		PoolAssetA := p.PoolAssets[i]

		compare := strings.Compare(PoolAssetA.Token.Denom, denom)
		return compare >= 0
	})

	if i < 0 || i >= len(p.PoolAssets) {
		return -1, PoolAsset{}, errorsmod.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(formatNoPoolAssetFoundErrFormat, denom))
	}

	if p.PoolAssets[i].Token.Denom != denom {
		return -1, PoolAsset{}, errorsmod.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(formatNoPoolAssetFoundErrFormat, denom))
	}

	return i, p.PoolAssets[i], nil
}

func (p Pool) parsePoolAssetsByDenoms(tokenADenom, tokenBDenom string) (
	Aasset PoolAsset, Basset PoolAsset, err error,
) {
	Aasset, found1 := getPoolAssetByDenom(p.PoolAssets, tokenADenom)
	Basset, found2 := getPoolAssetByDenom(p.PoolAssets, tokenBDenom)

	if !found1 {
		return PoolAsset{}, PoolAsset{}, fmt.Errorf("(%s) does not exist in the pool", tokenADenom)
	}
	if !found2 {
		return PoolAsset{}, PoolAsset{}, fmt.Errorf("(%s) does not exist in the pool", tokenBDenom)
	}
	return Aasset, Basset, nil
}

func (p Pool) parsePoolAssets(tokensA sdk.Coins, tokenBDenom string) (
	tokenA sdk.Coin, Aasset PoolAsset, Basset PoolAsset, err error,
) {
	if len(tokensA) != 1 {
		return tokenA, Aasset, Basset, errors.New("expected tokensB to be of length one")
	}
	Aasset, Basset, err = p.parsePoolAssetsByDenoms(tokensA[0].Denom, tokenBDenom)
	if err != nil {
		return sdk.Coin{}, PoolAsset{}, PoolAsset{}, err
	}
	return tokensA[0], Aasset, Basset, nil
}

func (p Pool) parsePoolAssetsCoins(tokensA sdk.Coins, tokensB sdk.Coins) (
	Aasset PoolAsset, Basset PoolAsset, err error,
) {
	if len(tokensB) != 1 {
		return Aasset, Basset, errors.New("expected tokensA to be of length one")
	}
	_, Aasset, Basset, err = p.parsePoolAssets(tokensA, tokensB[0].Denom)
	return Aasset, Basset, err
}

func (p *Pool) IncreaseLiquidity(sharesOut sdk.Int, coinsIn sdk.Coins) {
	err := p.addToPoolAssetBalances(coinsIn)
	if err != nil {
		panic(err)
	}
	p.AddTotalShares(sharesOut)
}

func (p *Pool) UpdatePoolAssetBalance(coin sdk.Coin) error {
	// Check that PoolAsset exists.
	assetIndex, existingAsset, err := p.getPoolAssetAndIndex(coin.Denom)
	if err != nil {
		return err
	}

	if coin.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("can't set the pool's balance of a token to be zero or negative")
	}

	// Update the supply of the asset
	existingAsset.Token = coin
	p.PoolAssets[assetIndex] = existingAsset
	return nil
}

func (p *Pool) UpdatePoolAssetBalances(coins sdk.Coins) error {
	// Ensures that there are no duplicate denoms, all denom's are valid,
	// and amount is > 0
	err := coins.Validate()
	if err != nil {
		return fmt.Errorf("provided coins are invalid, %v", err)
	}

	for _, coin := range coins {
		// TODO: We may be able to make this log(|coins|) faster in how it
		// looks up denom -> Coin by doing a multi-search,
		// but as we don't anticipate |coins| to be large, we omit this.
		err = p.UpdatePoolAssetBalance(coin)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) addToPoolAssetBalances(coins sdk.Coins) error {
	for _, coin := range coins {
		i, poolAsset, err := p.getPoolAssetAndIndex(coin.Denom)
		if err != nil {
			return err
		}
		poolAsset.Token.Amount = poolAsset.Token.Amount.Add(coin.Amount)
		p.PoolAssets[i] = poolAsset
	}
	return nil
}

func (p Pool) GetPoolAssets(denoms ...string) ([]PoolAsset, error) {
	result := make([]PoolAsset, 0, len(denoms))

	for _, denom := range denoms {
		PoolAsset, err := p.GetPoolAsset(denom)
		if err != nil {
			return nil, err
		}

		result = append(result, PoolAsset)
	}

	return result, nil
}

func (p Pool) GetAllPoolAssets() []PoolAsset {
	copyslice := make([]PoolAsset, len(p.PoolAssets))
	copy(copyslice, p.PoolAssets)
	return copyslice
}

// updateAllWeights updates all of the pool's internal weights to be equal to
// the new weights. It assumes that `newWeights` are sorted by denomination,
// and only contain the same denominations as the pool already contains.
// This does not affect the asset balances.
// If any of the above are not satisfied, this will panic.
// (As all input to this should be generated from the state machine)
// TODO: (post-launch) If newWeights includes a new denomination,
// add the balance as well to the pool's internal measurements.
// TODO: (post-launch) If newWeights excludes an existing denomination,
// remove the weight from the pool, and figure out something to do
// with any remaining coin.
func (p *Pool) updateAllWeights(newWeights []PoolAsset) {
	if len(p.PoolAssets) != len(newWeights) {
		panic("updateAllWeights called with invalid input, len(newWeights) != len(existingWeights)")
	}
	totalWeight := sdk.ZeroInt()
	for i, asset := range p.PoolAssets {
		if asset.Token.Denom != newWeights[i].Token.Denom {
			panic(fmt.Sprintf("updateAllWeights called with invalid input, "+
				"expected new weights' %vth asset to be %v, got %v",
				i, asset.Token.Denom, newWeights[i].Token.Denom))
		}
		err := newWeights[i].validateWeight()
		if err != nil {
			panic("updateAllWeights: Tried to set an invalid weight")
		}
		p.PoolAssets[i].Weight = newWeights[i].Weight
		totalWeight = totalWeight.Add(p.PoolAssets[i].Weight)
	}
	p.TotalWeight = totalWeight
}

// PokePool checks to see if the pool's token weights need to be updated, and
// if so, does so. Currently doesn't do anything outside out LBPs.
func (p *Pool) PokePool(blockTime time.Time) {
	// check if pool weights didn't change
	poolWeightsChanging := p.PoolParams.SmoothWeightChangeParams != nil
	if !poolWeightsChanging {
		return
	}

	params := *p.PoolParams.SmoothWeightChangeParams

	// The weights w(t) for the pool at time `t` is defined in one of three
	// possible ways:
	//
	// 1. t <= start_time: w(t) = initial_pool_weights
	//
	// 2. start_time < t <= start_time + duration:
	//     w(t) = initial_pool_weights + (t - start_time) *
	//       (target_pool_weights - initial_pool_weights) / (duration)
	//
	// 3. t > start_time + duration: w(t) = target_pool_weights
	switch {
	case blockTime.Before(params.StartTime) || params.StartTime.Equal(blockTime):
		// case 1: t <= start_time
		return

	case blockTime.After(params.StartTime.Add(params.Duration)):
		// case 2: start_time < t <= start_time + duration:

		// Update weights to be the target weights.
		//
		// TODO: When we add support for adding new assets via this method, ensure
		// the new asset has some token sent with it.
		p.updateAllWeights(params.TargetPoolWeights)

		// we've finished updating the weights, so reset the following fields
		p.PoolParams.SmoothWeightChangeParams = nil
		return

	default:
		// case 3: t > start_time + duration: w(t) = target_pool_weights

		shiftedBlockTime := blockTime.Sub(params.StartTime).Milliseconds()
		percentDurationElapsed := sdk.NewDec(shiftedBlockTime).QuoInt64(params.Duration.Milliseconds())

		// If the duration elapsed is equal to the total time, or a rounding error
		// makes it seem like it is, just set to target weight.
		if percentDurationElapsed.GTE(sdk.OneDec()) {
			p.updateAllWeights(params.TargetPoolWeights)
			return
		}

		// below will be auto-truncated according to internal weight precision routine
		totalWeightsDiff := subPoolAssetWeights(params.TargetPoolWeights, params.InitialPoolWeights)
		scaledDiff := poolAssetsMulDec(totalWeightsDiff, percentDurationElapsed)
		updatedWeights := addPoolAssetWeights(params.InitialPoolWeights, scaledDiff)

		p.updateAllWeights(updatedWeights)
	}
}

func (p Pool) GetTokenWeight(denom string) (sdk.Int, error) {
	PoolAsset, err := p.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Weight, nil
}

func (p Pool) GetTokenBalance(denom string) (sdk.Int, error) {
	PoolAsset, err := p.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Token.Amount, nil
}

func (p Pool) NumAssets() int {
	return len(p.PoolAssets)
}

func (p Pool) IsActive(ctx sdk.Context) bool {
	return true
}

func (p Pool) GetType() poolmanagertypes.PoolType {
	return poolmanagertypes.Balancer
}

// CalcOutAmtGivenIn calculates tokens to be swapped out given the provided
// amount and fee deducted, using solveConstantFunctionInvariant.
func (p Pool) CalcOutAmtGivenIn(
	ctx sdk.Context,
	tokensIn sdk.Coins,
	tokenOutDenom string,
	spreadFactor sdk.Dec,
) (sdk.Coin, error) {
	tokenIn, poolAssetIn, poolAssetOut, err := p.parsePoolAssets(tokensIn, tokenOutDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	tokenAmountInAfterFee := tokenIn.Amount.ToDec().Mul(sdk.OneDec().Sub(spreadFactor))
	poolTokenInBalance := poolAssetIn.Token.Amount.ToDec()
	poolPostSwapInBalance := poolTokenInBalance.Add(tokenAmountInAfterFee)

	// deduct spread factor on the tokensIn
	// delta balanceOut is positive(tokens inside the pool decreases)
	tokenAmountOut := solveConstantFunctionInvariant(
		poolTokenInBalance,
		poolPostSwapInBalance,
		poolAssetIn.Weight.ToDec(),
		poolAssetOut.Token.Amount.ToDec(),
		poolAssetOut.Weight.ToDec(),
	)

	// We ignore the decimal component, as we round down the token amount out.
	tokenAmountOutInt := tokenAmountOut.TruncateInt()
	if !tokenAmountOutInt.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}

	return sdk.NewCoin(tokenOutDenom, tokenAmountOutInt), nil
}

// SwapOutAmtGivenIn is a mutative method for CalcOutAmtGivenIn, which includes the actual swap.
func (p *Pool) SwapOutAmtGivenIn(
	ctx sdk.Context,
	tokensIn sdk.Coins,
	tokenOutDenom string,
	spreadFactor sdk.Dec,
) (
	tokenOut sdk.Coin, err error,
) {
	tokenOutCoin, err := p.CalcOutAmtGivenIn(ctx, tokensIn, tokenOutDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, tokensIn, sdk.Coins{tokenOutCoin})
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenOutCoin, nil
}

// CalcInAmtGivenOut calculates token to be provided, fee added,
// given the swapped out amount, using solveConstantFunctionInvariant.
func (p Pool) CalcInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, spreadFactor sdk.Dec) (
	tokenIn sdk.Coin, err error,
) {
	tokenOut, poolAssetOut, poolAssetIn, err := p.parsePoolAssets(tokensOut, tokenInDenom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// delta balanceOut is positive(tokens inside the pool decreases)
	poolTokenOutBalance := poolAssetOut.Token.Amount.ToDec()
	poolPostSwapOutBalance := poolTokenOutBalance.Sub(tokenOut.Amount.ToDec())
	// (x_0)(y_0) = (x_0 + in)(y_0 - out)
	tokenAmountIn := solveConstantFunctionInvariant(
		poolTokenOutBalance, poolPostSwapOutBalance, poolAssetOut.Weight.ToDec(),
		poolAssetIn.Token.Amount.ToDec(), poolAssetIn.Weight.ToDec()).Neg()

	// We deduct a spread factor on the input asset. The swap happens by following the invariant curve on the input * (1 - spread factor)
	// and then the spread factor is added to the pool.
	// Thus in order to give X amount out, we solve the invariant for the invariant input. However invariant input = (1 - spread factor) * trade input.
	// Therefore we divide by (1 - spread factor) here
	tokenAmountInBeforeFee := tokenAmountIn.Quo(sdk.OneDec().Sub(spreadFactor))

	// We round up tokenInAmt, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmt := tokenAmountInBeforeFee.Ceil().TruncateInt()

	if !tokenInAmt.IsPositive() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidMathApprox, "token amount must be positive")
	}
	return sdk.NewCoin(tokenInDenom, tokenInAmt), nil
}

// SwapInAmtGivenOut is a mutative method for CalcOutAmtGivenIn, which includes the actual swap.
func (p *Pool) SwapInAmtGivenOut(
	ctx sdk.Context, tokensOut sdk.Coins, tokenInDenom string, spreadFactor sdk.Dec) (
	tokenIn sdk.Coin, err error,
) {
	tokenInCoin, err := p.CalcInAmtGivenOut(ctx, tokensOut, tokenInDenom, spreadFactor)
	if err != nil {
		return sdk.Coin{}, err
	}

	err = p.applySwap(ctx, sdk.Coins{tokenInCoin}, tokensOut)
	if err != nil {
		return sdk.Coin{}, err
	}
	return tokenInCoin, nil
}

// ApplySwap.
func (p *Pool) applySwap(ctx sdk.Context, tokensIn sdk.Coins, tokensOut sdk.Coins) error {
	// Fixed gas consumption per swap to prevent spam
	ctx.GasMeter().ConsumeGas(types.BalancerGasFeeForSwap, "balancer swap computation")
	// Also ensures that len(tokensIn) = 1 = len(tokensOut)
	inPoolAsset, outPoolAsset, err := p.parsePoolAssetsCoins(tokensIn, tokensOut)
	if err != nil {
		return err
	}
	inPoolAsset.Token.Amount = inPoolAsset.Token.Amount.Add(tokensIn[0].Amount)
	outPoolAsset.Token.Amount = outPoolAsset.Token.Amount.Sub(tokensOut[0].Amount)

	return p.UpdatePoolAssetBalances(sdk.NewCoins(
		inPoolAsset.Token,
		outPoolAsset.Token,
	))
}

// SpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool.
// To reduce the propagated effect of incorrect trailing digits,
// we take the ratio of weights and divide this by ratio of supplies
// this is equivalent to spot_price = (Quote Supply / Quote Weight) / (Base Supply / Base Weight)
//
// As an example, assume equal weights. uosmo supply of 2 and uatom supply of 4.
//
// Case 1: base = uosmo, quote = uatom -> for one uosmo, get 2 uatom = 4 / 2 = 2
// In other words, it costs 2 uatom to get one uosmo.
//
// Case 2: base = uatom, quote = uosmo -> for one uatom, get 0.5 uosmo = 2 / 4 = 0.5
// In other words, it costs 0.5 uosmo to get one uatom.
//
// panics if the pool in state is incorrect, and has any weight that is 0.
func (p Pool) SpotPrice(ctx sdk.Context, quoteAsset, baseAsset string) (spotPrice sdk.Dec, err error) {
	quote, base, err := p.parsePoolAssetsByDenoms(quoteAsset, baseAsset)
	if err != nil {
		return sdk.Dec{}, err
	}
	if base.Weight.IsZero() || quote.Weight.IsZero() {
		return sdk.Dec{}, errors.New("pool is misconfigured, got 0 weight")
	}

	// spot_price = (Quote Supply / Quote Weight) / (Base Supply / Base Weight)
	//            = (Quote Supply / Quote Weight) * (Base Weight / Base Supply)
	//            = (Base Weight  / Quote Weight) * (Quote Supply / Base Supply)
	invWeightRatio := base.Weight.ToDec().Quo(quote.Weight.ToDec())
	supplyRatio := quote.Token.Amount.ToDec().Quo(base.Token.Amount.ToDec())
	spotPrice = supplyRatio.Mul(invWeightRatio)

	return spotPrice, err
}

// calcPoolOutGivenSingleIn - balance pAo.
func (p *Pool) calcSingleAssetJoin(tokenIn sdk.Coin, spreadFactor sdk.Dec, tokenInPoolAsset PoolAsset, totalShares sdk.Int) (numShares sdk.Int, err error) {
	_, err = p.GetPoolAsset(tokenIn.Denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	totalWeight := p.GetTotalWeight()
	if totalWeight.IsZero() {
		return sdk.ZeroInt(), errors.New("pool misconfigured, total weight = 0")
	}
	normalizedWeight := tokenInPoolAsset.Weight.ToDec().Quo(totalWeight.ToDec())
	return calcPoolSharesOutGivenSingleAssetIn(
		tokenInPoolAsset.Token.Amount.ToDec(),
		normalizedWeight,
		totalShares.ToDec(),
		tokenIn.Amount.ToDec(),
		spreadFactor,
	).TruncateInt(), nil
}

// JoinPool calculates the number of shares needed given tokensIn with spreadFactor applied.
// It updates the liquidity if the pool is joined successfully. If not, returns error.
// and updates pool accordingly.
func (p *Pool) JoinPool(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, err error) {
	numShares, newLiquidity, err := p.CalcJoinPoolShares(ctx, tokensIn, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	// update pool with the calculated share and liquidity needed to join pool
	p.IncreaseLiquidity(numShares, newLiquidity)
	return numShares, nil
}

// JoinPoolNoSwap calculates the number of shares needed for an all-asset join given tokensIn with spreadFactor applied.
// It updates the liquidity if the pool is joined successfully. If not, returns error.
func (p *Pool) JoinPoolNoSwap(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, err error) {
	numShares, tokensJoined, err := p.CalcJoinPoolNoSwapShares(ctx, tokensIn, spreadFactor)
	if err != nil {
		return sdk.Int{}, err
	}

	// update pool with the calculated share and liquidity needed to join pool
	p.IncreaseLiquidity(numShares, tokensJoined)
	return numShares, nil
}

// CalcJoinPoolShares calculates the number of shares created to join pool with the provided amount of `tokenIn`.
// The input tokens must either be:
// - a single token
// - contain exactly the same tokens as the pool contains
//
// It returns the number of shares created, the amount of coins actually joined into the pool
// (in case of not being able to fully join), or an error.
func (p *Pool) CalcJoinPoolShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	// 1) Get pool current liquidity + and token weights
	// 2) If single token provided, do single asset join and exit.
	// 3) If multi-asset join, first do as much of a join as we can with no swaps.
	// 4) Update pool shares / liquidity / remaining tokens to join accordingly
	// 5) For every remaining token to LP, do a single asset join, and update pool shares / liquidity.
	//
	// Note that all single asset joins do incur spread factor.
	//
	// Since CalcJoinPoolShares is non-mutative, the steps for updating pool shares / liquidity are
	// more complex / don't just alter the state.
	// We should simplify this logic further in the future, using balancer multi-join equations.

	// 1) get all 'pool assets' (aka current pool liquidity + balancer weight)
	poolAssetsByDenom, err := getPoolAssetsByDenom(p.GetAllPoolAssets())
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	// check to make sure the input denom exists in the pool
	err = ensureDenomInPool(poolAssetsByDenom, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	totalShares := p.GetTotalShares()
	if tokensIn.Len() == 1 {
		// 2) Single token provided, so do single asset join and exit.
		numShares, err = p.calcSingleAssetJoin(tokensIn[0], spreadFactor, poolAssetsByDenom[tokensIn[0].Denom], totalShares)
		if err != nil {
			return sdk.ZeroInt(), sdk.NewCoins(), err
		}
		// we join all the tokens.
		tokensJoined = tokensIn
		return numShares, tokensJoined, nil
	} else if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("balancer pool only supports LP'ing with one asset or all assets in pool")
	}

	// 3) JoinPoolNoSwap with as many tokens as we can. (What is in perfect ratio)
	// * numShares is how many shares are perfectly matched.
	// * remainingTokensIn is how many coins we have left to join, that have not already been used.
	// if remaining coins is empty, logic is done (we joined all tokensIn)
	numShares, tokensJoined, err = p.CalcJoinPoolNoSwapShares(ctx, tokensIn, spreadFactor)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	// safely ends the calculation if all input tokens are successfully LP'd
	if tokensJoined.IsAnyGT(tokensIn) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("an error has occurred, more coins joined than tokens passed in")
	} else if tokensJoined.IsEqual(tokensIn) {
		return numShares, tokensJoined, nil
	}

	// 4) Still more coins to join, so we update our pool tracker map here to account for
	// join that just happened. Importantly, this step does not actually change the pool state.
	// Instead, it mutates the pool assets argument to be further used by the caller.
	// * We add the joined coins to our "current pool liquidity" object (poolAssetsByDenom)
	// * We increment a variable for our "newTotalShares" to add in the shares that've been added.
	if err := updateIntermediaryPoolAssetsLiquidity(tokensJoined, poolAssetsByDenom); err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	newTotalShares := totalShares.Add(numShares)

	// 5) Now single asset join each remaining coin.
	remainingTokensIn := tokensIn.Sub(tokensJoined)
	newNumSharesFromRemaining, newLiquidityFromRemaining, err := p.calcJoinSingleAssetTokensIn(remainingTokensIn, newTotalShares, poolAssetsByDenom, spreadFactor)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}
	// update total amount LP'd variable, and total new LP shares variable, run safety check, and return
	numShares = numShares.Add(newNumSharesFromRemaining)
	tokensJoined = tokensJoined.Add(newLiquidityFromRemaining...)

	if tokensJoined.IsAnyGT(tokensIn) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("an error has occurred, more coins joined than token In")
	}

	return numShares, tokensJoined, nil
}

// CalcJoinPoolNoSwapShares calculates the number of shares created to execute an all-asset pool join with the provided amount of `tokensIn`.
// The input tokens must contain the same tokens as in the pool.
//
// Returns the number of shares created, the amount of coins actually joined into the pool, (in case of not being able to fully join),
// and the remaining tokens in `tokensIn` after joining. If an all-asset join is not possible, returns an error.
//
// Since CalcJoinPoolNoSwapShares is non-mutative, the steps for updating pool shares / liquidity are
// more complex / don't just alter the state.
// We should simplify this logic further in the future using multi-join equations.
func (p *Pool) CalcJoinPoolNoSwapShares(ctx sdk.Context, tokensIn sdk.Coins, spreadFactor sdk.Dec) (numShares sdk.Int, tokensJoined sdk.Coins, err error) {
	// get all 'pool assets' (aka current pool liquidity + balancer weight)
	poolAssetsByDenom, err := getPoolAssetsByDenom(p.GetAllPoolAssets())
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	err = ensureDenomInPool(poolAssetsByDenom, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	// ensure that there aren't too many or too few assets in `tokensIn`
	if tokensIn.Len() != p.NumAssets() {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("no-swap joins require LP'ing with all assets in pool")
	}

	// execute a no-swap join with as many tokens as possible given a perfect ratio:
	// * numShares is how many shares are perfectly matched.
	// * remainingTokensIn is how many coins we have left to join that have not already been used.
	numShares, remainingTokensIn, err := cfmm_common.MaximalExactRatioJoin(p, ctx, tokensIn)
	if err != nil {
		return sdk.ZeroInt(), sdk.NewCoins(), err
	}

	// ensure that no more tokens have been joined than is possible with the given `tokensIn`
	tokensJoined = tokensIn.Sub(remainingTokensIn)
	if tokensJoined.IsAnyGT(tokensIn) {
		return sdk.ZeroInt(), sdk.NewCoins(), errors.New("an error has occurred, more coins joined than token In")
	}

	return numShares, tokensJoined, nil
}

// calcJoinSingleAssetTokensIn attempts to calculate single
// asset join for all tokensIn given totalShares in pool,
// poolAssetsByDenom and spreadFactor. totalShares is the number
// of shares in pool before beginnning to join any of the tokensIn.
//
// Returns totalNewShares and totalNewLiquidity from joining all tokensIn
// by mimicking individually single asset joining each.
// or error if fails to calculate join for any of the tokensIn.
func (p *Pool) calcJoinSingleAssetTokensIn(tokensIn sdk.Coins, totalShares sdk.Int, poolAssetsByDenom map[string]PoolAsset, spreadFactor sdk.Dec) (sdk.Int, sdk.Coins, error) {
	totalNewShares := sdk.ZeroInt()
	totalNewLiquidity := sdk.NewCoins()
	for _, coin := range tokensIn {
		newShares, err := p.calcSingleAssetJoin(coin, spreadFactor, poolAssetsByDenom[coin.Denom], totalShares.Add(totalNewShares))
		if err != nil {
			return sdk.ZeroInt(), sdk.Coins{}, err
		}

		totalNewLiquidity = totalNewLiquidity.Add(coin)
		totalNewShares = totalNewShares.Add(newShares)
	}
	return totalNewShares, totalNewLiquidity, nil
}

func (p *Pool) ExitPool(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitingCoins sdk.Coins, err error) {
	exitingCoins, err = p.CalcExitPoolCoinsFromShares(ctx, exitingShares, exitFee)
	if err != nil {
		return sdk.Coins{}, err
	}

	if err := p.exitPool(ctx, exitingCoins, exitingShares); err != nil {
		return sdk.Coins{}, err
	}

	return exitingCoins, nil
}

// exitPool exits the pool given exitingCoins and exitingShares.
// updates the pool's liquidity and totalShares.
func (p *Pool) exitPool(ctx sdk.Context, exitingCoins sdk.Coins, exitingShares sdk.Int) error {
	balances := p.GetTotalPoolLiquidity(ctx).Sub(exitingCoins)
	if err := p.UpdatePoolAssetBalances(balances); err != nil {
		return err
	}

	totalShares := p.GetTotalShares()
	p.TotalShares = sdk.NewCoin(p.TotalShares.Denom, totalShares.Sub(exitingShares))

	return nil
}

func (p *Pool) CalcExitPoolCoinsFromShares(ctx sdk.Context, exitingShares sdk.Int, exitFee sdk.Dec) (exitedCoins sdk.Coins, err error) {
	return cfmm_common.CalcExitPool(ctx, p, exitingShares, exitFee)
}

func (p *Pool) CalcTokenInShareAmountOut(
	ctx sdk.Context,
	tokenInDenom string,
	shareOutAmount sdk.Int,
	spreadFactor sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	_, poolAssetIn, err := p.getPoolAssetAndIndex(tokenInDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := poolAssetIn.Weight.ToDec().Quo(p.GetTotalWeight().ToDec())

	// We round up tokenInAmount, as this is whats charged for the swap, for the precise amount out.
	// Otherwise, the pool would under-charge by this rounding error.
	tokenInAmount = calcSingleAssetInGivenPoolSharesOut(
		poolAssetIn.Token.Amount.ToDec(),
		normalizedWeight,
		p.GetTotalShares().ToDec(),
		shareOutAmount.ToDec(),
		spreadFactor,
	).Ceil().TruncateInt()

	if !tokenInAmount.IsPositive() {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrNotPositiveRequireAmount, nonPostiveTokenAmountErrFormat, tokenInAmount)
	}

	return tokenInAmount, nil
}

func (p *Pool) JoinPoolTokenInMaxShareAmountOut(
	ctx sdk.Context,
	tokenInDenom string,
	shareOutAmount sdk.Int,
) (tokenInAmount sdk.Int, err error) {
	_, poolAssetIn, err := p.getPoolAssetAndIndex(tokenInDenom)
	if err != nil {
		return sdk.Int{}, err
	}

	normalizedWeight := poolAssetIn.Weight.ToDec().Quo(p.GetTotalWeight().ToDec())

	tokenInAmount = calcSingleAssetInGivenPoolSharesOut(
		poolAssetIn.Token.Amount.ToDec(),
		normalizedWeight,
		p.GetTotalShares().ToDec(),
		shareOutAmount.ToDec(),
		p.GetSpreadFactor(ctx),
	).TruncateInt()

	if !tokenInAmount.IsPositive() {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrNotPositiveRequireAmount, nonPostiveTokenAmountErrFormat, tokenInAmount)
	}

	poolAssetIn.Token.Amount = poolAssetIn.Token.Amount.Add(tokenInAmount)
	err = p.UpdatePoolAssetBalance(poolAssetIn.Token)
	if err != nil {
		return sdk.Int{}, err
	}

	return tokenInAmount, nil
}

func (p *Pool) ExitSwapExactAmountOut(
	ctx sdk.Context,
	tokenOut sdk.Coin,
	shareInMaxAmount sdk.Int,
) (shareInAmount sdk.Int, err error) {
	_, poolAssetOut, err := p.getPoolAssetAndIndex(tokenOut.Denom)
	if err != nil {
		return sdk.Int{}, err
	}

	sharesIn := calcPoolSharesInGivenSingleAssetOut(
		poolAssetOut.Token.Amount.ToDec(),
		poolAssetOut.Weight.ToDec().Quo(p.TotalWeight.ToDec()),
		p.GetTotalShares().ToDec(),
		tokenOut.Amount.ToDec(),
		p.GetSpreadFactor(ctx),
		p.GetExitFee(ctx),
	).TruncateInt()

	if !sharesIn.IsPositive() {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrNotPositiveRequireAmount, nonPostiveSharesAmountErrFormat, sharesIn)
	}

	if sharesIn.GT(shareInMaxAmount) {
		return sdk.Int{}, errorsmod.Wrapf(types.ErrLimitMaxAmount, sharesLargerThanMaxErrFormat, sharesIn, shareInMaxAmount)
	}

	if err := p.exitPool(ctx, sdk.NewCoins(tokenOut), sharesIn); err != nil {
		return sdk.Int{}, err
	}

	return sharesIn, nil
}

func (p *Pool) AsSerializablePool() poolmanagertypes.PoolI {
	return p
}
