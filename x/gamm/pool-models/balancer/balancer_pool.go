package balancer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

// NewPool returns a weighted CPMM pool with the provided parameters, and initial assets.
// Invariants that are assumed to be satisfied and not checked:
// (This is handled in ValidateBasic)
// * 2 <= len(assets) <= 8
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewBalancerPool(poolId uint64, balancerPoolParams PoolParams, assets []types.PoolAsset, futureGovernor string, blockTime time.Time) (Pool, error) {
	poolAddr := types.NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address:            poolAddr.String(),
		Id:                 poolId,
		PoolParams:         PoolParams{},
		TotalWeight:        sdk.ZeroInt(),
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(poolId), sdk.ZeroInt()),
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

func (pa Pool) GetPoolSwapFee() sdk.Dec {
	return pa.PoolParams.SwapFee
}

func (pa Pool) GetPoolExitFee() sdk.Dec {
	return pa.PoolParams.ExitFee
}

func (pa Pool) GetPoolParams() PoolParams {
	return pa.PoolParams
}

func (pa Pool) GetTotalWeight() sdk.Int {
	return pa.TotalWeight
}

func (pa Pool) GetTotalShares() sdk.Coin {
	return pa.TotalShares
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
func (pa *Pool) setInitialPoolAssets(PoolAssets []types.PoolAsset) error {
	exists := make(map[string]bool)
	for _, asset := range pa.PoolAssets {
		exists[asset.Token.Denom] = true
	}

	newTotalWeight := pa.TotalWeight
	scaledPoolAssets := make([]types.PoolAsset, 0, len(PoolAssets))

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
		asset.Weight = asset.Weight.MulRaw(types.GuaranteedWeightPrecision)
		scaledPoolAssets = append(scaledPoolAssets, asset)
		newTotalWeight = newTotalWeight.Add(asset.Weight)
	}

	// TODO: Change this to a more efficient sorted insert algorithm.
	// Furthermore, consider changing the underlying data type to allow in-place modification if the
	// number of PoolAssets is expected to be large.
	pa.PoolAssets = append(pa.PoolAssets, scaledPoolAssets...)
	types.SortPoolAssetsByDenom(pa.PoolAssets)

	pa.TotalWeight = newTotalWeight

	return nil
}

