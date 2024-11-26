package events

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

func EmitSetSuperfluidAssetEvent(ctx sdk.Context, denom string, assetType types.SuperfluidAssetType) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSetSuperfluidAssetEvent(denom, assetType),
	})
}

func newSetSuperfluidAssetEvent(denom string, assetType types.SuperfluidAssetType) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSetSuperfluidAsset,
		sdk.NewAttribute(types.AttributeDenom, denom),
		sdk.NewAttribute(types.AttributeSuperfluidAssetType, assetType.String()),
	)
}

func EmitRemoveSuperfluidAsset(ctx sdk.Context, denom string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newRemoveSuperfluidAssetEvent(denom),
	})
}

func newRemoveSuperfluidAssetEvent(denom string) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtRemoveSuperfluidAsset,
		sdk.NewAttribute(types.AttributeDenom, denom),
	)
}

func EmitSuperfluidDelegateEvent(ctx sdk.Context, lockId uint64, valAddress string, lockCoins sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSuperfluidDelegateEvent(lockId, valAddress, lockCoins),
	})
}

func newSuperfluidDelegateEvent(lockId uint64, valAddress string, lockCoins sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSuperfluidDelegate,
		sdk.NewAttribute(types.AttributeLockId, osmoutils.Uint64ToString(lockId)),
		sdk.NewAttribute(types.AttributeLockAmount, lockCoins[0].Amount.String()),
		sdk.NewAttribute(types.AttributeLockDenom, lockCoins[0].Denom),
		sdk.NewAttribute(types.AttributeValidator, valAddress),
	)
}

func EmitCreateFullRangePositionAndSuperfluidDelegateEvent(ctx sdk.Context, lockId, positionId uint64, valAddress string) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newCreateFullRangePositionAndSuperfluidDelegateEvent(lockId, positionId, valAddress),
	})
}

func newCreateFullRangePositionAndSuperfluidDelegateEvent(lockId, positionId uint64, valAddress string) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtCreateFullRangePositionAndSFDelegate,
		sdk.NewAttribute(types.AttributeLockId, osmoutils.Uint64ToString(lockId)),
		sdk.NewAttribute(types.AttributePositionId, osmoutils.Uint64ToString(positionId)),
		sdk.NewAttribute(types.AttributeValidator, valAddress),
	)
}

func EmitSuperfluidIncreaseDelegationEvent(ctx sdk.Context, lockId uint64, amount sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSuperfluidIncreaseDelegationEvent(lockId, amount),
	})
}

func newSuperfluidIncreaseDelegationEvent(lockId uint64, amount sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSuperfluidIncreaseDelegation,
		sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", lockId)),
		sdk.NewAttribute(types.AttributeAmount, amount.String()),
	)
}

func EmitSuperfluidUndelegateEvent(ctx sdk.Context, lockId uint64, lockCoins sdk.Coins) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSuperfluidUndelegateEvent(lockId, lockCoins),
	})
}

func newSuperfluidUndelegateEvent(lockId uint64, lockCoins sdk.Coins) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSuperfluidUndelegate,
		sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", lockId)),
		sdk.NewAttribute(types.AttributeLockAmount, lockCoins[0].Amount.String()),
		sdk.NewAttribute(types.AttributeLockDenom, lockCoins[0].Denom),
	)
}

func EmitSuperfluidUnbondLockEvent(ctx sdk.Context, lockId uint64) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSuperfluidUnbondLockEvent(lockId),
	})
}

func newSuperfluidUnbondLockEvent(lockId uint64) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSuperfluidUnbondLock,
		sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", lockId)),
	)
}

func EmitSuperfluidUndelegateAndUnbondLockEvent(ctx sdk.Context, lockId uint64) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newSuperfluidUndelegateAndUnbondLockEvent(lockId),
	})
}

func newSuperfluidUndelegateAndUnbondLockEvent(lockId uint64) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtSuperfluidUndelegateAndUnbondLock,
		sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", lockId)),
	)
}

func EmitUnpoolIdEvent(ctx sdk.Context, sender string, lpShareDenom string, allExitedLockIDsSerialized []byte) {
	if ctx.EventManager() == nil {
		return
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		newUnpoolIdEvent(sender, lpShareDenom, allExitedLockIDsSerialized),
	})
}

func newUnpoolIdEvent(sender string, lpShareDenom string, allExitedLockIDsSerialized []byte) sdk.Event {
	return sdk.NewEvent(
		types.TypeEvtUnpoolId,
		sdk.NewAttribute(sdk.AttributeKeySender, sender),
		sdk.NewAttribute(types.AttributeDenom, lpShareDenom),
		sdk.NewAttribute(types.AttributeNewLockIds, string(allExitedLockIDsSerialized)),
	)
}
