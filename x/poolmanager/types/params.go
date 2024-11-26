package types

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyPoolCreationFee                                = []byte("PoolCreationFee")
	KeyDefaultTakerFee                                = []byte("DefaultTakerFee")
	KeyOsmoTakerFeeDistribution                       = []byte("OsmoTakerFeeDistribution")
	KeyNonOsmoTakerFeeDistribution                    = []byte("NonOsmoTakerFeeDistribution")
	KeyAdminAddresses                                 = []byte("AdminAddresses")
	KeyCommunityPoolDenomToSwapNonWhitelistedAssetsTo = []byte("CommunityPoolDenomToSwapNonWhitelistedAssetsTo")
	KeyAuthorizedQuoteDenoms                          = []byte("AuthorizedQuoteDenoms")
	KeyReducedTakerFeeByWhitelist                     = []byte("ReducedTakerFeeByWhitelist")

	ZeroDec = osmomath.ZeroDec()
	OneDec  = osmomath.OneDec()
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins,
	defaultTakerFee osmomath.Dec,
	osmoTakerFeeDistribution, nonOsmoTakerFeeDistribution TakerFeeDistributionPercentage,
	adminAddresses, authorizedQuoteDenoms []string,
	communityPoolDenomToSwapNonWhitelistedAssetsTo string) Params {
	return Params{
		PoolCreationFee: poolCreationFee,
		TakerFeeParams: TakerFeeParams{
			DefaultTakerFee:                                defaultTakerFee,
			OsmoTakerFeeDistribution:                       osmoTakerFeeDistribution,
			NonOsmoTakerFeeDistribution:                    nonOsmoTakerFeeDistribution,
			AdminAddresses:                                 adminAddresses,
			CommunityPoolDenomToSwapNonWhitelistedAssetsTo: communityPoolDenomToSwapNonWhitelistedAssetsTo,
		},
		AuthorizedQuoteDenoms: authorizedQuoteDenoms,
	}
}

// DefaultParams are the default poolmanager module parameters.
func DefaultParams() Params {
	return Params{
		PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
		TakerFeeParams: TakerFeeParams{
			DefaultTakerFee: ZeroDec, // 0%
			OsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
				StakingRewards: osmomath.MustNewDecFromStr("1"), // 100%
				CommunityPool:  osmomath.MustNewDecFromStr("0"), // 0%
			},
			NonOsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
				StakingRewards: osmomath.MustNewDecFromStr("0.67"), // 67%
				CommunityPool:  osmomath.MustNewDecFromStr("0.33"), // 33%
			},
			AdminAddresses: []string{},
			CommunityPoolDenomToSwapNonWhitelistedAssetsTo: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
			ReducedFeeWhitelist:                            []string{},
		},
		AuthorizedQuoteDenoms: []string{
			appparams.BaseCoinUnit,
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // ATOM
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", // DAI
			"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}
	if err := validateDefaultTakerFee(p.TakerFeeParams.DefaultTakerFee); err != nil {
		return err
	}
	if err := validateTakerFeeDistribution(p.TakerFeeParams.OsmoTakerFeeDistribution); err != nil {
		return err
	}
	if err := validateTakerFeeDistribution(p.TakerFeeParams.NonOsmoTakerFeeDistribution); err != nil {
		return err
	}
	if err := validateAdminAddresses(p.TakerFeeParams.AdminAddresses); err != nil {
		return err
	}
	if err := validateCommunityPoolDenomToSwapNonWhitelistedAssetsTo(p.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo); err != nil {
		return err
	}
	if err := osmoutils.ValidateAddressList(p.TakerFeeParams.ReducedFeeWhitelist); err != nil {
		return err
	}
	if err := validateAuthorizedQuoteDenoms(p.AuthorizedQuoteDenoms); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramtypes.NewParamSetPair(KeyDefaultTakerFee, &p.TakerFeeParams.DefaultTakerFee, validateDefaultTakerFee),
		paramtypes.NewParamSetPair(KeyOsmoTakerFeeDistribution, &p.TakerFeeParams.OsmoTakerFeeDistribution, validateTakerFeeDistribution),
		paramtypes.NewParamSetPair(KeyNonOsmoTakerFeeDistribution, &p.TakerFeeParams.NonOsmoTakerFeeDistribution, validateTakerFeeDistribution),
		paramtypes.NewParamSetPair(KeyAdminAddresses, &p.TakerFeeParams.AdminAddresses, validateAdminAddresses),
		paramtypes.NewParamSetPair(KeyCommunityPoolDenomToSwapNonWhitelistedAssetsTo, &p.TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo, validateCommunityPoolDenomToSwapNonWhitelistedAssetsTo),
		paramtypes.NewParamSetPair(KeyAuthorizedQuoteDenoms, &p.AuthorizedQuoteDenoms, validateAuthorizedQuoteDenoms),
		paramtypes.NewParamSetPair(KeyReducedTakerFeeByWhitelist, &p.TakerFeeParams.ReducedFeeWhitelist, osmoutils.ValidateAddressList),
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
	// Convert the given parameter to osmomath.Dec.
	defaultTakerFee, ok := i.(osmomath.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Ensure that the passed in discount rate is between 0 and 1.
	if defaultTakerFee.IsNegative() || defaultTakerFee.GT(OneDec) {
		return fmt.Errorf("invalid default taker fee: %s", defaultTakerFee)
	}

	return nil
}

func validateTakerFeeDistribution(i interface{}) error {
	// Convert the given parameter to osmomath.Dec.
	takerFeeDistribution, ok := i.(TakerFeeDistributionPercentage)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if takerFeeDistribution.StakingRewards.IsNegative() || takerFeeDistribution.StakingRewards.GT(OneDec) {
		return fmt.Errorf("invalid staking rewards distribution: %s", takerFeeDistribution.StakingRewards)
	}
	if takerFeeDistribution.CommunityPool.IsNegative() || takerFeeDistribution.CommunityPool.GT(OneDec) {
		return fmt.Errorf("invalid community pool distribution: %s", takerFeeDistribution.CommunityPool)
	}

	return nil
}

func validateAdminAddresses(i interface{}) error {
	adminAddresses, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(adminAddresses) > 0 {
		for _, adminAddress := range adminAddresses {
			if _, err := sdk.AccAddressFromBech32(adminAddress); err != nil {
				return fmt.Errorf("invalid account address: %s", adminAddress)
			}
		}
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

func validateDenomPairTakerFees(pairs []DenomPairTakerFee) error {
	if len(pairs) == 0 {
		return fmt.Errorf("Empty denom pair taker fee")
	}

	for _, record := range pairs {
		if record.TokenInDenom == record.TokenOutDenom {
			return fmt.Errorf("TokenInDenom and TokenOutDenom must be different")
		}

		if sdk.ValidateDenom(record.TokenInDenom) != nil {
			return fmt.Errorf("TokenInDenom is invalid: %s", sdk.ValidateDenom(record.TokenInDenom))
		}

		if sdk.ValidateDenom(record.TokenOutDenom) != nil {
			return fmt.Errorf("TokenOutDenom is invalid: %s", sdk.ValidateDenom(record.TokenOutDenom))
		}

		takerFee := record.TakerFee
		if takerFee.IsNegative() || takerFee.GTE(OneDec) {
			return fmt.Errorf("taker fee must be between 0 and 1: %s", takerFee.String())
		}
	}
	return nil
}
