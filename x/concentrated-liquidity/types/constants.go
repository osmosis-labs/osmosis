package types

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	// Precomputed values for min and max tick

	// Tick corresponding to the at launch min spot price of 10^-12.
	MinInitializedTick, MaxTick int64 = -108000000, 342000000
	// If we consume all liquidity and cross the min initialized tick,
	// our current tick will equal to MinInitializedTick - 1 with zero liquidity.
	// However, note that this tick cannot be crossed. If current tick
	// equals to this tick, it is only possible to swap in the right (one for zero)
	// direction.
	// Note, that this behavior is different from MaxTick since our "active range"
	// invariant is [lower tick, uppper tick). As a result, when we consume all lower
	// tick liquiditty, we must cross it and get kicked out of it.
	MinCurrentTick int64 = MinInitializedTick - 1
	// Tick corresponding to the extended min spot price of 10^-30.
	MinInitializedTickV2          int64 = -270000000
	MinCurrentTickV2              int64 = MinInitializedTickV2 - 1
	ExponentAtPriceOne            int64 = -6
	ConcentratedGasFeeForSwap           = 10_000
	BaseGasFeeForNewIncentive           = 10_000
	BaseGasFeeForInitializingTick       = 10_000
	BaseGasFeeForTransferPosition       = 10_000
)

var (
	MaxSpotPrice       = osmomath.MustNewDecFromStr("100000000000000000000000000000000000000")
	MaxSpotPriceBigDec = osmomath.BigDecFromDec(MaxSpotPrice)
	// TODO: remove when https://github.com/osmosis-labs/osmosis/issues/5726 is complete.
	MinSpotPrice = osmomath.MustNewDecFromStr("0.000000000001") // 10^-12
	// Note: this is the at launch min spot price that is getting lowered to 10^-30
	MinSpotPriceBigDec = osmomath.BigDecFromDec(MinSpotPrice)
	MinSpotPriceV2     = osmomath.NewBigDecWithPrec(1, 30)
	MaxSqrtPrice       = osmomath.MustMonotonicSqrt(MaxSpotPrice)
	MinSqrtPrice       = osmomath.MustMonotonicSqrt(MinSpotPrice)
	MaxSqrtPriceBigDec = osmomath.BigDecFromDec(MaxSqrtPrice)
	MinSqrtPriceBigDec = osmomath.BigDecFromDec(MinSqrtPrice)

	// Supported uptimes preset to 1 ns, 1 min, 1 hr, 1D, 1W, 2W
	SupportedUptimes        = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24, time.Hour * 24 * 7, time.Hour * 24 * 7 * 2}
	AuthorizedTickSpacing   = []uint64{1, 10, 100, 1000}
	AuthorizedSpreadFactors = []osmomath.Dec{
		osmomath.ZeroDec(),
		osmomath.MustNewDecFromStr("0.0001"), // 0.01%
		osmomath.MustNewDecFromStr("0.0005"), // 0.05%
		osmomath.MustNewDecFromStr("0.001"),  // 0.1%
		osmomath.MustNewDecFromStr("0.002"),  // 0.2%
		osmomath.MustNewDecFromStr("0.003"),  // 0.3%
		osmomath.MustNewDecFromStr("0.005"),  // 0.5%
	}
	DefaultBalancerSharesDiscount = osmomath.MustNewDecFromStr("0.05")
	// By default, we only authorize one nanosecond (one block) uptime as an option
	DefaultAuthorizedUptimes                = []time.Duration{time.Nanosecond}
	DefaultUnrestrictedPoolCreatorWhitelist = []string{}
	// This is a (very generous) gas limit intended to protect against CL hooks that are
	// executed with malicious intent in begin block code.
	//
	// This is an unlikely scenario as long as contract deployment is gated by governance
	// and protorev/txfee swaps are on whitelisted tokens, but letting contract calls take
	// unbounded gas in these scenarios is a risk we don't want to take regardless.
	//
	// 2M gas is enough to execute tens of expensive CL operations and is only set this high
	// to accommodate position withdrawals, which are unusually expensive.
	DefaultContractHookGasLimit = uint64(2_000_000)
)
