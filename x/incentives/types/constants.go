package types

import (
	time "time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

var (
	BaseGasFeeForCreateGauge      = 10_000
	BaseGasFeeForAddRewardToGauge = 10_000
	// We set the default value to 1ns, as this is the only uptime we support as long as charging is disabled (or
	// until more supported uptimes are authorized by governance).
	DefaultConcentratedUptime = time.Nanosecond

	// PerpetualNumEpochsPaidOver is the number of epochs that must be given
	// for a gauge to be perpetual. For any other number of epochs
	// other than zero, the gauge is non-perpetual. Zero is invalid.
	PerpetualNumEpochsPaidOver = uint64(0)
	DefaultMinValueForDistr    = sdk.NewCoin(appparams.BaseCoinUnit, sdkmath.NewInt(10000)) // 0.01 OSMO
)
