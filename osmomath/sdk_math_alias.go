// This file creates type and function aliases the sdkmath.LegacyDec
// This is done for reducing verbosity and improving readability.
//
// For consistency, we also alias Int and Uint so that sdkmath does not have
// to be directly imported in files where both decimal and integer types are used.
package osmomath

import (
	sdkmath "cosmossdk.io/math"
)

type (
	Dec  = sdkmath.LegacyDec
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

const (
	PrecisionDec         = sdkmath.LegacyPrecision
	DecimalPrecisionBits = sdkmath.LegacyDecimalPrecisionBits
)

var (
	// Dec
	NewDec                   = sdkmath.LegacyNewDec
	NewDecWithPrec           = sdkmath.LegacyNewDecWithPrec
	NewDecFromBigInt         = sdkmath.LegacyNewDecFromBigInt
	NewDecFromBigIntWithPrec = sdkmath.LegacyNewDecFromBigIntWithPrec
	NewDecFromInt            = sdkmath.LegacyNewDecFromInt
	NewDecFromIntWithPrec    = sdkmath.LegacyNewDecFromIntWithPrec
	NewDecFromStr            = sdkmath.LegacyNewDecFromStr
	MustNewDecFromStr        = sdkmath.LegacyMustNewDecFromStr
	ZeroDec                  = sdkmath.LegacyZeroDec
	OneDec                   = sdkmath.LegacyOneDec
	SmallestDec              = sdkmath.LegacySmallestDec

	// Int
	NewInt            = sdkmath.NewInt
	NewIntFromUint64  = sdkmath.NewIntFromUint64
	NewIntFromBigInt  = sdkmath.NewIntFromBigInt
	NewIntFromString  = sdkmath.NewIntFromString
	NewIntWithDecimal = sdkmath.NewIntWithDecimal
	ZeroInt           = sdkmath.ZeroInt
	OneInt            = sdkmath.OneInt
	IntEq             = sdkmath.IntEq
	MinInt            = sdkmath.MinInt
	MaxInt            = sdkmath.MaxInt

	// Uint
	NewUint           = sdkmath.NewUint
	NewUintFromString = sdkmath.NewUintFromString

	MinDec = sdkmath.LegacyMinDec
	MaxDec = sdkmath.LegacyMaxDec
)
