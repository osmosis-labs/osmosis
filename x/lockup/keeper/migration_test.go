package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestLockupMergeMigration() {
	s.SetupTest()

	m := make(map[string]int64)
	key := func(addr sdk.AccAddress, denom string, duration time.Duration) string {
		return fmt.Sprintf("%s/%s/%d", string(addr), denom, int64(duration))
	}
	get := func(addr sdk.AccAddress, denom string, duration time.Duration) int64 {
		res, ok := m[key(addr, denom, duration)]
		if !ok {
			return 0
		}
		return res
	}
	add := func(addr sdk.AccAddress, denom string, duration time.Duration, amount int64) {
		m[key(addr, denom, duration)] = get(addr, denom, duration) + amount
	}
	addr := func(i int) sdk.AccAddress {
		return sdk.AccAddress([]byte(fmt.Sprintf("addr%d---------------", i)))
	}
	denom := func(i int) string {
		return fmt.Sprintf("coin%d", i)
	}

	// simulate jitter
	for _, baseDuration := range keeper.BaselineDurations {
		for i := 0; i < 100; i++ {
			addr, denom := addr(rand.Intn(5)), denom(rand.Intn(5))
			duration := baseDuration + time.Duration(rand.Int63n(int64(keeper.HourDuration)))
			if rand.Intn(3) == 0 {
				duration = baseDuration
			}
			amount := rand.Int63n(99999) + 1 // ensure amount is never 0
			add(addr, denom, baseDuration, amount)
			s.LockTokens(addr, sdk.Coins{sdk.NewInt64Coin(denom, amount)}, duration)
		}
	}

	s.Require().NotPanics(func() {
		keeper.MergeLockupsForSimilarDurations(
			s.Ctx, *s.App.LockupKeeper, s.App.AccountKeeper,
			keeper.BaselineDurations, keeper.HourDuration,
		)
	})

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for _, duration := range keeper.BaselineDurations {
				locks := s.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(s.Ctx, addr(i), denom(j), duration)
				s.Require().True(len(locks) <= 1)
				if len(locks) == 1 {
					s.Require().Equal(locks[0].Coins[0].Amount.Int64(), get(addr(i), denom(j), duration),
						"amount not equal on %s", locks[0],
					)
				}
			}
		}
	}
}
