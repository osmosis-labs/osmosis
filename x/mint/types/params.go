package types

import (
	"errors"
	"fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	yaml "gopkg.in/yaml.v2"
)

// Parameter store keys
var (
	KeyMintDenom                           = []byte("MintDenom")
	KeyGenesisEpochProvisions              = []byte("GenesisEpochProvisions")
	KeyEpochIdentifier                     = []byte("EpochIdentifier")
	KeyReductionPeriodInEpochs             = []byte("ReductionPeriodInEpochs")
	KeyReductionFactor                     = []byte("ReductionFactor")
	KeyPoolAllocationRatio                 = []byte("PoolAllocationRatio")
	KeyDeveloperRewardsReceiver            = []byte("DeveloperRewardsReceiver")
	KeyMintingRewardsDistributionStartTime = []byte("MintingRewardsDistributionStartTime")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, genesisEpochProvisions sdk.Dec, epochIdentifier string,
	ReductionFactor sdk.Dec, reductionPeriodInEpochs int64, distrProportions DistributionProportions,
	devRewardsReceiver string, mintingRewardsDistributionStartTime time.Time,
) Params {

	return Params{
		MintDenom:                           mintDenom,
		GenesisEpochProvisions:              genesisEpochProvisions,
		EpochIdentifier:                     epochIdentifier,
		ReductionPeriodInEpochs:             reductionPeriodInEpochs,
		ReductionFactor:                     ReductionFactor,
		DistributionProportions:             distrProportions,
		DeveloperRewardsReceiver:            devRewardsReceiver,
		MintingRewardsDistributionStartTime: mintingRewardsDistributionStartTime,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:               sdk.DefaultBondDenom,
		GenesisEpochProvisions:  sdk.NewDec(5000000),
		EpochIdentifier:         "weekly",                 // 1 week
		ReductionPeriodInEpochs: 156,                      // 3 years
		ReductionFactor:         sdk.NewDecWithPrec(5, 1), // 0.5
		DistributionProportions: DistributionProportions{
			Staking:          sdk.NewDecWithPrec(5, 1), // 0.5
			PoolIncentives:   sdk.NewDecWithPrec(3, 1), // 0.3
			DeveloperRewards: sdk.NewDecWithPrec(2, 1), // 0.2
		},
		DeveloperRewardsReceiver:            "",
		MintingRewardsDistributionStartTime: time.Time{},
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateGenesisEpochProvisions(p.GenesisEpochProvisions); err != nil {
		return err
	}
	if err := validateEpochIdentifier(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateReductionPeriodInEpochs(p.ReductionPeriodInEpochs); err != nil {
		return err
	}
	if err := validateReductionFactor(p.ReductionFactor); err != nil {
		return err
	}
	if err := validateDistributionProportions(p.DistributionProportions); err != nil {
		return err
	}
	if err := validateDeveloperRewardsReceiver(p.DeveloperRewardsReceiver); err != nil {
		return err
	}
	if err := validateMintingRewardsDistributionStartTime(p.MintingRewardsDistributionStartTime); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyGenesisEpochProvisions, &p.GenesisEpochProvisions, validateGenesisEpochProvisions),
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, validateEpochIdentifier),
		paramtypes.NewParamSetPair(KeyReductionPeriodInEpochs, &p.ReductionPeriodInEpochs, validateReductionPeriodInEpochs),
		paramtypes.NewParamSetPair(KeyReductionFactor, &p.ReductionFactor, validateReductionFactor),
		paramtypes.NewParamSetPair(KeyPoolAllocationRatio, &p.DistributionProportions, validateDistributionProportions),
		paramtypes.NewParamSetPair(KeyDeveloperRewardsReceiver, &p.DeveloperRewardsReceiver, validateDeveloperRewardsReceiver),
		paramtypes.NewParamSetPair(KeyMintingRewardsDistributionStartTime, &p.MintingRewardsDistributionStartTime, validateMintingRewardsDistributionStartTime),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateGenesisEpochProvisions(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) {
		return fmt.Errorf("genesis epoch provision must be non-negative")
	}

	return nil
}

func validateEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("empty distribution epoch identifier: %+v", i)
	}

	return nil
}

func validateReductionPeriodInEpochs(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("max validators must be positive: %d", v)
	}

	return nil
}

func validateReductionFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	return nil
}

func validateDistributionProportions(i interface{}) error {
	v, ok := i.(DistributionProportions)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Staking.IsNegative() {
		return errors.New("staking distribution ratio should not be negative")
	}

	if v.PoolIncentives.IsNegative() {
		return errors.New("pool incentives distribution ratio should not be negative")
	}

	if v.DeveloperRewards.IsNegative() {
		return errors.New("developer rewards distribution ratio should not be negative")
	}

	totalProportions := v.Staking.Add(v.PoolIncentives).Add(v.DeveloperRewards)

	if !totalProportions.Equal(sdk.NewDec(1)) {
		return errors.New("total distributions ratio should be 1")
	}

	return nil
}

func validateDeveloperRewardsReceiver(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// fund community pool when rewards address is empty
	if v == "" {
		return nil
	}

	_, err := sdk.AccAddressFromBech32(v)
	return err
}

func validateMintingRewardsDistributionStartTime(i interface{}) error {
	_, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
