package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// PoolAccountI defines an account interface for pools that hold tokens.
type PoolAccountI interface {
	authtypes.AccountI

	GetId() uint64
	GetPoolParams() PoolParams
	GetTotalWeight() sdk.Int
	GetTotalShare() sdk.Coin
	AddTotalShare(amt sdk.Int)
	SubTotalShare(amt sdk.Int)
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
}

var (
	// TODO: Add `GenesisAccount` type
	_                         PoolAccountI = (*PoolAccount)(nil)
	MaxUserSpecifiedWeight    sdk.Int      = sdk.NewIntFromUint64(1 << 20)
	GuaranteedWeightPrecision int64        = 1 << 30
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash(append(PoolAddressPrefix, sdk.Uint64ToBigEndian(poolId)...)))
}

// NewPoolAccount returns a Balancer pool with the provided parameters, and initial assets.
// Invariants that are assumed to be satisfied and not checked:
// TODO: Why don't we check these in here?
// TODO: Why does this panic, not just return an error?
// * 2 <= len(assets) <= 8
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewPoolAccount(poolId uint64, poolParams PoolParams, assets []PoolAsset, futureGovernor string) (PoolAccountI, error) {
	poolAddr := NewPoolAddress(poolId)
	baseAcc := authtypes.NewBaseAccountWithAddress(poolAddr)

	// pool account thats created up to ensuring the assets and params are valid.
	// We assume that FuturePoolGovernor is valid.
	protoPoolAcc := &PoolAccount{
		BaseAccount:        baseAcc,
		Id:                 poolId,
		PoolParams:         poolParams,
		TotalWeight:        sdk.ZeroInt(),
		TotalShare:         sdk.NewCoin(GetPoolShareDenom(poolId), sdk.ZeroInt()),
		PoolAssets:         nil,
		FuturePoolGovernor: futureGovernor,
	}

	err := protoPoolAcc.setInitialPoolAssets(assets)
	if err != nil {
		return &PoolAccount{}, err
	}

	err = poolParams.Validate(protoPoolAcc.GetAllPoolAssets())
	if err != nil {
		return &PoolAccount{}, err
	}

	return protoPoolAcc, nil
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
		// TODO: We need to test that TargetPoolWeights only contains the same denoms as
		// the provided PoolWeights
		// for _, v := range params.SmoothWeightChangeParams.TargetPoolWeights {
		// 	err := ValidateUserSpecifiedWeight(v.Weight)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// TODO: Validate duration & start time
		// We do not need to validate InitialPoolWeights, as we should be setting that ourselves
		// TODO: Set that in create new pool.
	}

	return nil
}

func (pa PoolAccount) GetId() uint64 {
	return pa.Id
}

func (pa PoolAccount) GetPoolParams() PoolParams {
	return pa.PoolParams
}

func (pa PoolAccount) GetTotalWeight() sdk.Int {
	return pa.TotalWeight
}

func (pa PoolAccount) GetTotalShare() sdk.Coin {
	return pa.TotalShare
}

func (pa *PoolAccount) AddTotalShare(amt sdk.Int) {
	pa.TotalShare.Amount = pa.TotalShare.Amount.Add(amt)
}

func (pa *PoolAccount) SubTotalShare(amt sdk.Int) {
	pa.TotalShare.Amount = pa.TotalShare.Amount.Sub(amt)
}

// setInitialPoolAssets sets the PoolAssets in the pool.
// It is only designed to be called at the pool's creation.
// If the same denom's PoolAsset exists, will return error.
// The list of PoolAssets must be sorted. This is done to enable fast searching for a PoolAsset by denomination.
// TODO: Unify story for validation of []PoolAsset
func (pa *PoolAccount) setInitialPoolAssets(PoolAssets []PoolAsset) error {
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
	sort.Slice(pa.PoolAssets, func(i, j int) bool {
		PoolAssetA := pa.PoolAssets[i]
		PoolAssetB := pa.PoolAssets[j]

		return strings.Compare(PoolAssetA.Token.Denom, PoolAssetB.Token.Denom) == -1
	})

	pa.TotalWeight = newTotalWeight

	return nil
}

