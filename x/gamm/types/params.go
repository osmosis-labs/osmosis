package types

import (
	"fmt"

	appparams "github.com/dymensionxyz/dymension/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyPoolCreationFee   = []byte("PoolCreationFee")
	KeyEnabledGlobalFees = []byte("EnabledGlobalFees")
	KeyGlobalFees        = []byte("GlobalPoolFees")
	KeyTakerFees         = []byte("TakerFees")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins) Params {
	return Params{
		PoolCreationFee:      poolCreationFee,
		EnableGlobalPoolFees: false,
		GlobalFees:           GlobalFees{sdk.ZeroDec(), sdk.ZeroDec()},
		TakerFee:             sdk.ZeroDec(),
	}
}

// default gamm module parameters.
func DefaultParams() Params {
	return Params{
		// set correct defaults
		PoolCreationFee:      sdk.Coins{sdk.NewInt64Coin(appparams.BaseDenom, 1000_000_000)},
		EnableGlobalPoolFees: false,
		GlobalFees:           GlobalFees{sdk.ZeroDec(), sdk.ZeroDec()},
		TakerFee:             sdk.ZeroDec(),
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}
	if err := validateGlobalFees(p.GlobalFees); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyEnabledGlobalFees, &p.EnableGlobalPoolFees, func(value interface{}) error { return nil }),
		paramtypes.NewParamSetPair(KeyGlobalFees, &p.GlobalFees, validateGlobalFees),
		paramtypes.NewParamSetPair(KeyTakerFees, &p.TakerFee, validateTakerFees),
	}
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid pool creation fee: %+v", i)
	}

	return nil
}

func validateGlobalFees(i interface{}) error {
	v, ok := i.(GlobalFees)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.ExitFee.IsNil() || v.SwapFee.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", i)
	}
	if v.ExitFee.IsNegative() {
		return ErrNegativeExitFee
	}

	if v.ExitFee.GTE(sdk.OneDec()) {
		return ErrTooMuchExitFee
	}

	if v.SwapFee.IsNegative() {
		return ErrNegativeSwapFee
	}

	if v.SwapFee.GTE(sdk.OneDec()) {
		return ErrTooMuchSwapFee
	}

	return nil
}

func validateTakerFees(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", i)
	}
	if v.IsNegative() {
		return ErrNegativeExitFee
	}

	if v.GTE(sdk.OneDec()) {
		return ErrTooMuchExitFee
	}

	return nil
}
