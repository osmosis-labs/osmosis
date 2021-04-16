package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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
	SetPoolParams(params PoolParams)
	GetTotalWeight() sdk.Int
	GetTotalShare() sdk.Coin
	AddTotalShare(amt sdk.Int)
	SubTotalShare(amt sdk.Int)
	AddPoolAssets(PoolAssets []PoolAsset) error
	GetPoolAsset(denom string) (PoolAsset, error)
	// TODO: Rename this function, as it expects the asset to already exist
	SetPoolAsset(denom string, asset PoolAsset) error
	GetPoolAssets(denoms ...string) ([]PoolAsset, error)
	SetPoolAssets(assets []PoolAsset) error
	GetAllPoolAssets() []PoolAsset
	SetTokenWeight(denom string, weight sdk.Int) error
	GetTokenWeight(denom string) (sdk.Int, error)
	SetTokenBalance(denom string, amount sdk.Int) error
	GetTokenBalance(denom string) (sdk.Int, error)
	NumAssets() int
}

var (
	// TODO: Add `GenesisAccount` type
	_ PoolAccountI = (*PoolAccount)(nil)
)

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash(append(PoolAddressPrefix, sdk.Uint64ToBigEndian(poolId)...)))
}

func NewPoolAccount(poolId uint64, poolParams PoolParams, futureGovernor string) PoolAccountI {
	poolAddr := NewPoolAddress(poolId)
	baseAcc := authtypes.NewBaseAccountWithAddress(poolAddr)

	err := poolParams.Validate()
	if err != nil {
		panic(err)
	}

	return &PoolAccount{
		BaseAccount:        baseAcc,
		Id:                 poolId,
		PoolParams:         poolParams,
		TotalWeight:        sdk.ZeroInt(),
		TotalShare:         sdk.NewCoin(fmt.Sprintf("osmosis/pool/%d", poolId), sdk.ZeroInt()),
		PoolAssets:         nil,
		FuturePoolGovernor: futureGovernor,
	}
}

func (params PoolParams) Validate() error {
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

	return nil
}

func (pa PoolAccount) GetId() uint64 {
	return pa.Id
}

func (pa PoolAccount) GetPoolParams() PoolParams {
	return pa.PoolParams
}

func (pa *PoolAccount) SetPoolParams(params PoolParams) {
	pa.PoolParams = params
	return
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

// AddPoolAssets adds the PoolAssets to the pool. If the same denom's PoolAsset exists, will return error.
// The list of PoolAssets must be sorted. This is done to enable fast searching for a PoolAsset by denomination.
func (pa *PoolAccount) AddPoolAssets(PoolAssets []PoolAsset) error {
	exists := make(map[string]bool)
	for _, asset := range pa.PoolAssets {
		exists[asset.Token.Denom] = true
	}

	newTotalWeight := pa.TotalWeight

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

		newTotalWeight = newTotalWeight.Add(asset.Weight)
	}

	// TODO: Change this to a more efficient sorted insert algorithm.
	// Furthermore, consider changing the underlying data type to allow im-place modification if the
	// number of PoolAssets is expected to be large.
	pa.PoolAssets = append(pa.PoolAssets, PoolAssets...)
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

func (pa *PoolAccount) SetPoolAsset(denom string, asset PoolAsset) error {
	// Check that PoolAsset exists.
	assetIndex, existingAsset, err := pa.getPoolAssetAndIndex(denom)
	if err != nil {
		return err
	}

	if asset.Token.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("can't add the zero or negative balance of token")
	}

	err = asset.ValidateWeight()
	if err != nil {
		return err
	}

	// Update the total weight in the pool
	weightDifference := asset.Weight.Sub(existingAsset.Weight)
	pa.TotalWeight = pa.TotalWeight.Add(weightDifference)
	pa.PoolAssets[assetIndex] = asset
	return nil
}

func (pa *PoolAccount) SetPoolAssets(assets []PoolAsset) error {
	exists := make(map[string]int)
	for index, asset := range pa.PoolAssets {
		exists[asset.Token.Denom] = index
	}

	addingPoolAssetsExists := make(map[string]bool)

	deltaTotalWeight := sdk.ZeroInt()

	for _, asset := range assets {
		if asset.Token.Amount.LTE(sdk.ZeroInt()) {
			return fmt.Errorf("can't have an asset in the pool with no reserve supply.")
		}

		err := asset.ValidateWeight()
		if err != nil {
			return err
		}

		index, ok := exists[asset.Token.Denom]
		if !ok {
			return fmt.Errorf("PoolAsset doesn't exists")
		}

		if addingPoolAssetsExists[asset.Token.Denom] {
			return fmt.Errorf("adding PoolAssets duplicated")
		}
		addingPoolAssetsExists[asset.Token.Denom] = true

		oldPoolAsset := pa.PoolAssets[index]
		deltaTotalWeight = deltaTotalWeight.Add(asset.Weight.Sub(oldPoolAsset.Weight))

		pa.PoolAssets[index] = asset
	}

	pa.TotalWeight = pa.TotalWeight.Add(deltaTotalWeight)

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

func (pa *PoolAccount) SetTokenWeight(denom string, weight sdk.Int) error {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return err
	}

	PoolAsset.Weight = weight

	return pa.SetPoolAsset(denom, PoolAsset)
}

func (pa PoolAccount) GetTokenWeight(denom string) (sdk.Int, error) {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return sdk.Int{}, err
	}

	return PoolAsset.Weight, nil
}

func (pa *PoolAccount) SetTokenBalance(denom string, amount sdk.Int) error {
	PoolAsset, err := pa.GetPoolAsset(denom)
	if err != nil {
		return err
	}

	PoolAsset.Token.Amount = amount

	return pa.SetPoolAsset(denom, PoolAsset)
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
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	Id            uint64         `json:"id" yaml:"id"`
	PoolParams    PoolParams     `json:"pool_params" yaml:"pool_params"`
	TotalWeight   sdk.Int        `json:"total_weight" yaml:"total_weight"`
	TotalShare    sdk.Coin       `json:"total_share" yaml:"total_share"`
	PoolAssets    []PoolAsset    `json:"pool_assets" yaml:"pool_assets"`
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
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: pa.AccountNumber,
		Id:            pa.Id,
		PoolParams:    pa.PoolParams,
		TotalWeight:   pa.TotalWeight,
		TotalShare:    pa.TotalShare,
		PoolAssets:    pa.PoolAssets,
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
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: pa.AccountNumber,
		Id:            pa.Id,
		PoolParams:    pa.PoolParams,
		TotalWeight:   pa.TotalWeight,
		TotalShare:    pa.TotalShare,
		PoolAssets:    pa.PoolAssets,
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
	pa.TotalWeight = alias.TotalWeight
	pa.TotalShare = alias.TotalShare
	pa.PoolAssets = alias.PoolAssets

	return nil
}
