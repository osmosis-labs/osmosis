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
	KeyMintDenom               = []byte("MintDenom")
	KeyAnnualProvisions        = []byte("AnnualProvisions")
	KeyEpochDuration           = []byte("EpochDuration")
	KeyReductionPeriodInEpochs = []byte("ReductionPeriodInEpochs")
	KeyReductionFactorForEvent = []byte("ReductionFactorForEvent")
	KeyEpochsPerYear           = []byte("EpochsPerYear")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, annualProvisions sdk.Dec, epochDuration time.Duration,
	reductionFactorForEvent sdk.Dec, reductionPeriodInEpochs, epochsPerYear int64,
) Params {

	return Params{
		MintDenom:               mintDenom,
		AnnualProvisions:        annualProvisions,
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
	if err := validateAnnualProvisions(p.AnnualProvisions); err != nil {
		return err
	}
	if err := validateEpochDuration(p.EpochDuration); err != nil {
		return err
	}
	if err := validateReductionPeriodInEpochs(p.ReductionPeriodInEpochs); err != nil {
		return err
	}
	if err := validateReductionFactorForEvent(p.ReductionFactorForEvent); err != nil {
		return err
	}
	if err := validateEpochsPerYear(p.EpochsPerYear); err != nil {
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
		paramtypes.NewParamSetPair(KeyAnnualProvisions, &p.AnnualProvisions, validateAnnualProvisions),
		paramtypes.NewParamSetPair(KeyEpochDuration, &p.EpochDuration, validateEpochDuration),
		paramtypes.NewParamSetPair(KeyReductionPeriodInEpochs, &p.ReductionPeriodInEpochs, validateReductionPeriodInEpochs),
		paramtypes.NewParamSetPair(KeyReductionFactorForEvent, &p.ReductionFactorForEvent, validateReductionFactorForEvent),
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

func validateAnnualProvisions(i interface{}) error {

	// TODO

	return nil
}

func validateEpochDuration(i interface{}) error {

	// TODO

	return nil
}

func validateReductionPeriodInEpochs(i interface{}) error {

	// TODO

	return nil
}

func validateReductionFactorForEvent(i interface{}) error {

	// TODO

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
