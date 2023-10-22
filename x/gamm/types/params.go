package types

import (
	"fmt"

	appparams "github.com/dymensionxyz/dymension/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyPoolCreationFee  = []byte("PoolCreationFee")
	KeyGlobalFees       = []byte("GlobalFees")
	KeyGlobalPoolParams = []byte("GlobalPoolParams")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins) Params {
	return Params{
		PoolCreationFee: poolCreationFee,
		GlobalFees:      false,
		PoolParams:      GlobalPoolParams{sdk.ZeroDec(), sdk.ZeroDec()},
	}
}

// default gamm module parameters.
func DefaultParams() Params {
	return Params{
		PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseDenom, 1000_000_000)},
		PoolParams:      GlobalPoolParams{sdk.ZeroDec(), sdk.ZeroDec()},
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}
	if err := validateGlobalFees(p.PoolParams); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyGlobalFees, &p.GlobalFees, func(value interface{}) error { return nil }),
		paramtypes.NewParamSetPair(KeyGlobalPoolParams, &p.PoolParams, validateGlobalFees),
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
	v, ok := i.(GlobalPoolParams)
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
