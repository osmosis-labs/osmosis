package types

import (
	"fmt"
	"strings"
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

// IsUnlocking returns lock started unlocking already
func (p SyntheticLock) IsUnlocking() bool {
	return !p.EndTime.Equal(time.Time{})
}

func (p PeriodLock) SingleCoin() (sdk.Coin, error) {
	if len(p.Coins) != 1 {
		return sdk.Coin{}, fmt.Errorf("PeriodLock %d has no single coin: %s", p.ID, p.Coins)
	}
	return p.Coins[0], nil
}

func (s SyntheticLock) Coin(p PeriodLock) (sdk.Coin, error) {
	// sanity check
	if s.UnderlyingLockId != p.ID {
		panic(fmt.Sprintf("invalid argument on SyntheticLock.Coin: %s, %d, %d", s.SynthDenom, s.UnderlyingLockId, p.ID))
	}

	coin, err := p.SingleCoin()
	if err != nil {
		return sdk.Coin{}, err
	}
	return sdk.Coin{s.SynthDenom, coin.Amount}, nil
}

func SumLocksByDenom(locks []PeriodLock, denom string) sdk.Int {
	sum := sdk.NewInt(0)
	// validate the denom once, so we can avoid the expensive validate check in the hot loop.
	err := sdk.ValidateDenom(denom)
	if err != nil {
		panic(fmt.Errorf("invalid denom used internally: %s, %v", denom, err))
	}
	for _, lock := range locks {
		sum = sum.Add(lock.Coins.AmountOfNoDenomValidation(denom))
	}
	return sum
}

// quick fix for getting native denom from synthetic denom
func NativeDenom(denom string) string {
	if strings.Contains(denom, "superbonding") {
		return strings.Split(denom, "superbonding")[0]
	}
	if strings.Contains(denom, "superunbonding") {
		return strings.Split(denom, "superunbonding")[0]
	}
	return denom
}

// func SyntheticSuffix(denom string) string {
// 	return strings.TrimLeft(denom, NativeDenom(denom))
// }

// func IsSyntheticDenom(denom string) bool {
// 	return NativeDenom(denom) != denom
// }

// func IsStakingSuffix(suffix string) bool {
// 	return strings.Contains(suffix, "superbonding")
// }

// func IsUnstakingSuffix(suffix string) bool {
// 	return strings.Contains(suffix, "superunbonding")
// }
