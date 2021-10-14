package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewPeriodLock returns a new instance of period lock
func NewPeriodLock(ID uint64, owner sdk.AccAddress, duration time.Duration, endTime time.Time, coins sdk.Coins) PeriodLock {
	return PeriodLock{
		ID:       ID,
		Owner:    owner.String(),
		Duration: duration,
		EndTime:  endTime,
		Coins:    coins,
	}
}

// IsUnlocking returns lock started unlocking already
func (p PeriodLock) IsUnlocking() bool {
	return !p.EndTime.Equal(time.Time{})
}

func SumLocksByDenom(locks []PeriodLock, denom string) sdk.Int {
	sum := sdk.NewInt(0)
	err := sdk.ValidateDenom(denom)
	if err != nil {
		panic(fmt.Errorf("invalid denom used internally: %s, %v", denom, err))
	}
	for _, lock := range locks {
		sum = sum.Add(lock.Coins.AmountOfNoDenomValidation(denom))
	}
	return sum
}
