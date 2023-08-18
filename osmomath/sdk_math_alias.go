// This file creates type and function aliases the sdk.Dec, sdk.Int and sdk.Uint types.
// This is done so that we can easily swap out the sdk.Dec with cosmossdk.io/math.LegacyDec
// once ready.
package osmomath

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	SDKDec  = sdk.Dec
	SDKInt  = sdk.Int
	SDKUint = sdk.Uint
)

const (
	PrecisionSDKDec         = sdk.Precision
	SDKDecimalPrecisionBits = sdk.DecimalPrecisionBits
)

var (
	// Dec
	NewSDKDec                   = sdk.NewDec
	NewSDKDecWithPrec           = sdk.NewDecWithPrec
	NewSDKDecFromBigInt         = sdk.NewDecFromBigInt
	NewSDKDecFromBigIntWithPrec = sdk.NewDecFromBigIntWithPrec
	NewSDKDecFromInt            = sdk.NewDecFromInt
	NewSDKDecFromIntWithPrec    = sdk.NewDecFromIntWithPrec
	NewSDKDecFromStr            = sdk.NewDecFromStr
	MustNewSDKDecFromStr        = sdk.MustNewDecFromStr
	ZeroSDKDec                  = sdk.ZeroDec
	OneSDKDec                   = sdk.OneDec
	SDKSmallestDec              = sdk.SmallestDec

	// Int
	NewSDKInt            = sdk.NewInt
	NewSDKIntFromUint64  = sdk.NewIntFromUint64
	NewSDKIntFromBigInt  = sdk.NewIntFromBigInt
	NewSDKIntFromString  = sdk.NewIntFromString
	NewSDKIntWithDecimal = sdk.NewIntWithDecimal
	ZeroSDKInt           = sdk.ZeroInt
	OneSDKInt            = sdk.OneInt
	SDKIntEq             = sdk.IntEq
	MinSDKInt            = sdk.MinInt

	// Uint
	NewSDKUint = sdk.NewUint
)
