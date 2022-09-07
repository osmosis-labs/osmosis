package types

import (
	"fmt"

	appparams "github.com/osmosis-labs/osmosis/v12/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyValSetCreationFee = []byte("ValSetCreationFee")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(valSetCreationFee sdk.Coins) Params {
	return Params{
		ValsetCreationFee: valSetCreationFee,
	}
}

// default gamm module parameters.
func DefaultParams() Params {
	return Params{
		ValsetCreationFee: sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 10_000_000)), // 10 OSMO
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validateValSetCreationFee(p.ValsetCreationFee); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyValSetCreationFee, &p.ValsetCreationFee, validateValSetCreationFee),
	}
}

func validateValSetCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid val-set creation fee: %+v", i)
	}

	return nil
}
