package types

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v28/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyOptInFee                             = []byte("OptInFee")
	KeyStakeMinRequirement                  = []byte("StakeMinRequirement")
	KeyRollingWindow                        = []byte("RollingWindow")
	KeyMinTradeValueToInitializeDayTracking = []byte("MinTradeValueToInitializeDayTracking")
	KeyFeeTiers                             = []byte("FeeTiers")
	KeyOsmoUsdPoolId                        = []byte("OsmoUsdPoolId")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(optInFee sdk.Coin,
	stakeMinRequirement osmomath.Int,
	rollingWindow uint64,
	minTradeValueToInitializeDayTracking osmomath.Int,
	feeTiers FeeTiers, osmoUsdPoolId uint64) Params {
	return Params{
		OptInFee:                             optInFee,
		StakeMinRequirement:                  stakeMinRequirement,
		RollingWindow:                        rollingWindow,
		MinTradeValueToInitializeDayTracking: minTradeValueToInitializeDayTracking,
		FeeTiers:                             feeTiers,
		OsmoUsdPoolId:                        osmoUsdPoolId,
	}
}

// DefaultParams are the default poolmanager module parameters.
func DefaultParams() Params {
	return Params{
		OptInFee:                             sdk.NewInt64Coin(appparams.BaseCoinUnit, 5_000_000), // 5 OSMO
		StakeMinRequirement:                  osmomath.NewInt(500_000_000),                        // 500 OSMO
		RollingWindow:                        30,                                                  // 30 days
		MinTradeValueToInitializeDayTracking: osmomath.NewInt(50_000_000),                         // 50 OSMO,
		FeeTiers: FeeTiers{
			[]FeeTier{
				{
					TierId:                     1,
					TierMinStakeRequirement:    osmomath.NewInt(500_000_000), // 500 OSMO
					TierMinRollingWindowVolume: osmomath.NewInt(10_000),      // 10_000 USD
				},
				{
					TierId:                     2,
					TierMinStakeRequirement:    osmomath.NewInt(1_000_000_000), // 1_000 OSMO
					TierMinRollingWindowVolume: osmomath.NewInt(50_000),        // 50_000 USD
				},
				{
					TierId:                     3,
					TierMinStakeRequirement:    osmomath.NewInt(5_000_000_000), // 5_000 OSMO
					TierMinRollingWindowVolume: osmomath.NewInt(100_000),       // 100_000 USD
				},
				{
					TierId:                     4,
					TierMinStakeRequirement:    osmomath.NewInt(25_000_000_000), // 25_000 OSMO
					TierMinRollingWindowVolume: osmomath.NewInt(250_000),        // 250_000 USD
				},
				{
					TierId:                     5,
					TierMinStakeRequirement:    osmomath.NewInt(50_000_000_000), // 50_000 OSMO
					TierMinRollingWindowVolume: osmomath.NewInt(500_000),        // 500_000 USD
				},
			},
		},
		OsmoUsdPoolId: 1464, // https://app.osmosis.zone/pool/1464
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validateOptInFee(p.OptInFee); err != nil {
		return err
	}
	if err := validateStakeMinRequirement(p.StakeMinRequirement); err != nil {
		return err
	}
	if err := validateRollingWindow(p.RollingWindow); err != nil {
		return err
	}
	if err := validateMinTradeValueToInitializeDayTracking(p.MinTradeValueToInitializeDayTracking); err != nil {
		return err
	}
	if err := validateFeeTiers(p.FeeTiers); err != nil {
		return err
	}
	if err := validateOsmoUsdPoolId(p.OsmoUsdPoolId); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOptInFee, &p.OptInFee, validateOptInFee),
		paramtypes.NewParamSetPair(KeyStakeMinRequirement, &p.StakeMinRequirement, validateStakeMinRequirement),
		paramtypes.NewParamSetPair(KeyRollingWindow, &p.RollingWindow, validateRollingWindow),
		paramtypes.NewParamSetPair(KeyMinTradeValueToInitializeDayTracking, &p.MinTradeValueToInitializeDayTracking, validateMinTradeValueToInitializeDayTracking),
		paramtypes.NewParamSetPair(KeyFeeTiers, &p.FeeTiers, validateFeeTiers),
		paramtypes.NewParamSetPair(KeyOsmoUsdPoolId, &p.OsmoUsdPoolId, validateOsmoUsdPoolId),
	}
}

func validateOptInFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("opt-in fee cannot be negative: %s", v)
	}

	return nil
}

func validateStakeMinRequirement(i interface{}) error {
	v, ok := i.(osmomath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("stake minimum requirement cannot be negative: %s", v)
	}

	return nil
}

func validateRollingWindow(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("rolling window must be positive: %d", v)
	}

	return nil
}

func validateMinTradeValueToInitializeDayTracking(i interface{}) error {
	v, ok := i.(osmomath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum trade value to initialize day tracking cannot be negative: %s", v)
	}

	return nil
}

func validateFeeTiers(i interface{}) error {
	v, ok := i.(FeeTiers)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, tier := range v.FeeTiers {
		if tier.TierId <= 0 {
			return fmt.Errorf("fee tier id must be positive: %d", tier.TierId)
		}
		if tier.TierMinRollingWindowVolume.IsNegative() {
			return fmt.Errorf("fee tier min rolling window volume cannot be negative: %s", tier.TierMinRollingWindowVolume)
		}
		if tier.TierMinStakeRequirement.IsNegative() {
			return fmt.Errorf("fee tier min stake requirement cannot be negative: %s", tier.TierMinStakeRequirement)
		}
	}

	return nil
}

func validateOsmoUsdPoolId(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("osmo-usd pool id cannot be zero: %d", v)
	}

	return nil
}
