package types

import (
	"fmt"

	appparams "github.com/osmosis-labs/osmosis/v17/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyPoolCreationFee                                = []byte("PoolCreationFee")
	KeyDefaultTakerFee                                = []byte("DefaultTakerFee")
	KeyStableswapTakerFee                             = []byte("StableswapTakerFee")
	KeyCustomPoolTakerFee                             = []byte("CustomPoolTakerFee")
	KeyOsmoTakerFeeDistribution                       = []byte("OsmoTakerFeeDistribution")
	KeyNonOsmoTakerFeeDistribution                    = []byte("NonOsmoTakerFeeDistribution")
	KeyAuthorizedQuoteDenoms                          = []byte("AuthorizedQuoteDenoms")
	KeyCommunityPoolDenomToSwapNonWhitelistedAssetsTo = []byte("CommunityPoolDenomToSwapNonWhitelistedAssetsTo")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins, defaultTakerFee, stableswapTakerFee sdk.Dec, customPoolTakerFee []CustomPoolTakerFee, osmoTakerFeeDistribution, nonOsmoTakerFeeDistribution TakerFeeDistributionPercentage, authorizedQuoteDenoms []string, communityPoolDenomToSwapNonWhitelistedAssetsTo string) Params {
	return Params{
		PoolCreationFee:                                poolCreationFee,
		DefaultTakerFee:                                defaultTakerFee,
		StableswapTakerFee:                             stableswapTakerFee,
		CustomPoolTakerFee:                             customPoolTakerFee,
		OsmoTakerFeeDistribution:                       osmoTakerFeeDistribution,
		NonOsmoTakerFeeDistribution:                    nonOsmoTakerFeeDistribution,
		AuthorizedQuoteDenoms:                          authorizedQuoteDenoms,
		CommunityPoolDenomToSwapNonWhitelistedAssetsTo: communityPoolDenomToSwapNonWhitelistedAssetsTo,
	}
}

// DefaultParams are the default poolmanager module parameters.
func DefaultParams() Params {
	return Params{
		PoolCreationFee:    sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
		DefaultTakerFee:    sdk.MustNewDecFromStr("0.0015"),                                   // 0.15%
		StableswapTakerFee: sdk.MustNewDecFromStr("0.0002"),                                   // 0.02%
		CustomPoolTakerFee: []CustomPoolTakerFee{},
		OsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
			StakingRewards: sdk.MustNewDecFromStr("1"), // 100%
			CommunityPool:  sdk.MustNewDecFromStr("0"), // 0%
		},
		NonOsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
			StakingRewards: sdk.MustNewDecFromStr("0.67"), // 67%
			CommunityPool:  sdk.MustNewDecFromStr("0.33"), // 33%
		},
		AuthorizedQuoteDenoms: []string{
			"uosmo",
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // ATOM
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", // DAI
			"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
		CommunityPoolDenomToSwapNonWhitelistedAssetsTo: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}
	if err := validateDefaultTakerFee(p.DefaultTakerFee); err != nil {
		return err
	}
	if err := validateStableswapTakerFee(p.StableswapTakerFee); err != nil {
		return err
	}
	if err := validateCustomPoolTakerFee(p.CustomPoolTakerFee); err != nil {
		return err
	}
	if err := validateTakerFeeDistribution(p.OsmoTakerFeeDistribution); err != nil {
		return err
	}
	if err := validateTakerFeeDistribution(p.NonOsmoTakerFeeDistribution); err != nil {
		return err
	}
	if err := validateAuthorizedQuoteDenoms(p.AuthorizedQuoteDenoms); err != nil {
		return err
	}
	if err := validateCommunityPoolDenomToSwapNonWhitelistedAssetsTo(p.CommunityPoolDenomToSwapNonWhitelistedAssetsTo); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyDefaultTakerFee, &p.DefaultTakerFee, validateDefaultTakerFee),
		paramtypes.NewParamSetPair(KeyStableswapTakerFee, &p.StableswapTakerFee, validateStableswapTakerFee),
		paramtypes.NewParamSetPair(KeyCustomPoolTakerFee, &p.CustomPoolTakerFee, validateCustomPoolTakerFee),
		paramtypes.NewParamSetPair(KeyOsmoTakerFeeDistribution, &p.OsmoTakerFeeDistribution, validateTakerFeeDistribution),
		paramtypes.NewParamSetPair(KeyNonOsmoTakerFeeDistribution, &p.NonOsmoTakerFeeDistribution, validateTakerFeeDistribution),
		paramtypes.NewParamSetPair(KeyAuthorizedQuoteDenoms, &p.AuthorizedQuoteDenoms, validateAuthorizedQuoteDenoms),
		paramtypes.NewParamSetPair(KeyCommunityPoolDenomToSwapNonWhitelistedAssetsTo, &p.CommunityPoolDenomToSwapNonWhitelistedAssetsTo, validateCommunityPoolDenomToSwapNonWhitelistedAssetsTo),
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

func validateDefaultTakerFee(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	defaultTakerFee, ok := i.(sdk.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Ensure that the passed in discount rate is between 0 and 1.
	if defaultTakerFee.IsNegative() || defaultTakerFee.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid default taker fee: %s", defaultTakerFee)
	}

	return nil
}

func validateStableswapTakerFee(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	stableswapTakerFee, ok := i.(sdk.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Ensure that the passed in discount rate is between 0 and 1.
	if stableswapTakerFee.IsNegative() || stableswapTakerFee.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid stableswap taker fee: %s", stableswapTakerFee)
	}

	return nil
}

func validateCustomPoolTakerFee(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	customPoolTakerFee, ok := i.([]CustomPoolTakerFee)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, customPoolTakerFee := range customPoolTakerFee {
		if customPoolTakerFee.PoolId <= 0 {
			return fmt.Errorf("invalid pool ID: %d", customPoolTakerFee.PoolId)
		}
		if customPoolTakerFee.TakerFee.IsNegative() || customPoolTakerFee.TakerFee.GT(sdk.OneDec()) {
			return fmt.Errorf("invalid taker fee: %s", customPoolTakerFee.TakerFee)
		}
	}

	return nil
}

func validateTakerFeeDistribution(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	takerFeeDistribution, ok := i.(TakerFeeDistributionPercentage)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if takerFeeDistribution.StakingRewards.IsNegative() || takerFeeDistribution.StakingRewards.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid staking rewards distribution: %s", takerFeeDistribution.StakingRewards)
	}
	if takerFeeDistribution.CommunityPool.IsNegative() || takerFeeDistribution.CommunityPool.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid community pool distribution: %s", takerFeeDistribution.CommunityPool)
	}

	return nil
}

// validateAuthorizedQuoteDenoms validates a slice of authorized quote denoms.
//
// Parameters:
// - i: The parameter to validate.
//
// Returns:
// - An error if given type is not string slice.
// - An error if given slice is empty.
// - An error if any of the denoms are invalid.
func validateAuthorizedQuoteDenoms(i interface{}) error {
	authorizedQuoteDenoms, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(authorizedQuoteDenoms) == 0 {
		return fmt.Errorf("authorized quote denoms cannot be empty")
	}

	for _, denom := range authorizedQuoteDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return err
		}
	}

	return nil
}

func validateCommunityPoolDenomToSwapNonWhitelistedAssetsTo(i interface{}) error {
	communityPoolDenomToSwapNonWhitelistedAssetsTo, ok := i.(string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := sdk.ValidateDenom(communityPoolDenomToSwapNonWhitelistedAssetsTo); err != nil {
		return err
	}

	return nil
}