// setInitialPoolParams
func (pa *Pool) setInitialPoolParams(params PoolParams, sortedAssets []types.PoolAsset, curBlockTime time.Time) error {
	pa.PoolParams = params
	if params.SmoothWeightChangeParams != nil {
		// set initial assets
		initialWeights := make([]types.PoolAsset, len(sortedAssets))
		for i, v := range sortedAssets {
			initialWeights[i] = types.PoolAsset{
				Weight: v.Weight,
				Token:  sdk.Coin{Denom: v.Token.Denom, Amount: sdk.ZeroInt()},
			}
		}
		params.SmoothWeightChangeParams.InitialPoolWeights = initialWeights

		// sort target weights by denom
		targetPoolWeights := params.SmoothWeightChangeParams.TargetPoolWeights
		types.SortPoolAssetsByDenom(targetPoolWeights)

		// scale target pool weights by GuaranteedWeightPrecision
		for i, v := range targetPoolWeights {
			err := types.ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
			pa.PoolParams.SmoothWeightChangeParams.TargetPoolWeights[i] = types.PoolAsset{
				Weight: v.Weight.MulRaw(types.GuaranteedWeightPrecision),
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
func (pa Pool) GetPoolAsset(denom string) (types.PoolAsset, error) {
	_, asset, err := pa.getPoolAssetAndIndex(denom)
	return asset, err
}

// Returns a pool asset, and its index. If err != nil, then the index will be valid.
func (pa Pool) getPoolAssetAndIndex(denom string) (int, types.PoolAsset, error) {
	if denom == "" {
		return -1, types.PoolAsset{}, fmt.Errorf("you tried to find the PoolAsset with empty denom")
	}

	if len(pa.PoolAssets) == 0 {
		return -1, types.PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
	}

	i := sort.Search(len(pa.PoolAssets), func(i int) bool {
		PoolAssetA := pa.PoolAssets[i]

		compare := strings.Compare(PoolAssetA.Token.Denom, denom)
		return compare >= 0
	})

	if i < 0 || i >= len(pa.PoolAssets) {
		return -1, types.PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
	}

	if pa.PoolAssets[i].Token.Denom != denom {
		return -1, types.PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
	}

	return i, pa.PoolAssets[i], nil
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

func (pa Pool) GetPoolAssets(denoms ...string) ([]types.PoolAsset, error) {
	result := make([]types.PoolAsset, 0, len(denoms))

	for _, denom := range denoms {
		PoolAsset, err := pa.GetPoolAsset(denom)
		if err != nil {
			return nil, err
		}

		result = append(result, PoolAsset)
	}

	return result, nil
}

func (pa Pool) GetAllPoolAssets() []types.PoolAsset {
	copyslice := make([]types.PoolAsset, len(pa.PoolAssets))
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
func (pa *Pool) updateAllWeights(newWeights []types.PoolAsset) {
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

// PokeTokenWeights checks to see if the pool's token weights need to be updated,
// and if so, does so.
func (pa *Pool) PokeTokenWeights(blockTime time.Time) {
	// Pool weights aren't changing, do nothing.
	poolWeightsChanging := (pa.PoolParams.SmoothWeightChangeParams != nil)
	if !poolWeightsChanging {
		return
	}
	// Pool weights are changing.
	// TODO: Add intra-block cache check that we haven't already poked
	// the block yet.
	params := *pa.PoolParams.SmoothWeightChangeParams

	// the weights w(t) for the pool at time `t` is the following:
	//   t <= start_time: w(t) = initial_pool_weights
	//   start_time < t <= start_time + duration:
	//     w(t) = initial_pool_weights + (t - start_time) *
	//       (target_pool_weights - initial_pool_weights) / (duration)
	//   t > start_time + duration: w(t) = target_pool_weights

	// t <= StartTime
	if blockTime.Before(params.StartTime) || params.StartTime.Equal(blockTime) {
		// Do nothing
		return
	} else if blockTime.After(params.StartTime.Add(params.Duration)) {
		// t > start_time + duration
		// Update weights to be the target weights.
		// TODO: When we add support for adding new assets via this method,
		// 		 Ensure the new asset has some token sent with it.
		pa.updateAllWeights(params.TargetPoolWeights)
		// We've finished updating weights, so delete this parameter
		pa.PoolParams.SmoothWeightChangeParams = nil
		return
	} else {
		//	w(t) = initial_pool_weights + (t - start_time) *
		//       (target_pool_weights - initial_pool_weights) / (duration)
		// We first compute percent duration elapsed = (t - start_time) / duration, via Unix time.
		shiftedBlockTime := blockTime.Sub(params.StartTime).Milliseconds()
		percentDurationElapsed := sdk.NewDec(shiftedBlockTime).QuoInt64(params.Duration.Milliseconds())
		// If the duration elapsed is equal to the total time,
		// or a rounding error makes it seem like it is, just set to target weight
		if percentDurationElapsed.GTE(sdk.OneDec()) {
			pa.updateAllWeights(params.TargetPoolWeights)
			return
		}
		totalWeightsDiff := subPoolAssetWeights(params.TargetPoolWeights, params.InitialPoolWeights)
		// Below will be auto-truncated according to internal weight precision routine.
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

func (pa Pool) IsActive(curBlockTime time.Time) bool {

	// Add frozen pool checking, etc...

	return true
}

func (params PoolParams) Validate(poolWeights []types.PoolAsset) error {
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
			err := types.ValidateUserSpecifiedWeight(v.Weight)
			if err != nil {
				return err
			}
		}
		// Ensure that all the target weight denoms are same as pool asset weights
		sortedTargetPoolWeights := types.SortPoolAssetsOutOfPlaceByDenom(targetWeights)
		sortedPoolWeights := types.SortPoolAssetsOutOfPlaceByDenom(poolWeights)
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

func ValidateFutureGovernor(governor string) error {
	// allow empty governor
	if governor == "" {
		return nil
	}

	// validation for future owner
	// "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
	_, err := sdk.AccAddressFromBech32(governor)
	if err == nil {
		return nil
	}

	lockTimeStr := ""
	splits := strings.Split(governor, ",")
	if len(splits) > 2 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
	}

	// token,100h
	if len(splits) == 2 {
		lpTokenStr := splits[0]
		if sdk.ValidateDenom(lpTokenStr) != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
		}
		lockTimeStr = splits[1]
	}

	// 100h
	if len(splits) == 1 {
		lockTimeStr = splits[0]
	}

	// Note that a duration of 0 is allowed
	_, err = time.ParseDuration(lockTimeStr)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", governor))
	}
	return nil
}

// subPoolAssetWeights subtracts the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can (and probably will have some) be negative.
func subPoolAssetWeights(base []types.PoolAsset, other []types.PoolAsset) []types.PoolAsset {
	weightDifference := make([]types.PoolAsset, len(base))
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
		weightDifference[i] = types.PoolAsset{Token: asset.Token, Weight: curWeightDiff}
	}
	return weightDifference
}

// addPoolAssetWeights adds the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can be negative.
func addPoolAssetWeights(base []types.PoolAsset, other []types.PoolAsset) []types.PoolAsset {
	weightSum := make([]types.PoolAsset, len(base))
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
		weightSum[i] = types.PoolAsset{Token: asset.Token, Weight: curWeightSum}
	}
	return weightSum
}

// assumes 0 < d < 1
func poolAssetsMulDec(base []types.PoolAsset, d sdk.Dec) []types.PoolAsset {
	newWeights := make([]types.PoolAsset, len(base))
	for i, asset := range base {
		// TODO: This can adversarially panic at the moment! (as can Pool.TotalWeight)
		// Ensure this won't be able to panic in the future PR where we bound
		// each assets weight, and add precision
		newWeight := d.MulInt(asset.Weight).RoundInt()
		newWeights[i] = types.PoolAsset{Token: asset.Token, Weight: newWeight}
	}
	return newWeights
}
