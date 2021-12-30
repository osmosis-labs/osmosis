package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
)

// Parameter store keys
var (
	KeyPoolCreationFee    = []byte("PoolCreationFee")
	KeyDefaultPoolSwapFee = []byte("DefaultPoolSwapFee")
	KeyDefaultPoolExitFee = []byte("DefaultPoolExitFee")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins, defaultPoolSwapFee, defaultPoolExitFee sdk.Dec) Params {
	return Params{
		PoolCreationFee:    poolCreationFee,
		DefaultPoolSwapFee: defaultPoolSwapFee,
		DefaultPoolExitFee: defaultPoolExitFee,
	}
}

// default gamm module parameters
func DefaultParams() Params {
	return Params{
		PoolCreationFee:    sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
		DefaultPoolSwapFee: sdk.MustNewDecFromStr("0.02"),                                     // 1000 OSMO
		DefaultPoolExitFee: sdk.MustNewDecFromStr("0"),                                        // 1000 OSMO
	}
}

// validate params
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}

	if err := validateDefaultPoolSwapFee(p.DefaultPoolSwapFee); err != nil {
		return err
	}

	if err := validateDefaultPoolExitFee(p.DefaultPoolExitFee); err != nil {
		return err
	}

	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyDefaultPoolSwapFee, &p.DefaultPoolSwapFee, validateDefaultPoolSwapFee),
		paramtypes.NewParamSetPair(KeyDefaultPoolExitFee, &p.DefaultPoolExitFee, validateDefaultPoolExitFee),
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

func validateDefaultPoolSwapFee(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) || v.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid default pool swap fee: %+v", i)
	}

	return nil
}

func validateDefaultPoolExitFee(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) || v.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid default pool exit fee: %+v", i)
	}

	return nil
}
