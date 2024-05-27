package types

import (
	"fmt"
	"math/big"
	"math/bits"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"

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

// TODO: Can we use sumtree instead here?
// Assumes that caller is passing in locks that contain denom
func SumLocksByDenom(locks []*PeriodLock, denom string) (osmomath.Int, error) {
	sumBi := big.NewInt(0)
	// validate the denom once, so we can avoid the expensive validate check in the hot loop.
	err := sdk.ValidateDenom(denom)
	if err != nil {
		return osmomath.Int{}, fmt.Errorf("invalid denom used internally: %s, %v", denom, err)
	}
	for _, lock := range locks {
		var amt osmomath.Int
		// skip a 1second cumulative runtimeEq check
		if len(lock.Coins) == 1 {
			amt = lock.Coins[0].Amount
		} else {
			amt = lock.Coins.AmountOfNoDenomValidation(denom)
		}
		sumBi.Add(sumBi, amt.BigIntMut())
	}

	// handle overflow check here so we don't panic.
	err = checkBigInt(sumBi)
	if err != nil {
		return osmomath.ZeroInt(), err
	}
	return osmomath.NewIntFromBigInt(sumBi), nil
}

// Max number of words a sdk.Int's big.Int can contain.
// This is predicated on MaxBitLen being divisible by 64
var maxWordLen = sdkmath.MaxBitLen / bits.UintSize

// check if a bigInt would overflow max sdk.Int. If it does, return an error.
func checkBigInt(bi *big.Int) error {
	if len(bi.Bits()) > maxWordLen {
		if bi.BitLen() > sdkmath.MaxBitLen {
			return fmt.Errorf("bigInt overflow")
		}
	}
	return nil
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