// GetPoolAssets returns the denom's PoolAsset, If the PoolAsset doesn't exist, will return error.
// As above, it will search the denom's PoolAsset by using binary search.
// So, it is important to make sure that the PoolAssets are sorted.
func (pa PoolAccount) GetPoolAsset(denom string) (PoolAsset, error) {
	_, asset, err := pa.getPoolAssetAndIndex(denom)
	return asset, err
}

// Returns a pool asset, and its index. If err != nil, then the index will be valid.
func (pa PoolAccount) getPoolAssetAndIndex(denom string) (int, PoolAsset, error) {
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

func (pa *PoolAccount) UpdatePoolAssetBalance(coin sdk.Coin) error {
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

func (pa *PoolAccount) UpdatePoolAssetBalances(coins sdk.Coins) error {
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

func (pa PoolAccount) GetPoolAssets(denoms ...string) ([]PoolAsset, error) {
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

func (pa PoolAccount) GetAllPoolAssets() []PoolAsset {
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
func (pa PoolAccount) updateAllWeights(newWeights []PoolAsset) {
	if len(pa.PoolAssets) != len(newWeights) {
		panic("updateAllWeights called with invalid input, len(newWeights) != len(existingWeights)")
	}
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
	}
}

// PokeTokenWeights checks to see if the pool's token weights need to be updated,
// and if so, does so.
func (pa PoolAccount) PokeTokenWeights(blockTime time.Time) {
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
		// TODO: This line doesn't work, since this is a non-pointer receiever,
		// and pa.PoolParams gets copied.
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

	return
}

func (pa PoolAccount) GetTokenWeight(denom string) (sdk.Int, error) {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Weight, nil
}

func (pa PoolAccount) GetTokenBalance(denom string) (sdk.Int, error) {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Token.Amount, nil
}

func (pa PoolAccount) NumAssets() int {
	return len(pa.PoolAssets)
}

// SetPubKey - Implements AccountI
func (pa PoolAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for pool accounts")
}

// SetSequence - Implements AccountI
func (pa PoolAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for pool accounts")
}

type poolAccountPretty struct {
	Address            sdk.AccAddress `json:"address" yaml:"address"`
	PubKey             string         `json:"public_key" yaml:"public_key"`
	AccountNumber      uint64         `json:"account_number" yaml:"account_number"`
	Sequence           uint64         `json:"sequence" yaml:"sequence"`
	Id                 uint64         `json:"id" yaml:"id"`
	PoolParams         PoolParams     `json:"pool_params" yaml:"pool_params"`
	FuturePoolGovernor string         `json:"future_pool_governor" yaml:"future_pool_governor"`
	TotalWeight        sdk.Int        `json:"total_weight" yaml:"total_weight"`
	TotalShare         sdk.Coin       `json:"total_share" yaml:"total_share"`
	PoolAssets         []PoolAsset    `json:"pool_assets" yaml:"pool_assets"`
}

func (pa PoolAccount) String() string {
	out, _ := pa.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a PoolAccount.
func (pa PoolAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	bs, err := yaml.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        pa.TotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         pa.PoolAssets,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}

// MarshalJSON returns the JSON representation of a PoolAccount.
func (pa PoolAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(pa.Address)
	if err != nil {
		return nil, err
	}

	return json.Marshal(poolAccountPretty{
		Address:            accAddr,
		PubKey:             "",
		AccountNumber:      pa.AccountNumber,
		Id:                 pa.Id,
		PoolParams:         pa.PoolParams,
		FuturePoolGovernor: pa.FuturePoolGovernor,
		TotalWeight:        pa.TotalWeight,
		TotalShare:         pa.TotalShare,
		PoolAssets:         pa.PoolAssets,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a PoolAccount.
func (pa *PoolAccount) UnmarshalJSON(bz []byte) error {
	var alias poolAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	pa.BaseAccount = authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	pa.Id = alias.Id
	pa.PoolParams = alias.PoolParams
	pa.FuturePoolGovernor = alias.FuturePoolGovernor
	pa.TotalWeight = alias.TotalWeight
	pa.TotalShare = alias.TotalShare
	pa.PoolAssets = alias.PoolAssets

	return nil
}
