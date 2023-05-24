package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	OneShareExponent = 18
	// Raise 10 to the power of SigFigsExponent to determine number of significant figures.
	// i.e. SigFigExponent = 8 is 10^8 which is 100000000. This gives 8 significant figures.
	SigFigsExponent       = 8
	BalancerGasFeeForSwap = 10_000

	StableswapMinScaledAmtPerAsset = 1
	// We keep this multiplier at 1, but can increase if needed in the unlikely scenario where default scaling factors of 1 cannot accommodate enough assets
	ScalingFactorMultiplier = 1

	// pools can be created with min and max number of assets defined with this constants
	MinNumOfAssetsInPool = 2
	MaxNumOfAssetsInPool = 8
)

var (
	// OneShare represents the amount of subshares in a single pool share.
	OneShare = sdk.NewIntWithDecimal(1, OneShareExponent)

	// InitPoolSharesSupply is the amount of new shares to initialize a pool with.
	InitPoolSharesSupply = OneShare.MulRaw(100)

	// SpotPriceSigFigs is the amount of significant figures used in return value of calculate SpotPrice
	SpotPriceSigFigs = sdk.NewDec(10).Power(SigFigsExponent).TruncateInt()
	// MaxSpotPrice is the maximum supported spot price. Anything greater than this will error.
	// Internal note: Ctrl+F for MaxSpotPrice in code if ever changed.
	// Other tests depend on being equal to MaxSpotPrice,
	// but don't directly import it due to import issues.
	MaxSpotPrice = sdk.NewDec(2).Power(128).Sub(sdk.OneDec())
	// MinSpotPrice is the minimum supported spot price. Anything less than this will error.
	// It is limited by sdk.Dec's precision.
	MinSpotPrice = sdk.SmallestDec()

	// MultihopSpreadFactorMultiplierForOsmoPools if a spread factor multiplier for trades consists of just two OSMO pools during a single transaction.
	MultihopSpreadFactorMultiplierForOsmoPools = sdk.NewDecWithPrec(5, 1) // 0.5

	// Maximum amount per asset after the application of scaling factors should be 10e34.
	StableswapMaxScaledAmtPerAsset = sdk.NewDec(10).Power(34).TruncateInt()
)
