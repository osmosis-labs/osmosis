package balancer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

const (
	errMsgFormatNoPoolAssetFound = "can't find the PoolAsset (%s)"
)

var (
	_ types.PoolI                  = &Pool{}
	_ types.PoolAmountOutExtension = &Pool{}
)

// NewPool returns a weighted CPMM pool with the provided parameters, and initial assets.
// Invariants that are assumed to be satisfied and not checked:
// (This is handled in ValidateBasic)
// * 2 <= len(assets) <= 8
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewBalancerPool(poolId uint64, balancerPoolParams PoolParams, assets []PoolAsset, futureGovernor string, blockTime time.Time) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)

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

	err := pool.setInitialPoolAssets(assets)
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
func (pa Pool) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		panic(fmt.Sprintf("could not bech32 decode address of pool with id: %d", pa.GetId()))
	}
	return addr
}

func (pa Pool) GetId() uint64 {
	return pa.Id
}

func (pa Pool) GetSwapFee(_ sdk.Context) sdk.Dec {
	return pa.PoolParams.SwapFee
}

func (pa Pool) GetTotalPoolLiquidity(_ sdk.Context) sdk.Coins {
	return PoolAssetsCoins(pa.PoolAssets)
}

func (pa Pool) GetExitFee(_ sdk.Context) sdk.Dec {
	return pa.PoolParams.ExitFee
}

func (pa Pool) GetPoolParams() PoolParams {
	return pa.PoolParams
}

func (pa Pool) GetTotalWeight() sdk.Int {
	return pa.TotalWeight
}

func (pa Pool) GetTotalShares() sdk.Int {
	return pa.TotalShares.Amount
}

func (pa *Pool) AddTotalShares(amt sdk.Int) {
	pa.TotalShares.Amount = pa.TotalShares.Amount.Add(amt)
}

func (pa *Pool) SubTotalShares(amt sdk.Int) {
	pa.TotalShares.Amount = pa.TotalShares.Amount.Sub(amt)
}

// setInitialPoolAssets sets the PoolAssets in the pool.
// It is only designed to be called at the pool's creation.
// If the same denom's PoolAsset exists, will return error.
// The list of PoolAssets must be sorted. This is done to enable fast searching for a PoolAsset by denomination.
// TODO: Unify story for validation of []PoolAsset, some is here, some is in CreatePool.ValidateBasic()
func (pa *Pool) setInitialPoolAssets(PoolAssets []PoolAsset) error {
	exists := make(map[string]bool)
	for _, asset := range pa.PoolAssets {
		exists[asset.Token.Denom] = true
	}

	newTotalWeight := pa.TotalWeight
	scaledPoolAssets := make([]PoolAsset, 0, len(PoolAssets))

	// TODO: Refactor this into PoolAsset.validate()
	for _, asset := range PoolAssets {
		if asset.Token.Amount.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't add the zero or negative balance of token")
		}

		err := asset.ValidateWeight()
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
	pa.PoolAssets = append(pa.PoolAssets, scaledPoolAssets...)
	SortPoolAssetsByDenom(pa.PoolAssets)

	pa.TotalWeight = newTotalWeight

	return nil
}

// ValidateUserSpecifiedWeight ensures that a weight that is provided from user-input anywhere
// for creating a pool obeys the expected guarantees.
// Namely, that the weight is in the range [1, MaxUserSpecifiedWeight)
func ValidateUserSpecifiedWeight(weight sdk.Int) error {
	if !weight.IsPositive() {
		return sdkerrors.Wrap(types.ErrNotPositiveWeight, weight.String())
	}

	if weight.GTE(MaxUserSpecifiedWeight) {
		return sdkerrors.Wrap(types.ErrWeightTooLarge, weight.String())
	}
	return nil
}

