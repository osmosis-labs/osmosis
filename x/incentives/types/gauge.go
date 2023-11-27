package types

import (
	"math/big"
	time "time"

	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// 1 DYM
	DYM = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	// CreateGaugeFee is the fee required to create a new gauge.
	CreateGaugeFee = DYM.Mul(sdk.NewInt(50))
	// AddToGagugeFee is the fee required to add to gauge.
	AddToGaugeFee = DYM.Mul(sdk.NewInt(25))
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
	if curTime.After(gauge.StartTime) || curTime.Equal(gauge.StartTime) && (gauge.IsPerpetual || gauge.FilledEpochs < gauge.NumEpochsPaidOver) {
		return true
	}
	return false
}

// IsFinishedGauge returns true if the gauge is in a finished state during the provided time.
func (gauge Gauge) IsFinishedGauge(curTime time.Time) bool {
	return !gauge.IsUpcomingGauge(curTime) && !gauge.IsActiveGauge(curTime)
}
