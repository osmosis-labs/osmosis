package types

import (
	fmt "fmt"

	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyCreateGaugeFee       = []byte("CreateGaugeFee")
	KeyAddToGaugeFee        = []byte("AddToGaugeFee")
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams takes an epoch distribution identifier, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string, createGaugeFee, addToGaugeFee sdk.Coins) Params {
	return Params{
		DistrEpochIdentifier: distrEpochIdentifier,
		CreateGaugeFee:       createGaugeFee,
		AddToGaugeFee:        addToGaugeFee,
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier: "week",
		CreateGaugeFee:       sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(50000000))),
		AddToGaugeFee:        sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(25000000))),
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	if err := validateFeeCoin(p.CreateGaugeFee); err != nil {
		return err
	}
	if err := validateFeeCoin(p.AddToGaugeFee); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyCreateGaugeFee, &p.CreateGaugeFee, validateFeeCoin),
		paramtypes.NewParamSetPair(KeyAddToGaugeFee, &p.AddToGaugeFee, validateFeeCoin),
	}
}

func validateFeeCoin(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid fee parameter: %+v", i)
	}

	return nil
}
