package types

import (
	"errors"
	"fmt"
)

// NewGenesisState reeturns genesis state with the given parameters and twap records.
func NewGenesisState(params Params, twapRecords []TwapRecord) *GenesisState {
	return &GenesisState{
		Params: params,
		Twaps:  twapRecords,
	}
}

// DefaultGenesis returns the default twap genesis state.
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []TwapRecord{})
}

// Validate validates the genesis state. Retursn nil on success, error otherwise.
// TODO: test
func (g *GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}

	for _, twap := range g.Twaps {
		if err := twap.validate(); err != nil {
			return err
		}
	}
	return nil
}

// validate validates the twap record, returns nil on success, error otherwise.
// TODO: test
func (t TwapRecord) validate() error {
	if t.PoolId == 0 {
		return errors.New("pool id cannot be 0")
	}

	if t.Asset0Denom == "" {
		return fmt.Errorf("twap record asset0 denom cannot be empty, was (%s)", t.Asset0Denom)
	}

	if t.Asset1Denom == "" {
		return fmt.Errorf("twap record asset1 denom cannot be empty, was (%s)", t.Asset1Denom)
	}

	if t.Height < 0 {
		return fmt.Errorf("twap record height must be positive, was (%d)", t.Height)
	}

	if t.Time.IsZero() {
		return errors.New("twap record time cannot be 0")
	}

	if t.P0LastSpotPrice.IsNil() || !t.P0LastSpotPrice.IsPositive() {
		return fmt.Errorf("twap record p0 last spot price mut be positive, was (%s)", t.P0LastSpotPrice)
	}

	if t.P1LastSpotPrice.IsNil() || !t.P1LastSpotPrice.IsPositive() {
		return fmt.Errorf("twap record p1 last spot price mut be positive, was (%s)", t.P1LastSpotPrice)
	}

	if t.P0ArithmeticTwapAccumulator.IsNil() || !t.P0ArithmeticTwapAccumulator.IsPositive() {
		return fmt.Errorf("twap record p0 accumulator mut be positive, was (%s)", t.P0ArithmeticTwapAccumulator)
	}

	if t.P1ArithmeticTwapAccumulator.IsNil() || !t.P1ArithmeticTwapAccumulator.IsPositive() {
		return fmt.Errorf("twap record p1 accumulator mut be positive, was (%s)", t.P1ArithmeticTwapAccumulator)
	}
	return nil
}
