package types

import (
	time "time"

	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewPot(id uint64, isPerpetual bool, distrTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64, filledEpochs uint64, distrCoins sdk.Coins) Pot {
	return Pot{
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

func (pot Pot) IsUpcomingPot(curTime time.Time) bool {
	if curTime.After(pot.StartTime) {
		return true
	}
	return false
}

func (pot Pot) IsActivePot(curTime time.Time) bool {
	if curTime.Before(pot.StartTime) && (pot.IsPerpetual || pot.FilledEpochs < pot.NumEpochsPaidOver) {
		return true
	}
	return false
}

func (pot Pot) IsFinishedPot(curTime time.Time) bool {
	return !pot.IsUpcomingPot(curTime) && !pot.IsActivePot(curTime)
}
