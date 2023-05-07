package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Precomputed values for min and max tick
	MinTick, MaxTick int64 = -162000000, 342000000
)

var (
	ConcentratedGasFeeForSwap = 10_000
	MaxSpotPrice              = sdk.MustNewDecFromStr("100000000000000000000000000000000000000")
	MinSpotPrice              = sdk.MustNewDecFromStr("0.000000000000000001")
	MaxSqrtPrice, _           = MaxSpotPrice.ApproxRoot(2)
	MinSqrtPrice, _           = MinSpotPrice.ApproxRoot(2)
	// Supported uptimes preset to 1 ns, 1 min, 1 hr, 1D, 1W
	SupportedUptimes      = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24, time.Hour * 24 * 7}
	ExponentAtPriceOne    = sdk.NewInt(-6)
	AuthorizedTickSpacing = []uint64{1, 10, 100, 1000}
	AuthorizedSwapFees    = []sdk.Dec{
		sdk.ZeroDec(),
		sdk.MustNewDecFromStr("0.0001"), // 0.01%
		sdk.MustNewDecFromStr("0.0005"), // 0.05%
		sdk.MustNewDecFromStr("0.001"),  // 0.1%
		sdk.MustNewDecFromStr("0.002"),  // 0.2%
		sdk.MustNewDecFromStr("0.003"),  // 0.3%
		sdk.MustNewDecFromStr("0.005")}  // 0.5%
	BaseGasFeeForNewIncentive     = 10_000
	DefaultBalancerSharesDiscount = sdk.MustNewDecFromStr("0.05")
	// By default, we only authorize one nanosecond (one block) uptime as an option
	DefaultAuthorizedUptimes      = []time.Duration{time.Nanosecond}
	BaseGasFeeForInitializingTick = 10_000
)
