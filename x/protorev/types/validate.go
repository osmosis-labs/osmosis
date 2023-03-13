package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------------------- BaseDenom Validation ---------------------- //
// Validates the base denoms that are used to generate highest liquidity routes.
func (base *BaseDenom) Validate() error {
	if base.Denom == "" {
		return fmt.Errorf("base denom cannot be empty")
	}

	if base.StepSize.IsNil() || base.StepSize.LT(sdk.OneInt()) {
		return fmt.Errorf("step size must be greater than 0")
	}

	return nil
}

// ValidateBaseDenoms validates the base denoms that are used to generate highest liquidity routes.
func ValidateBaseDenoms(denoms []BaseDenom) error {
	// The first base denom must be the Osmosis denomination
	if len(denoms) == 0 || denoms[0].Denom != OsmosisDenomination {
		return fmt.Errorf("the first base denom must be the Osmosis denomination")
	}

	seenDenoms := make(map[string]bool)
	for _, denom := range denoms {
		if err := denom.Validate(); err != nil {
			return err
		}

		// Ensure that the base denom is unique
		if seenDenoms[denom.Denom] {
			return fmt.Errorf("duplicate base denom %s", denom)
		}
		seenDenoms[denom.Denom] = true
	}
	return nil
}

// ---------------------- PoolWeights Validation ---------------------- //
// Validates that the pool weights object is ready for use in the module.
func (pw *PoolWeights) Validate() error {
	if pw == nil {
		return fmt.Errorf("pool weights cannot be nil")
	}

	if pw.BalancerWeight == 0 || pw.StableWeight == 0 || pw.ConcentratedWeight == 0 {
		return fmt.Errorf("pool weights cannot be 0")
	}

	return nil
}

// ---------------------- DeveloperFee Validation ---------------------- //
// ValidateDeveloperFees does some basic validation on the developer fees passed into the module genesis.
func ValidateDeveloperFees(fees []sdk.Coin) error {
	seenDenoms := make(map[string]bool)
	for _, fee := range fees {
		if err := fee.Validate(); err != nil {
			return err
		}

		// Ensure that the developer fee is unique
		if seenDenoms[fee.Denom] {
			return fmt.Errorf("duplicate developer fee %s", fee)
		}
		seenDenoms[fee.Denom] = true
	}
	return nil
}

// ---------------------- Pool Point Validation ---------------------- //
// ValidateMaxPoolPointsPerBlock validates the max pool points per block.
func ValidateMaxPoolPointsPerBlock(points uint64) error {
	if points == 0 || points > MaxPoolPointsPerBlock {
		return fmt.Errorf("max pool points per block must be between 1 and %d", MaxPoolPointsPerBlock)
	}

	return nil
}

// ValidateMaxPoolPointsPerTx validates the max pool points per tx.
func ValidateMaxPoolPointsPerTx(points uint64) error {
	if points == 0 || points > MaxPoolPointsPerTx {
		return fmt.Errorf("max pool points per tx must be between 1 and %d", MaxPoolPointsPerTx)
	}

	return nil
}
