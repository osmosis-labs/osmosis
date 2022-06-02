package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v3/v043_temp/address"
)

// PoolI defines an interface for pools that hold tokens.
type PoolI interface {
	proto.Message

	GetAddress() sdk.AccAddress
	String() string

	GetId() uint64
	GetPoolParams() PoolParams
	GetTotalWeight() sdk.Int
	GetTotalShares() sdk.Coin
	AddTotalShares(amt sdk.Int)
	SubTotalShares(amt sdk.Int)
	GetPoolAsset(denom string) (PoolAsset, error)
	// UpdatePoolAssetBalance updates the balances for
	// the token with denomination coin.denom
	UpdatePoolAssetBalance(coin sdk.Coin) error
	// UpdatePoolAssetBalances calls UpdatePoolAssetBalance
	// on each constituent coin.
	UpdatePoolAssetBalances(coins sdk.Coins) error
	GetPoolAssets(denoms ...string) ([]PoolAsset, error)
	GetAllPoolAssets() []PoolAsset
	PokeTokenWeights(blockTime time.Time)
	GetTokenWeight(denom string) (sdk.Int, error)
	GetTokenBalance(denom string) (sdk.Int, error)
	NumAssets() int
	IsActive(curBlockTime time.Time) bool
}

var (
	_                         PoolI   = (*Pool)(nil)
	MaxUserSpecifiedWeight    sdk.Int = sdk.NewIntFromUint64(1 << 20)
	GuaranteedWeightPrecision int64   = 1 << 30
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	key := append([]byte("pool"), sdk.Uint64ToBigEndian(poolId)...)
	return address.Module(ModuleName, key)
}

// NewPool returns a weighted CPMM pool with the provided parameters, and initial assets.
// Invariants that are assumed to be satisfied and not checked:
// (This is handled in ValidateBasic)
// * 2 <= len(assets) <= 8
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewPool(poolId uint64, poolParams PoolParams, assets []PoolAsset, futureGovernor string, blockTime time.Time) (PoolI, error) {
	poolAddr := NewPoolAddress(poolId)

	// pool thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	pool := &Pool{
		Address:            poolAddr.String(),
		Id:                 poolId,
		PoolParams:         PoolParams{},
		TotalWeight:        sdk.ZeroInt(),
		TotalShares:        sdk.NewCoin(GetPoolShareDenom(poolId), sdk.ZeroInt()),
		PoolAssets:         nil,
		FuturePoolGovernor: futureGovernor,
	}

	err := pool.setInitialPoolAssets(assets)
	if err != nil {
		return &Pool{}, err
	}

	sortedPoolAssets := pool.GetAllPoolAssets()
	err = poolParams.Validate(sortedPoolAssets)
	if err != nil {
		return &Pool{}, err
	}

	err = pool.setInitialPoolParams(poolParams, sortedPoolAssets, blockTime)
	if err != nil {
		return &Pool{}, err
	}

	return pool, nil
}

func (params PoolParams) Validate(poolWeights []PoolAsset) error {
	if params.ExitFee.IsNegative() {
		return ErrNegativeExitFee
	}

	if params.ExitFee.GTE(sdk.OneDec()) {
		return ErrTooMuchExitFee
	}

	if params.SwapFee.IsNegative() {
		return ErrNegativeSwapFee
	}

	if params.SwapFee.GTE(sdk.OneDec()) {
		return ErrTooMuchSwapFee
	}

	if params.SmoothWeightChangeParams != nil {
		targetWeights := params.SmoothWeightChangeParams.TargetPoolWeights
		// Ensure it has the right number of weights
		if len(targetWeights) != len(poolWeights) {
			return ErrPoolParamsInvalidNumDenoms
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
				return ErrPoolParamsInvalidDenom
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
		return -1, PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
	}

	i := sort.Search(len(pa.PoolAssets), func(i int) bool {
		PoolAssetA := pa.PoolAssets[i]

		compare := strings.Compare(PoolAssetA.Token.Denom, denom)
		return compare >= 0
	})

	if i < 0 || i >= len(pa.PoolAssets) {
		return -1, PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
	}

	if pa.PoolAssets[i].Token.Denom != denom {
		return -1, PoolAsset{}, fmt.Errorf("can't find the PoolAsset (%s)", denom)
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
