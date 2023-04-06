package types

import (
	"errors"
	"fmt"
)

// NewGenesisState returns genesis state with the given parameters and twap records.
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

// Validate validates the genesis state. Returns nil on success, error otherwise.
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

	if t.Height <= 0 {
		return fmt.Errorf("twap record height must be positive, was (%d)", t.Height)
	}

	if t.Time.IsZero() {
		return errors.New("twap record time cannot be 0")
	}

	// if there was an error in this record, the spot prices should be 0.
	// else, the the spot prices must be positive.
	if t.LastErrorTime.Equal(t.Time) {
		if t.P0LastSpotPrice.IsNil() || !t.P0LastSpotPrice.IsZero() {
			return fmt.Errorf("twap record p0 last spot price must be zero due to having an error, was (%s)", t.P0LastSpotPrice)
		}

		if t.P1LastSpotPrice.IsNil() || !t.P1LastSpotPrice.IsZero() {
			return fmt.Errorf("twap record p1 last spot price must be zero due to having an error, was (%s)", t.P1LastSpotPrice)
		}
	} else {
		if t.P0LastSpotPrice.IsNil() || !t.P0LastSpotPrice.IsPositive() {
			return fmt.Errorf("twap record p0 last spot price must be positive, was (%s)", t.P0LastSpotPrice)
		}

		if t.P1LastSpotPrice.IsNil() || !t.P1LastSpotPrice.IsPositive() {
			return fmt.Errorf("twap record p1 last spot price must be positive, was (%s)", t.P1LastSpotPrice)
		}
	}

	if t.P0ArithmeticTwapAccumulator.IsNil() || t.P0ArithmeticTwapAccumulator.IsNegative() {
		return fmt.Errorf("twap record p0 accumulator cannot be negative, was (%s)", t.P0ArithmeticTwapAccumulator)
	}

	if t.P1ArithmeticTwapAccumulator.IsNil() || t.P1ArithmeticTwapAccumulator.IsNegative() {
		return fmt.Errorf("twap record p1 accumulator cannot be negative, was (%s)", t.P1ArithmeticTwapAccumulator)
	}

	if t.GeometricTwapAccumulator.IsNil() {
		return fmt.Errorf("twap record geometric accumulator cannot be nil, was (%s)", t.GeometricTwapAccumulator)
	}
	return nil
}
