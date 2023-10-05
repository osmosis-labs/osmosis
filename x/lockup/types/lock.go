package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// defaultOwnerReceiverPlaceholder is used as a place holder for default owner receiver for the
// reward receiver field.
// Using this as the value for reward receiver would indicate that the lock's reward receiver is the owner.
const DefaultOwnerReceiverPlaceholder = ""

// NewPeriodLock returns a new instance of period lock.
func NewPeriodLock(ID uint64, owner sdk.AccAddress, reward_address string, duration time.Duration, endTime time.Time, coins sdk.Coins) PeriodLock {
	// sanity check once more to ensure if reward_address == owner, we store empty string
	if owner.String() == reward_address {
		reward_address = DefaultOwnerReceiverPlaceholder
	}
	return PeriodLock{
		ID:                    ID,
		Owner:                 owner.String(),
		RewardReceiverAddress: reward_address,
		Duration:              duration,
		EndTime:               endTime,
		Coins:                 coins,
	}
}

// IsUnlocking returns lock started unlocking already.
func (p PeriodLock) IsUnlocking() bool {
	return !p.EndTime.Equal(time.Time{})
}

// IsUnlocking returns lock started unlocking already.
func (p SyntheticLock) IsUnlocking() bool {
	return !p.EndTime.Equal(time.Time{})
}

// IsNil returns if the synthetic lock is nil.
func (p SyntheticLock) IsNil() bool {
	return p == (SyntheticLock{})
}

// OwnerAddress returns locks owner address.
func (p PeriodLock) OwnerAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(p.Owner)
	if err != nil {
		panic(err)
	}
	return addr
}

func (p PeriodLock) SingleCoin() (sdk.Coin, error) {
	if len(p.Coins) != 1 {
		return sdk.Coin{}, fmt.Errorf("PeriodLock %d has no single coin: %s", p.ID, p.Coins)
	}
	return p.Coins[0], nil
}

func SumLocksByDenom(locks []PeriodLock, denom string) osmomath.Int {
	sum := osmomath.NewInt(0)
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

// quick fix for getting native denom from synthetic denom.
func NativeDenom(denom string) string {
	if strings.Contains(denom, "/superbonding") {
		return strings.Split(denom, "/superbonding")[0]
	}
	if strings.Contains(denom, "/superunbonding") {
		return strings.Split(denom, "/superunbonding")[0]
	}
	return denom
}

func IsSyntheticDenom(denom string) bool {
	return NativeDenom(denom) != denom
}
