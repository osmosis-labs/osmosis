package events_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

type SuperfluidEventsTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	addressString = "addr1---------------"
	testDenomA    = "denoma"
	testDenomB    = "denomb"
)

func TestSuperfluidEventsTestSuite(t *testing.T) {
	suite.Run(t, new(SuperfluidEventsTestSuite))
}

func (suite *SuperfluidEventsTestSuite) TestEmitSetSuperfluidAssetEvent() {
	testcases := map[string]struct {
		ctx       sdk.Context
		denom     string
		assetType types.SuperfluidAssetType
	}{
		"basic valid": {
			ctx:       suite.CreateTestContext(),
			denom:     testDenomA,
			assetType: types.SuperfluidAssetTypeNative,
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSetSuperfluidAsset,
					sdk.NewAttribute(types.AttributeDenom, tc.denom),
					sdk.NewAttribute(types.AttributeSuperfluidAssetType, tc.assetType.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSetSuperfluidAssetEvent(tc.ctx, tc.denom, tc.assetType)

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

func (suite *SuperfluidEventsTestSuite) TestEmitRemoveSuperfluidAsset() {
	testcases := map[string]struct {
		ctx   sdk.Context
		denom string
	}{
		"basic valid": {
			ctx:   suite.CreateTestContext(),
			denom: testDenomA,
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtRemoveSuperfluidAsset,
					sdk.NewAttribute(types.AttributeDenom, tc.denom),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitRemoveSuperfluidAsset(tc.ctx, tc.denom)

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

func (suite *SuperfluidEventsTestSuite) TestEmitSuperfluidDelegateEvent() {
	testcases := map[string]struct {
		ctx       sdk.Context
		lockID    uint64
		valAddr   string
		lockCoins sdk.Coins
	}{
		"basic valid": {
			ctx:       suite.CreateTestContext(),
			lockID:    1,
			valAddr:   sdk.AccAddress([]byte(addressString)).String(),
			lockCoins: sdk.NewCoins(sdk.NewInt64Coin("foo", 10)),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSuperfluidDelegateEvent(tc.ctx, tc.lockID, tc.valAddr, tc.lockCoins)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSuperfluidDelegate,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
					sdk.NewAttribute(types.AttributeLockAmount, tc.lockCoins[0].Amount.String()),
					sdk.NewAttribute(types.AttributeLockDenom, tc.lockCoins[0].Denom),
					sdk.NewAttribute(types.AttributeValidator, tc.valAddr),
				),
			}
			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *SuperfluidEventsTestSuite) TestEmitCreateFullRangePositionAndSuperfluidDelegateEvent() {
	testcases := map[string]struct {
		ctx        sdk.Context
		lockID     uint64
		positionID uint64
		valAddr    string
	}{
		"basic valid": {
			ctx:        suite.CreateTestContext(),
			lockID:     1,
			positionID: 1,
			valAddr:    sdk.AccAddress([]byte(addressString)).String(),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtCreateFullRangePositionAndSFDelegate,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
					sdk.NewAttribute(types.AttributePositionId, fmt.Sprintf("%d", tc.positionID)),
					sdk.NewAttribute(types.AttributeValidator, tc.valAddr),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitCreateFullRangePositionAndSuperfluidDelegateEvent(tc.ctx, tc.lockID, tc.positionID, tc.valAddr)

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

func (suite *SuperfluidEventsTestSuite) TestEmitSuperfluidIncreaseDelegationEvent() {
	testcases := map[string]struct {
		ctx    sdk.Context
		lockID uint64
		amount sdk.Coins
	}{
		"basic valid": {
			ctx:    suite.CreateTestContext(),
			lockID: 1,
			amount: sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(100))),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
		"valid with multiple tokens in and out": {
			ctx:    suite.CreateTestContext(),
			lockID: 1,
			amount: sdk.NewCoins(sdk.NewCoin(testDenomA, osmomath.NewInt(100)), sdk.NewCoin(testDenomB, osmomath.NewInt(10))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSuperfluidIncreaseDelegation,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
					sdk.NewAttribute(types.AttributeAmount, tc.amount.String()),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSuperfluidIncreaseDelegationEvent(tc.ctx, tc.lockID, tc.amount)

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

func (suite *SuperfluidEventsTestSuite) TestEmitSuperfluidUndelegateEvent() {
	testcases := map[string]struct {
		ctx       sdk.Context
		lockID    uint64
		lockCoins sdk.Coins
	}{
		"basic valid": {
			ctx:       suite.CreateTestContext(),
			lockID:    1,
			lockCoins: sdk.NewCoins(sdk.NewInt64Coin("foo", 10)),
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSuperfluidUndelegateEvent(tc.ctx, tc.lockID, tc.lockCoins)

			// Assertions
			if hasNoEventManager {
				// If there is no event manager on context, this is a no-op.
				return
			}

			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSuperfluidUndelegate,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
					sdk.NewAttribute(types.AttributeLockAmount, tc.lockCoins[0].Amount.String()),
					sdk.NewAttribute(types.AttributeLockDenom, tc.lockCoins[0].Denom),
				),
			}

			eventManager := tc.ctx.EventManager()
			actualEvents := eventManager.Events()
			suite.Equal(expectedEvents, actualEvents)
		})
	}
}

func (suite *SuperfluidEventsTestSuite) TestEmitSuperfluidUnbondLockEvent() {
	testcases := map[string]struct {
		ctx    sdk.Context
		lockID uint64
	}{
		"basic valid": {
			ctx:    suite.CreateTestContext(),
			lockID: 1,
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSuperfluidUnbondLock,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSuperfluidUnbondLockEvent(tc.ctx, tc.lockID)

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

func (suite *SuperfluidEventsTestSuite) TestEmitSuperfluidUndelegateAndUnbondLockEvent() {
	testcases := map[string]struct {
		ctx    sdk.Context
		lockID uint64
	}{
		"basic valid": {
			ctx:    suite.CreateTestContext(),
			lockID: 1,
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtSuperfluidUndelegateAndUnbondLock,
					sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", tc.lockID)),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitSuperfluidUndelegateAndUnbondLockEvent(tc.ctx, tc.lockID)

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

func (suite *SuperfluidEventsTestSuite) TestEmitUnpoolIdEvent() {
	testAllExitedLockIDsSerialized, _ := json.Marshal([]uint64{1})

	testcases := map[string]struct {
		ctx                        sdk.Context
		sender                     string
		lpShareDenom               string
		allExitedLockIDsSerialized []byte
	}{
		"basic valid": {
			ctx:                        suite.CreateTestContext(),
			sender:                     sdk.AccAddress([]byte(addressString)).String(),
			lpShareDenom:               "pool1",
			allExitedLockIDsSerialized: testAllExitedLockIDsSerialized,
		},
		"context with no event manager": {
			ctx: sdk.Context{},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtUnpoolId,
					sdk.NewAttribute(sdk.AttributeKeySender, tc.sender),
					sdk.NewAttribute(types.AttributeDenom, tc.lpShareDenom),
					sdk.NewAttribute(types.AttributeNewLockIds, string(tc.allExitedLockIDsSerialized)),
				),
			}

			hasNoEventManager := tc.ctx.EventManager() == nil

			// System under test.
			events.EmitUnpoolIdEvent(tc.ctx, tc.sender, tc.lpShareDenom, tc.allExitedLockIDsSerialized)

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
