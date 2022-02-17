package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
)

func (suite *KeeperTestSuite) TestLockupMergreMigration() {
	suite.SetupTest()

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
			amount := rand.Int63n(100000)
			add(addr, denom, baseDuration, amount)
			suite.LockTokens(addr, sdk.Coins{sdk.NewInt64Coin(denom, amount)}, duration)
		}
	}

	suite.Require().NotPanics(func() {
		keeper.MergeLockupsForSimilarDurations(
			suite.ctx, *suite.app.LockupKeeper, suite.app.AccountKeeper,
			keeper.BaselineDurations, keeper.HourDuration,
		)
	})

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			for _, duration := range keeper.BaselineDurations {
				locks := suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr(i), denom(j), duration)
				suite.Require().True(len(locks) <= 1)
				if len(locks) == 1 {
					suite.Require().Equal(locks[0].Coins[0].Amount.Int64(), get(addr(i), denom(j), duration),
						"amount not equal on %s", locks[0],
					)
				}
			}
		}
	}
}
