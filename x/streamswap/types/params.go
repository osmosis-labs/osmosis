package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/v10/app/params"
)

// Parameter store keys
var (
	KeySaleCreationFee               = []byte("SaleCreationFee")
	KeySaleCreationFeeRecipient      = []byte("SaleCreationFeeRecipient")
	KeyMinimumDurationUntilStartTime = []byte("MinimumDurationUntilStartTime")
	KeyMinimumSaleDuration           = []byte("MinimumSaleDuration")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamTable for streamswap module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(saleCreationFee sdk.Coins, saleCreationFeeRecipient string, minimumDurationUntilStartTime, minimumSaleDuration time.Duration) Params {
	return Params{
		SaleCreationFee:               saleCreationFee,
		SaleCreationFeeRecipient:      saleCreationFeeRecipient,
		MinimumDurationUntilStartTime: minimumDurationUntilStartTime,
		MinimumSaleDuration:           minimumSaleDuration,
	}
}

// default streamswap module parameters
func DefaultParams() Params {
	return Params{
		SaleCreationFee:               sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 200_000_000)}, // 200 OSMO
		MinimumDurationUntilStartTime: time.Hour * 24,                                                   // 1 Day
		MinimumSaleDuration:           time.Hour * 24,                                                   // 1 Day
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
		paramtypes.NewParamSetPair(KeySaleCreationFeeRecipient, &p.SaleCreationFeeRecipient, validateSaleCreationFeeRecipient),
		paramtypes.NewParamSetPair(KeyMinimumDurationUntilStartTime, &p.MinimumDurationUntilStartTime, validateDuration),
		paramtypes.NewParamSetPair(KeyMinimumSaleDuration, &p.MinimumSaleDuration, validateDuration),
	}
}
func validateSaleCreationFeeRecipient(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T, expected string bech32", i)
	}
	if _, err := sdk.AccAddressFromBech32(v); err != nil {
		return fmt.Errorf("invalid parameter type: expected string bech32")
	}
	return nil
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
