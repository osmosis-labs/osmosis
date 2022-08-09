package events_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/utils"
	"github.com/osmosis-labs/osmosis/v10/x/lockup/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	"github.com/stretchr/testify/suite"
)

type LockupEventsTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	addressString = "addr1---------------"
	testDenomA    = "denoma"
)

func TestLockupEventsTestSuite(t *testing.T) {
	suite.Run(t, new(LockupEventsTestSuite))
}

func (suite *LockupEventsTestSuite) TestEmitLockTokenEvent() {
	testcases := map[string]struct {
		ctx  sdk.Context
		lock types.PeriodLock
	}{
		"basic valid": {
			ctx: suite.CreateTestContext(),
			lock: types.PeriodLock{
				ID:       1,
				Owner:    sdk.AccAddress([]byte(addressString)).String(),
				Duration: time.Second,
				EndTime:  time.Time{},
				Coins:    sdk.Coins{sdk.NewInt64Coin(testDenomA, 10)},
			},
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtLockTokens,
					sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(tc.lock.ID)),
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.lock.Owner),
					sdk.NewAttribute(types.AttributePeriodLockAmount, tc.lock.Coins.String()),
					sdk.NewAttribute(types.AttributePeriodLockDuration, tc.lock.Duration.String()),
					sdk.NewAttribute(types.AttributePeriodLockUnlockTime, tc.lock.EndTime.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitLockToken(tc.ctx, &tc.lock)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *LockupEventsTestSuite) TestEmitExtendLockToken() {
	testcases := map[string]struct {
		ctx  sdk.Context
		lock types.PeriodLock
	}{
		"basic valid": {
			ctx: suite.CreateTestContext(),
			lock: types.PeriodLock{
				ID:       1,
				Owner:    sdk.AccAddress([]byte(addressString)).String(),
				Duration: time.Second,
			},
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtLockTokens,
					sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(tc.lock.ID)),
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.lock.Owner),
					sdk.NewAttribute(types.AttributePeriodLockDuration, tc.lock.Duration.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitExtendLockToken(tc.ctx, &tc.lock)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *LockupEventsTestSuite) TestEmitAddTokenToLock() {
	testcases := map[string]struct {
		ctx    sdk.Context
		lockId uint64
		owner  string
		coins  sdk.Coins
	}{
		"basic valid": {
			ctx:    suite.CreateTestContext(),
			lockId: 1,
			owner:  sdk.AccAddress([]byte(addressString)).String(),
			coins:  sdk.NewCoins(sdk.NewCoin(testDenomA, sdk.NewInt(100))),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtAddTokensToLock,
					sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(tc.lockId)),
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.owner),
					sdk.NewAttribute(types.AttributePeriodLockAmount, tc.coins.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitAddTokenToLock(tc.ctx, tc.lockId, tc.owner, tc.coins.String())

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *LockupEventsTestSuite) TestEmitBeginUnlock() {
	testcases := map[string]struct {
		ctx  sdk.Context
		lock types.PeriodLock
	}{
		"basic valid": {
			ctx: suite.CreateTestContext(),
			lock: types.PeriodLock{
				ID:       1,
				Owner:    sdk.AccAddress([]byte(addressString)).String(),
				Duration: time.Second,
				EndTime:  time.Time{},
			},
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtBeginUnlock,
					sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(tc.lock.ID)),
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.lock.Owner),
					sdk.NewAttribute(types.AttributePeriodLockDuration, tc.lock.Duration.String()),
					sdk.NewAttribute(types.AttributePeriodLockUnlockTime, tc.lock.EndTime.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitBeginUnlock(tc.ctx, &tc.lock)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *LockupEventsTestSuite) TestEmitBeginUnlockAll() {
	testcases := map[string]struct {
		ctx           sdk.Context
		unlockedCoins sdk.Coins
		owner         string
	}{
		"basic valid": {
			ctx:           suite.CreateTestContext(),
			unlockedCoins: sdk.NewCoins(sdk.NewCoin(testDenomA, sdk.NewInt(100))),
			owner:         sdk.AccAddress([]byte(addressString)).String(),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtBeginUnlockAll,
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.owner),
					sdk.NewAttribute(types.AttributeUnlockedCoins, tc.unlockedCoins.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitBeginUnlockAll(tc.ctx, tc.unlockedCoins.String(), tc.owner)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}