// setInitialPoolParams
func (pa *Pool) setInitialPoolParams(params PoolParams, sortedAssets []PoolAsset, curBlockTime time.Time) error {
	pa.PoolParams = params
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
		SortPoolAssetsByDenom(targetPoolWeights)

		// scale target pool weights by GuaranteedWeightPrecision
		for i, v := range targetPoolWeights {
			err := ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
			pa.PoolParams.SmoothWeightChangeParams.TargetPoolWeights[i] = PoolAsset{
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
func (pa Pool) GetPoolAsset(denom string) (PoolAsset, error) {
	_, asset, err := pa.getPoolAssetAndIndex(denom)
	return asset, err
}

// Returns a pool asset, and its index. If err != nil, then the index will be valid.
func (pa Pool) getPoolAssetAndIndex(denom string) (int, PoolAsset, error) {
	if denom == "" {
		return -1, PoolAsset{}, fmt.Errorf("you tried to find the PoolAsset with empty denom")
	}

	if len(pa.PoolAssets) == 0 {
		return -1, PoolAsset{}, sdkerrors.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(errMsgFormatNoPoolAssetFound, denom))
	}

	i := sort.Search(len(pa.PoolAssets), func(i int) bool {
		PoolAssetA := pa.PoolAssets[i]

		compare := strings.Compare(PoolAssetA.Token.Denom, denom)
		return compare >= 0
	})

	if i < 0 || i >= len(pa.PoolAssets) {
		return -1, PoolAsset{}, sdkerrors.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(errMsgFormatNoPoolAssetFound, denom))
	}

	if pa.PoolAssets[i].Token.Denom != denom {
		return -1, PoolAsset{}, sdkerrors.Wrapf(types.ErrDenomNotFoundInPool, fmt.Sprintf(errMsgFormatNoPoolAssetFound, denom))
	}

	return i, pa.PoolAssets[i], nil
}

func (p Pool) parsePoolAssetsByDenoms(tokenADenom, tokenBDenom string) (
	Aasset PoolAsset, Basset PoolAsset, err error,
) {
	Aasset, found1 := GetPoolAssetByDenom(p.PoolAssets, tokenADenom)
	Basset, found2 := GetPoolAssetByDenom(p.PoolAssets, tokenBDenom)
	if !(found1 && found2) {
		return Aasset, Basset, errors.New("one of the provided pool denoms does not exist in pool")
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
	return tokensA[0], Aasset, Basset, err
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

func (pa *Pool) UpdatePoolAssetBalance(coin sdk.Coin) error {
	// Check that PoolAsset exists.
	assetIndex, existingAsset, err := pa.getPoolAssetAndIndex(coin.Denom)
	if err != nil {
		return err
	}

	if coin.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("can't set the pool's balance of a token to be zero or negative")
	}

	// Update the supply of the asset
	existingAsset.Token = coin
	pa.PoolAssets[assetIndex] = existingAsset
	return nil
}

func (pa *Pool) UpdatePoolAssetBalances(coins sdk.Coins) error {
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
		err = pa.UpdatePoolAssetBalance(coin)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pa *Pool) addToPoolAssetBalances(coins sdk.Coins) error {
	for _, coin := range coins {
		i, poolAsset, err := pa.getPoolAssetAndIndex(coin.Denom)
		if err != nil {
			return err
		}
		poolAsset.Token.Amount = poolAsset.Token.Amount.Add(coin.Amount)
		pa.PoolAssets[i] = poolAsset
	}
	return nil
}

func (pa Pool) GetPoolAssets(denoms ...string) ([]PoolAsset, error) {
	result := make([]PoolAsset, 0, len(denoms))

	for _, denom := range denoms {
		PoolAsset, err := pa.GetPoolAsset(denom)
		if err != nil {
			return nil, err
		}

		result = append(result, PoolAsset)
	}

	return result, nil
}

func (pa Pool) GetAllPoolAssets() []PoolAsset {
	copyslice := make([]PoolAsset, len(pa.PoolAssets))
	copy(copyslice, pa.PoolAssets)
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
func (pa *Pool) updateAllWeights(newWeights []PoolAsset) {
	if len(pa.PoolAssets) != len(newWeights) {
		panic("updateAllWeights called with invalid input, len(newWeights) != len(existingWeights)")
	}
	totalWeight := sdk.ZeroInt()
	for i, asset := range pa.PoolAssets {
		if asset.Token.Denom != newWeights[i].Token.Denom {
			panic(fmt.Sprintf("updateAllWeights called with invalid input, "+
				"expected new weights' %vth asset to be %v, got %v",
				i, asset.Token.Denom, newWeights[i].Token.Denom))
		}
		err := newWeights[i].ValidateWeight()
		if err != nil {
			panic("updateAllWeights: Tried to set an invalid weight")
		}
		pa.PoolAssets[i].Weight = newWeights[i].Weight
		totalWeight = totalWeight.Add(pa.PoolAssets[i].Weight)
	}
	pa.TotalWeight = totalWeight
}

// PokePool checks to see if the pool's token weights need to be updated, and
// if so, does so.
func (pa *Pool) PokePool(blockTime time.Time) {
	// check if pool weights didn't change
	poolWeightsChanging := pa.PoolParams.SmoothWeightChangeParams != nil
	if !poolWeightsChanging {
		return
	}

	params := *pa.PoolParams.SmoothWeightChangeParams

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
		pa.updateAllWeights(params.TargetPoolWeights)

		// we've finished updating the weights, so reset the following fields
		pa.PoolParams.SmoothWeightChangeParams = nil
		return

	default:
		// case 3: t > start_time + duration: w(t) = target_pool_weights

		shiftedBlockTime := blockTime.Sub(params.StartTime).Milliseconds()
		percentDurationElapsed := sdk.NewDec(shiftedBlockTime).QuoInt64(params.Duration.Milliseconds())

		// If the duration elapsed is equal to the total time, or a rounding error
		// makes it seem like it is, just set to target weight.
		if percentDurationElapsed.GTE(sdk.OneDec()) {
			pa.updateAllWeights(params.TargetPoolWeights)
			return
		}

		// below will be auto-truncated according to internal weight precision routine
		totalWeightsDiff := subPoolAssetWeights(params.TargetPoolWeights, params.InitialPoolWeights)
		scaledDiff := poolAssetsMulDec(totalWeightsDiff, percentDurationElapsed)
		updatedWeights := addPoolAssetWeights(params.InitialPoolWeights, scaledDiff)

		pa.updateAllWeights(updatedWeights)
	}
}

func (pa Pool) GetTokenWeight(denom string) (sdk.Int, error) {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Weight, nil
}

func (pa Pool) GetTokenBalance(denom string) (sdk.Int, error) {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Token.Amount, nil
}

func (pa Pool) NumAssets() int {
	return len(pa.PoolAssets)
}

func (pa Pool) IsActive(ctx sdk.Context) bool {
	return true
}

func NewPoolParams(swapFee, exitFee sdk.Dec, params *SmoothWeightChangeParams) PoolParams {
	return PoolParams{
		SwapFee:                  swapFee,
		ExitFee:                  exitFee,
		SmoothWeightChangeParams: params,
	}
}

func (params PoolParams) Validate(poolWeights []PoolAsset) error {
	if params.ExitFee.IsNegative() {
		return types.ErrNegativeExitFee
	}

	if params.ExitFee.GTE(sdk.OneDec()) {
		return types.ErrTooMuchExitFee
	}

	if params.SwapFee.IsNegative() {
		return types.ErrNegativeSwapFee
	}

	if params.SwapFee.GTE(sdk.OneDec()) {
		return types.ErrTooMuchSwapFee
	}

	if params.SmoothWeightChangeParams != nil {
		targetWeights := params.SmoothWeightChangeParams.TargetPoolWeights
		// Ensure it has the right number of weights
		if len(targetWeights) != len(poolWeights) {
			return types.ErrPoolParamsInvalidNumDenoms
		}
		// Validate all user specified weights
		for _, v := range targetWeights {
			err := ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
		}
		// Ensure that all the target weight denoms are same as pool asset weights
		sortedTargetPoolWeights := SortPoolAssetsOutOfPlaceByDenom(targetWeights)
		sortedPoolWeights := SortPoolAssetsOutOfPlaceByDenom(poolWeights)
		for i, v := range sortedPoolWeights {
			if sortedTargetPoolWeights[i].Token.Denom != v.Token.Denom {
				return types.ErrPoolParamsInvalidDenom
			}
		}

		// No start time validation needed

		// We do not need to validate InitialPoolWeights, as we set that ourselves
		// in setInitialPoolParams

		// TODO: Is there anything else we can validate for duration?
		if params.SmoothWeightChangeParams.Duration <= 0 {
			return errors.New("params.SmoothWeightChangeParams must have a positive duration")
		}
	}

	return nil
}

func (params PoolParams) GetPoolSwapFee() sdk.Dec {
	return params.SwapFee
}

func (params PoolParams) GetPoolExitFee() sdk.Dec {
	return params.ExitFee
}

// subPoolAssetWeights subtracts the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can (and probably will have some) be negative.
func subPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightDifference := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("subPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("subPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightDiff := asset.Weight.Sub(other[i].Weight)
		weightDifference[i] = PoolAsset{Token: asset.Token, Weight: curWeightDiff}
	}
	return weightDifference
}

// addPoolAssetWeights adds the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can be negative.
func addPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightSum := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("addPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("addPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightSum := asset.Weight.Add(other[i].Weight)
		weightSum[i] = PoolAsset{Token: asset.Token, Weight: curWeightSum}
	}
	return weightSum
}

// assumes 0 < d < 1
func poolAssetsMulDec(base []PoolAsset, d sdk.Dec) []PoolAsset {
	newWeights := make([]PoolAsset, len(base))
	for i, asset := range base {
		// TODO: This can adversarially panic at the moment! (as can Pool.TotalWeight)
		// Ensure this won't be able to panic in the future PR where we bound
		// each assets weight, and add precision
		newWeight := d.MulInt(asset.Weight).RoundInt()
		newWeights[i] = PoolAsset{Token: asset.Token, Weight: newWeight}
	}
	return newWeights
}
