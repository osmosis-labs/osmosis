package types

import (
	time "time"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// CreateGaugeFee is the fee required to create a new gauge.
	CreateGaugeFee = osmomath.NewInt(50 * 1_000_000)
	// AddToGagugeFee is the fee required to add to gauge.
	AddToGaugeFee = osmomath.NewInt(25 * 1_000_000)
)

// NewGauge creates a new gauge struct given the required gauge parameters.
func NewGauge(id uint64, isPerpetual bool, distrTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64, filledEpochs uint64, distrCoins sdk.Coins) Gauge {
	return Gauge{
		Id:                id,
		IsPerpetual:       isPerpetual,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      filledEpochs,
		DistributedCoins:  distrCoins,
	}
}

// IsUpcomingGauge returns true if the gauge's distribution start time is after the provided time.
func (gauge Gauge) IsUpcomingGauge(curTime time.Time) bool {
	return curTime.Before(gauge.StartTime)
}

// IsActiveGauge returns true if the gauge is in an active state during the provided time.
func (gauge Gauge) IsActiveGauge(curTime time.Time) bool {
	if (curTime.After(gauge.StartTime) || curTime.Equal(gauge.StartTime)) && (gauge.IsPerpetual || gauge.FilledEpochs < gauge.NumEpochsPaidOver) {
		return true
	}
	return false
}

// IsFinishedGauge returns true if the gauge is in a finished state during the provided time.
func (gauge Gauge) IsFinishedGauge(curTime time.Time) bool {
	return !gauge.IsUpcomingGauge(curTime) && !gauge.IsActiveGauge(curTime)
}

// IsLastNonPerpetualDistribution returns true if the this is the last distribution of the gauge.
// The last distribution is defined for non-perpetual gauges where FilledEpochs+1 >= NumEpochsPaidOver.
// Assumes that this is called before updating the gauge's state at the end of the epoch.
// If called after update, it still safe because of >= comparison.
func (gauge Gauge) IsLastNonPerpetualDistribution() bool {
	// Note, it is impossible to create a gauge with NumEpochsPaidOver == 0 due to a check
	// in MsgCreateGauge.ValidateBasic. Additionally, FilledEpoch is always initialized to 0
	// at gauge creation time.
	return !gauge.IsPerpetual && gauge.FilledEpochs+1 >= gauge.NumEpochsPaidOver
}
