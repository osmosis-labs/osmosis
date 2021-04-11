package types

import (
	"errors"
	"fmt"
	"strings"
	time "time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom         = []byte("MintDenom")
	KeyMaxRewardPerEpoch = []byte("MaxRewardPerEpoch")
	KeyMinRewardPerEpoch = []byte("MinRewardPerEpoch")
	KeyEpochsPerYear     = []byte("EpochsPerYear")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, annualProvisions, maxRewardPerEpoch, minRewardPerEpoch sdk.Dec, epochDuration time.Duration,
	reductionFactorForEvent sdk.Dec, reductionPeriodInEpochs, epochsPerYear int64,
) Params {

	return Params{
		MintDenom:               mintDenom,
		AnnualProvisions:        annualProvisions,
		MaxRewardPerEpoch:       maxRewardPerEpoch,
		MinRewardPerEpoch:       minRewardPerEpoch,
		EpochDuration:           epochDuration,
		ReductionPeriodInEpochs: reductionPeriodInEpochs,
		ReductionFactorForEvent: reductionFactorForEvent,
		EpochsPerYear:           epochsPerYear,
	}
}

// default minting module parameters
func DefaultParams() Params {
	epochDuration, _ := time.ParseDuration("168h") // 1 week
	return Params{
		MintDenom:               sdk.DefaultBondDenom,
		AnnualProvisions:        sdk.NewDec(5000000).Mul(sdk.NewDec(52)), // yearly rewards
		MaxRewardPerEpoch:       sdk.NewDec(6000000),                     // per epoch max
		MinRewardPerEpoch:       sdk.NewDec(4000000),                     // per epoch min
		EpochDuration:           epochDuration,                           // 1 week
		ReductionPeriodInEpochs: 156,                                     // 3 years
		ReductionFactorForEvent: sdk.NewDecWithPrec(5, 1),                // 0.5
		EpochsPerYear:           52,                                      // assuming 5 second block times
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateMaxRewardPerEpoch(p.MaxRewardPerEpoch); err != nil {
		return err
	}
	if err := validateMinRewardPerEpoch(p.MinRewardPerEpoch); err != nil {
		return err
	}
	if err := validateEpochsPerYear(p.EpochsPerYear); err != nil {
		return err
	}
	if p.MaxRewardPerEpoch.LT(p.MinRewardPerEpoch) {
		return fmt.Errorf(
			"max rewards (%s) must be greater than or equal to min rewards (%s)",
			p.MaxRewardPerEpoch, p.MinRewardPerEpoch,
		)
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
		paramtypes.NewParamSetPair(KeyMaxRewardPerEpoch, &p.MaxRewardPerEpoch, validateMaxRewardPerEpoch),
		paramtypes.NewParamSetPair(KeyMinRewardPerEpoch, &p.MinRewardPerEpoch, validateMinRewardPerEpoch),
		paramtypes.NewParamSetPair(KeyEpochsPerYear, &p.EpochsPerYear, validateEpochsPerYear),
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

func validateMaxRewardPerEpoch(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max rewards cannot be negative: %s", v)
	}

	return nil
}

func validateMinRewardPerEpoch(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min rewards cannot be negative: %s", v)
	}

	return nil
}

func validateEpochsPerYear(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}
