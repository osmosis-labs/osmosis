package api

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
)

// Parameter store keys
var (
	KeySaleCreationFee               = []byte("SaleCreationFee")
	KeyMinimumDurationUntilStartTime = []byte("MinimumDurationUntilStartTime")
	KeyMinimumSaleDuration           = []byte("MinimumSaleDuration")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamTable for launchpad module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(saleCreationFee sdk.Coins, minimumDurationUntilStartTime, minimumSaleDuration time.Duration) Params {
	return Params{
		SaleCreationFee:               saleCreationFee,
		MinimumDurationUntilStartTime: minimumDurationUntilStartTime,
		MinimumSaleDuration:           minimumSaleDuration,
	}
}

// default launchpad module parameters
func DefaultParams() Params {
	return Params{
		SaleCreationFee:               sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
		MinimumDurationUntilStartTime: time.Hour * 24,                                                    // 1 Day
		MinimumSaleDuration:           time.Hour * 72,                                                    // 3 Days
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateSaleCreationFee(p.SaleCreationFee); err != nil {
		return err
	}
	if err := validateDuration(p.MinimumDurationUntilStartTime); err != nil {
		return err
	}
	if err := validateDuration(p.MinimumSaleDuration); err != nil {
		return err
	}
	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySaleCreationFee, &p.SaleCreationFee, validateSaleCreationFee),
		paramtypes.NewParamSetPair(KeyMinimumDurationUntilStartTime, &p.MinimumDurationUntilStartTime, validateDuration),
		paramtypes.NewParamSetPair(KeyMinimumSaleDuration, &p.MinimumSaleDuration, validateDuration),
	}
}

func validateSaleCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid sale creation fee: %+v", i)
	}

	return nil
}

func validateDuration(i interface{}) error {
	_, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
