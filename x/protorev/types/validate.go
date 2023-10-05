package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// ---------------------- BaseDenom Validation ---------------------- //
// Validates the base denoms that are used to generate highest liquidity routes.
func (base *BaseDenom) Validate() error {
	if base.Denom == "" {
		return fmt.Errorf("base denom cannot be empty")
	}

	if base.StepSize.IsNil() || base.StepSize.LT(osmomath.OneInt()) {
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

// ---------------------- InfoByPoolType Validation ---------------------- //
// Validates the information about each pool type that is used throughout the module.
func (info *InfoByPoolType) Validate() error {
	if info == nil {
		return fmt.Errorf("pool type info cannot be nil")
	}

	if err := info.Balancer.Validate(); err != nil {
		return err
	}

	if err := info.Stable.Validate(); err != nil {
		return err
	}

	if err := info.Concentrated.Validate(); err != nil {
		return err
	}

	if err := info.Cosmwasm.Validate(); err != nil {
		return err
	}

	return nil
}

// Validates balancer pool information.
func (b *BalancerPoolInfo) Validate() error {
	if b == nil {
		return fmt.Errorf("balancer pool info cannot be nil")
	}

	if b.Weight == 0 {
		return fmt.Errorf("balancer pool weight cannot be 0")
	}

	return nil
}

// Validates stable pool information.
func (s *StablePoolInfo) Validate() error {
	if s == nil {
		return fmt.Errorf("stable pool info cannot be nil")
	}

	if s.Weight == 0 {
		return fmt.Errorf("stable pool weight cannot be 0")
	}

	return nil
}

// Validates concentrated pool information.
func (c *ConcentratedPoolInfo) Validate() error {
	if c == nil {
		return fmt.Errorf("concentrated pool info cannot be nil")
	}

	if c.Weight == 0 {
		return fmt.Errorf("concentrated pool weight cannot be 0")
	}

	if c.MaxTicksCrossed == 0 || c.MaxTicksCrossed > MaxTicksCrossed {
		return fmt.Errorf("max ticks moved cannot be 0 or greater than %d", MaxTicksCrossed)
	}

	return nil
}

// Validates cosmwasm pool information.
func (c *CosmwasmPoolInfo) Validate() error {
	if c == nil {
		return fmt.Errorf("cosmwasm pool info cannot be nil")
	}

	for _, weightMap := range c.WeightMaps {
		address, err := sdk.AccAddressFromBech32(weightMap.ContractAddress)
		if err != nil {
			return err
		}

		if weightMap.Weight == 0 {
			return fmt.Errorf("cosmwasm pool weight cannot be 0 for contract address %s", address)
		}
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
