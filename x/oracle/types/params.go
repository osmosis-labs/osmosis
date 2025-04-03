package types

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"

	"gopkg.in/yaml.v2"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyVotePeriodEpochIdentifier  = []byte("VotePeriodEpochIdentifier")
	KeyVoteThreshold              = []byte("VoteThreshold")
	KeyRewardBand                 = []byte("RewardBand")
	KeyRewardDistributionWindow   = []byte("RewardDistributionWindow")
	KeyWhitelist                  = []byte("Whitelist")
	KeySlashFraction              = []byte("SlashFraction")
	KeySlashWindowEpochIdentifier = []byte("SlashWindowEpochIdentifier")
	KeyMinValidPerWindow          = []byte("MinValidPerWindow")
)

// Default parameter values
var (
	DefaultVoteThreshold              = osmomath.NewDecWithPrec(50, 2) // 50%
	DefaultRewardBand                 = osmomath.NewDecWithPrec(2, 2)  // 2% (-1, 1)
	DefaultTobinTax                   = osmomath.NewDecWithPrec(25, 4) // 0.25%
	DefaultWhitelist                  = DenomList{}
	DefaultSlashFraction              = osmomath.NewDecWithPrec(1, 4) // 0.01%
	DefaultMinValidPerWindow          = osmomath.NewDecWithPrec(5, 2) // 5%
	DefaultVotePeriodEpochIdentifier  = "minute"
	DefaultSlashWindowEpochIdentifier = "week"
)

var _ paramstypes.ParamSet = &Params{}

// DefaultParams creates default oracle module parameters
func DefaultParams() Params {
	return Params{
		VotePeriodEpochIdentifier:  DefaultVotePeriodEpochIdentifier,
		VoteThreshold:              DefaultVoteThreshold,
		RewardBand:                 DefaultRewardBand,
		RewardDistributionWindow:   1, // TODO: yurii: this is not used
		Whitelist:                  DefaultWhitelist,
		SlashFraction:              DefaultSlashFraction,
		SlashWindowEpochIdentifier: DefaultSlashWindowEpochIdentifier,
		MinValidPerWindow:          DefaultMinValidPerWindow,
	}
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of oracle module's parameters.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyVotePeriodEpochIdentifier, &p.VotePeriodEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramstypes.NewParamSetPair(KeyVoteThreshold, &p.VoteThreshold, validateVoteThreshold),
		paramstypes.NewParamSetPair(KeyRewardBand, &p.RewardBand, validateRewardBand),
		paramstypes.NewParamSetPair(KeyRewardDistributionWindow, &p.RewardDistributionWindow, validateRewardDistributionWindow),
		paramstypes.NewParamSetPair(KeyWhitelist, &p.Whitelist, validateWhitelist),
		paramstypes.NewParamSetPair(KeySlashFraction, &p.SlashFraction, validateSlashFraction),
		paramstypes.NewParamSetPair(KeySlashWindowEpochIdentifier, &p.SlashWindowEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramstypes.NewParamSetPair(KeyMinValidPerWindow, &p.MinValidPerWindow, validateMinValidPerWindow),
	}
}

// String implements fmt.Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs basic validation on oracle parameters.
func (p Params) Validate() error {
	if epochtypes.ValidateEpochIdentifierString(p.VotePeriodEpochIdentifier) != nil {
		return fmt.Errorf("oracle parameter VotePeriodEpochIdentifier must be valid, is %s", p.VotePeriodEpochIdentifier)
	}
	if epochtypes.ValidateEpochIdentifierString(p.SlashWindowEpochIdentifier) != nil {
		return fmt.Errorf("oracle parameter SlashWindowEpochIdentifier must be valid, is %s", p.SlashWindowEpochIdentifier)
	}
	if p.VoteThreshold.LTE(osmomath.NewDecWithPrec(33, 2)) {
		return fmt.Errorf("oracle parameter VoteThreshold must be greater than 33 percent")
	}

	if p.RewardBand.GT(osmomath.OneDec()) || p.RewardBand.IsNegative() {
		return fmt.Errorf("oracle parameter RewardBand must be between [0, 1]")
	}

	if p.SlashFraction.GT(osmomath.OneDec()) || p.SlashFraction.IsNegative() {
		return fmt.Errorf("oracle parameter SlashFraction must be between [0, 1]")
	}

	if p.MinValidPerWindow.GT(osmomath.OneDec()) || p.MinValidPerWindow.IsNegative() {
		return fmt.Errorf("oracle parameter MinValidPerWindow must be between [0, 1]")
	}

	for _, denom := range p.Whitelist {
		if denom.TobinTax.GT(osmomath.OneDec()) || denom.TobinTax.IsNegative() {
			return fmt.Errorf("oracle parameter Whitelist Denom must have TobinTax between [0, 1]")
		}
		if len(denom.Name) == 0 {
			return fmt.Errorf("oracle parameter Whitelist Denom must have name")
		}
	}
	return nil
}

func validateVotePeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("vote period must be positive: %d", v)
	}

	return nil
}

func validateVoteThreshold(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(osmomath.NewDecWithPrec(33, 2)) {
		return fmt.Errorf("vote threshold must be bigger than 33%%: %s", v)
	}

	if v.GT(osmomath.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", v)
	}

	return nil
}

func validateRewardBand(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("reward band must be positive: %s", v)
	}

	if v.GT(osmomath.OneDec()) {
		return fmt.Errorf("reward band is too large: %s", v)
	}

	return nil
}

func validateRewardDistributionWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("reward distribution window must be positive: %d", v)
	}

	return nil
}

func validateWhitelist(i interface{}) error {
	v, ok := i.(DenomList)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, d := range v {
		if d.TobinTax.GT(osmomath.OneDec()) || d.TobinTax.IsNegative() {
			return fmt.Errorf("oracle parameter Whitelist Denom must have TobinTax between [0, 1]")
		}
		if len(d.Name) == 0 {
			return fmt.Errorf("oracle parameter Whitelist Denom must have name")
		}
	}

	return nil
}

func validateSlashFraction(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("slash fraction must be positive: %s", v)
	}

	if v.GT(osmomath.OneDec()) {
		return fmt.Errorf("slash fraction is too large: %s", v)
	}

	return nil
}

func validateSlashWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("slash window must be positive: %d", v)
	}

	return nil
}

func validateMinValidPerWindow(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min valid per window must be positive: %s", v)
	}

	if v.GT(osmomath.OneDec()) {
		return fmt.Errorf("min valid per window is too large: %s", v)
	}

	return nil
}
