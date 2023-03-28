package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Precomputed values for min and max ticks
	MinTickNegTwelve, MaxTickNegTwelve int64 = -162000000000000, 342000000000000
	MinTickNegEleven, MaxTickNegEleven int64 = -16200000000000, 34200000000000
	MinTickNegTen, MaxTickNegTen       int64 = -1620000000000, 3420000000000
	MinTickNegNine, MaxTickNegNine     int64 = -162000000000, 342000000000
	MinTickNegEight, MaxTickNegEight   int64 = -16200000000, 34200000000
	MinTickNegSeven, MaxTickNegSeven   int64 = -1620000000, 3420000000
	MinTickNegSix, MaxTickNegSix       int64 = -162000000, 342000000
	MinTickNegFive, MaxTickNegFive     int64 = -16200000, 34200000
	MinTickNegFour, MaxTickNegFour     int64 = -1620000, 3420000
	MinTickNegThree, MaxTickNegThree   int64 = -162000, 342000
	MinTickNegTwo, MaxTickNegTwo       int64 = -16200, 34200
	MinTickNegOne, MaxTickNegOne       int64 = -1620, 3420
)

var (
	ConcentratedGasFeeForSwap = 10_000
	ExponentAtPriceOneMax     = sdk.NewInt(-1)
	ExponentAtPriceOneMin     = sdk.NewInt(-12)
	MaxSpotPrice              = sdk.MustNewDecFromStr("100000000000000000000000000000000000000")
	MinSpotPrice              = sdk.MustNewDecFromStr("0.000000000000000001")
	MaxSqrtPrice, _           = MaxSpotPrice.ApproxRoot(2)
	MinSqrtPrice, _           = MinSpotPrice.ApproxRoot(2)
	// Supported uptimes preset to 1 ns, 1 min, 1 hr, 1D, 1W
	SupportedUptimes          = []time.Duration{time.Nanosecond, time.Minute, time.Hour, time.Hour * 24, time.Hour * 24 * 7}
	AuthorizedTickSpacing     = []uint64{1, 10, 60, 200}
	BaseGasFeeForNewIncentive = 10_000
)
