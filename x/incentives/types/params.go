package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyGroupCreationFee     = []byte("GroupCreationFee")
	KeyCreatorWhitelist     = []byte("CreatorWhitelist")

	// 100 OSMO
	DefaultGroupCreationFee = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100_000_000)))
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams takes an epoch distribution identifier and group creation fee, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string, groupCreationFee sdk.Coins) Params {
	return Params{
		DistrEpochIdentifier:         distrEpochIdentifier,
		GroupCreationFee:             groupCreationFee,
		UnrestrictedCreatorWhitelist: []string{},
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier:         "week",
		GroupCreationFee:             DefaultGroupCreationFee,
		UnrestrictedCreatorWhitelist: []string{},
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}

	if err := ValidateGroupCreaionFee(p.GroupCreationFee); err != nil {
		return err
	}

	if err := osmoutils.ValidateAddressList(p.UnrestrictedCreatorWhitelist); err != nil {
		return err
	}

	return nil
}

func ValidateGroupCreaionFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return v.Validate()
}

func ValidateGroupCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return v.Validate()
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyGroupCreationFee, &p.GroupCreationFee, ValidateGroupCreaionFee),
		paramtypes.NewParamSetPair(KeyCreatorWhitelist, &p.UnrestrictedCreatorWhitelist, osmoutils.ValidateAddressList),
	}
}
