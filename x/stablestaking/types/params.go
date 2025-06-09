package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
)

const MinUnbondingTime = time.Hour * 24 * 6
const MaxUnbondingTime = time.Hour * 24 * 22

// Parameter keys
var (
	// KeyRewardRate the rate value to calculate user's rewards
	KeyRewardRate = []byte("RewardRate")
	// KeyUnbondingTime the period required to get tokens back
	KeyUnbondingTime = []byte("UnbondingTime")
	// KeySupportedTokens the list of tokens which can be staked
	KeySupportedTokens = []byte("SupportedTokens")
	// KeyMaxStakingAmount the value of maximum staking
	KeyMaxStakingAmount = []byte("MaxStakingAmount")
	// KeyRewardEpochIdentifier the period required to distribute rewards
	KeyRewardEpochIdentifier = []byte("RewardEpochIdentifier")
	// KeyUnbondingEpochIdentifier the period required to unbond stake
	KeyUnbondingEpochIdentifier = []byte("UnbondingEpochIdentifier")
)

// AllowedTokens the list of stable coins to be allowed to stake
var AllowedTokens = []string{appParams.MicroUSDDenom, appParams.MicroHKDDenom, appParams.MicroVNDDenom}
var AllowedEpochs = []string{"week", "day"}

var _ paramstypes.ParamSet = &Params{}

func DefaultParams() Params {
	return Params{
		RewardRate:               osmomath.NewDecWithPrec(5, 2).String(), // 0.05%
		UnbondingDuration:        time.Hour * 24 * 14,
		SupportedTokens:          AllowedTokens,
		UnbondingEpochIdentifier: AllowedEpochs[1],
		RewardEpochIdentifier:    AllowedEpochs[0],
	}
}

func (p Params) Validate() error {
	return nil
}

func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		types.NewParamSetPair(KeyRewardEpochIdentifier, &p.RewardEpochIdentifier, validateEpoch),
		types.NewParamSetPair(KeyUnbondingEpochIdentifier, &p.UnbondingEpochIdentifier, validateEpoch),
		types.NewParamSetPair(KeyRewardRate, &p.RewardRate, validateRate),
		types.NewParamSetPair(KeyUnbondingTime, &p.UnbondingDuration, validateUnbondingDuration),
		types.NewParamSetPair(KeySupportedTokens, &p.SupportedTokens, validateSupportedTokens),
	}
}

func validateEpoch(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, a := range AllowedEpochs {
		if v == a {
			return nil
		}
	}

	return fmt.Errorf("invalid epoch identifier: %s", v)
}

func validateRate(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	rewardRate, err := osmomath.NewDecFromStr(v)
	if err != nil {
		return fmt.Errorf("invalid reward rate: %v", err)
	}

	if !rewardRate.IsPositive() {
		return fmt.Errorf("reward rate must be positive")
	}

	return nil
}

func IsAllowedToken(token string) bool {
	for _, t := range AllowedTokens {
		if t == token {
			return true
		}
	}
	return false
}

func validateSupportedTokens(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("supported tokens cannot be empty")
	}

	for _, token := range v {
		if !IsAllowedToken(token) {
			return fmt.Errorf("unsupported token: %s", token)
		}
	}

	return nil
}

func validateUnbondingDuration(i interface{}) error {
	unbondingTime, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if unbondingTime <= 0 {
		return fmt.Errorf("unbonding time must be greater than zero")
	}

	if unbondingTime < MinUnbondingTime {
		return fmt.Errorf("unbonding time must not be lower %d seconds (6 days)", MinUnbondingTime)
	}

	if unbondingTime > MaxUnbondingTime {
		return fmt.Errorf("unbonding time must not exceed %d seconds (22 days)", MaxUnbondingTime)
	}

	return nil
}
