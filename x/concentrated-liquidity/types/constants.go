package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	// Precomputed values for min and max tick
	MinInitializedTick, MaxTick int64 = -108000000, 342000000
	// If we consume all liquidity and cross the min initialized tick,
	// our current tick will equal to MinInitializedTick - 1 with zero liquidity.
	// However, note that this tick cannot be crossed. If current tick
	// equals to this tick, it is only possible to swap in the right (one for zero)
	// direction.
	// Note, that this behavior is different from MaxTick since our "active range"
	// invariant is [lower tick, uppper tick). As a result, when we consume all lower
	// tick liquiditty, we must cross it and get kicked out of it.
	MinCurrentTick                int64 = MinInitializedTick - 1
	ExponentAtPriceOne            int64 = -6
	ConcentratedGasFeeForSwap           = 10_000
	BaseGasFeeForNewIncentive           = 10_000
	BaseGasFeeForInitializingTick       = 10_000
)

var (
	MaxSpotPrice       = sdk.MustNewDecFromStr("100000000000000000000000000000000000000")
	MinSpotPrice       = sdk.MustNewDecFromStr("0.000000000001") // 10^-12
	MaxSqrtPrice       = osmomath.MustMonotonicSqrt(MaxSpotPrice)
	MinSqrtPrice       = osmomath.MustMonotonicSqrt(MinSpotPrice)
	MaxSqrtPriceBigDec = osmomath.BigDecFromSDKDec(MaxSqrtPrice)
	MinSqrtPriceBigDec = osmomath.BigDecFromSDKDec(MinSqrtPrice)
	// Supported uptimes preset to 1 ns, 1 min, 1 hr, 1D, 1W, 2W
	SupportedUptimes        = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24, time.Hour * 24 * 7, time.Hour * 24 * 7 * 2}
	AuthorizedTickSpacing   = []uint64{1, 10, 100, 1000}
	AuthorizedSpreadFactors = []sdk.Dec{
		sdk.ZeroDec(),
		sdk.MustNewDecFromStr("0.0001"), // 0.01%
		sdk.MustNewDecFromStr("0.0005"), // 0.05%
		sdk.MustNewDecFromStr("0.001"),  // 0.1%
		sdk.MustNewDecFromStr("0.002"),  // 0.2%
		sdk.MustNewDecFromStr("0.003"),  // 0.3%
		sdk.MustNewDecFromStr("0.005"),  // 0.5%
	}
	DefaultBalancerSharesDiscount = sdk.MustNewDecFromStr("0.05")
	// By default, we only authorize one nanosecond (one block) uptime as an option
	DefaultAuthorizedUptimes = []time.Duration{time.Nanosecond}
)
