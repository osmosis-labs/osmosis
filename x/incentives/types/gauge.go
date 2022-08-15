package types

import (
	time "time"

	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// CreateGaugeFee is the fee required to create a new gauge.
	CreateGaugeFee = sdk.NewInt(50 * 1_000_000)
	// AddToGagugeFee is the fee required to add to gauge.
	AddToGaugeFee = sdk.NewInt(25 * 1_000_000)
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

func (gauge Gauge) IsUpcomingGauge(curTime time.Time) bool {
	return curTime.After(gauge.StartTime)
}

func (gauge Gauge) IsActiveGauge(curTime time.Time) bool {
	if curTime.Before(gauge.StartTime) && (gauge.IsPerpetual || gauge.FilledEpochs < gauge.NumEpochsPaidOver) {
		return true
	}
	return false
}

func (gauge Gauge) IsFinishedGauge(curTime time.Time) bool {
	return !gauge.IsUpcomingGauge(curTime) && !gauge.IsActiveGauge(curTime)
}
