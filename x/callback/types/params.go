package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultCallbackGasLimit               = uint64(280000000)
	DefaultMaxBlockReservationLimit       = uint64(26)
	DefaultMaxFutureReservationLimit      = uint64(10000)
	DefaultBlockReservationFeeMultiplier  = math.LegacyZeroDec()
	DefaultFutureReservationFeeMultiplier = math.LegacyZeroDec()
	DefaultMinPriceOfGas                  = sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())
)

// NewParams creates a new Params instance.
func NewParams(
	callbackGasLimit uint64,
	maxBlockReservationLimit uint64,
	maxFutureReservationLimit uint64,
	blockReservationFeeMultiplier math.LegacyDec,
	futureReservationFeeMultiplier math.LegacyDec,
	minPriceOfGas sdk.Coin,
) Params {
	return Params{
		CallbackGasLimit:               callbackGasLimit,
		MaxBlockReservationLimit:       maxBlockReservationLimit,
		MaxFutureReservationLimit:      maxFutureReservationLimit,
		BlockReservationFeeMultiplier:  blockReservationFeeMultiplier,
		FutureReservationFeeMultiplier: futureReservationFeeMultiplier,
		MinPriceOfGas:                  minPriceOfGas,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultCallbackGasLimit,
		DefaultMaxBlockReservationLimit,
		DefaultMaxFutureReservationLimit,
		DefaultBlockReservationFeeMultiplier,
		DefaultFutureReservationFeeMultiplier,
		DefaultMinPriceOfGas,
	)
}

// Validate perform object fields validation.
func (p Params) Validate() error {
	if p.CallbackGasLimit == 0 {
		return fmt.Errorf("CallbackGasLimit must be greater than 0")
	}
	if p.BlockReservationFeeMultiplier.IsNegative() {
		return fmt.Errorf("BlockReservationFeeMultiplier must be greater than 0")
	}
	if p.FutureReservationFeeMultiplier.IsNegative() {
		return fmt.Errorf("FutureReservationFeeMultiplier must be greater than 0")
	}
	return nil
}
