package types

import (
	"fmt"

	appparams "github.com/osmosis-labs/osmosis/v19/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyPoolCreationFee = []byte("PoolCreationFee")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(poolCreationFee sdk.Coins) Params {
	return Params{
		PoolCreationFee: poolCreationFee,
	}
}

// DefaultParams are the default poolmanager module parameters.
func DefaultParams() Params {
	return Params{
		PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
<<<<<<< HEAD
=======
		TakerFeeParams: TakerFeeParams{
			DefaultTakerFee: sdk.ZeroDec(), // 0%
			OsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
				StakingRewards: sdk.MustNewDecFromStr("1"), // 100%
				CommunityPool:  sdk.MustNewDecFromStr("0"), // 0%
			},
			NonOsmoTakerFeeDistribution: TakerFeeDistributionPercentage{
				StakingRewards: sdk.MustNewDecFromStr("0.67"), // 67%
				CommunityPool:  sdk.MustNewDecFromStr("0.33"), // 33%
			},
			AdminAddresses: []string{},
			CommunityPoolDenomToSwapNonWhitelistedAssetsTo: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
		AuthorizedQuoteDenoms: []string{
			"uosmo",
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // ATOM
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", // DAI
			"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
>>>>>>> afcdf429 (feat(taker fees): Implement taker fee collection and tracking tests (#6183))
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validatePoolCreationFee(p.PoolCreationFee); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
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
