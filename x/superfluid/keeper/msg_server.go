package keeper

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper/internal/events"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// SuperfluidDelegate creates a delegation for the given lock ID and the validator to delegate to.
// This requires the lock to have locked tokens that have been already registered as a superfluid asset via governance.
// The pre-requisites for a lock to be able to be eligible for superfluid delegation are
// - assets in the lock should be a superfluid registered asset
// - lock should only have a single asset
// - lock should not be unlocking
// - lock should not have a different superfluid staking position
// - lock duration should be greater or equal to the staking.Unbonding time
// Note that the amount of delegation is not equal to the equivalent amount of osmo within the lock.
// Instead, we use the osmo equivalent multiplier stored in the latest epoch, calculate how much
// osmo equivalent is in lock, and use the risk adjusted osmo value. The minimum risk ratio works as a parameter
// to better incentivize and balance between superfluid staking and vanilla staking.
// Delegation does not happen directly from msg.Sender, but instead delegation is done via intermediary account.
func (server msgServer) SuperfluidDelegate(goCtx context.Context, msg *types.MsgSuperfluidDelegate) (*types.MsgSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidDelegate(ctx, msg.Sender, msg.LockId, msg.ValAddr)
	if err == nil {
		events.EmitSuperfluidDelegateEvent(ctx, msg.LockId, msg.ValAddr)
	}
	return &types.MsgSuperfluidDelegateResponse{}, err
}

// SuperfluidUndelegate undelegates currently superfluid delegated position.
// Old synthetic lock is deleted and a new synthetic lock is created to indicate the unbonding position.
// The actual staking position is instantly undelegated and the undelegated tokens are instantly sent from
// the intermediary account to the module account.
// Note that SuperfluidUndelegation does not start unbonding of the underlying lock iteslf.
func (server msgServer) SuperfluidUndelegate(goCtx context.Context, msg *types.MsgSuperfluidUndelegate) (*types.MsgSuperfluidUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUndelegate(ctx, msg.Sender, msg.LockId)
	if err == nil {
		events.EmitSuperfluidUndelegateEvent(ctx, msg.LockId)
	}
	return &types.MsgSuperfluidUndelegateResponse{}, err
}

// SuperfluidRedelegate is a method to redelegate superfluid staked asset into a different validator.
// Currently this feature is not supported.
// func (server msgServer) SuperfluidRedelegate(goCtx context.Context, msg *types.MsgSuperfluidRedelegate) (*types.MsgSuperfluidRedelegateResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	err := server.keeper.SuperfluidRedelegate(ctx, msg.Sender, msg.LockId, msg.NewValAddr)
// 	return &types.MsgSuperfluidRedelegateResponse{}, err
// }

// SuperfluidUnbondLock starts unbonding for currently superfluid undelegating lock.
// This method would return an error when the underlying lock is not in an superfluid undelegating state,
// or if the lock is not used in superfluid staking.
func (server msgServer) SuperfluidUnbondLock(goCtx context.Context, msg *types.MsgSuperfluidUnbondLock) (
	*types.MsgSuperfluidUnbondLockResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUnbondLock(ctx, msg.LockId, msg.Sender)
	if err == nil {
		events.EmitSuperfluidUnbondLockEvent(ctx, msg.LockId)
	}
	return &types.MsgSuperfluidUnbondLockResponse{}, err
}

// SuperfluidUndelegateAndUnbondLock undelegates and unbonds partial amount from a lock.
func (server msgServer) SuperfluidUndelegateAndUnbondLock(goCtx context.Context, msg *types.MsgSuperfluidUndelegateAndUnbondLock) (
	*types.MsgSuperfluidUndelegateAndUnbondLockResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := server.keeper.SuperfluidUndelegateAndUnbondLock(ctx, msg.LockId, msg.Sender, msg.Coin.Amount)
	if err == nil {
		events.EmitSuperfluidUndelegateAndUnbondLockEvent(ctx, msg.LockId)
	}
	return &types.MsgSuperfluidUndelegateAndUnbondLockResponse{}, err
}

// LockAndSuperfluidDelegate locks and superfluid delegates given tokens in a single message.
// This method consists of multiple messages, `LockTokens` from the lockup module msg server, and
// `SuperfluidDelegate` from the superfluid module msg server.
func (server msgServer) LockAndSuperfluidDelegate(goCtx context.Context, msg *types.MsgLockAndSuperfluidDelegate) (*types.MsgLockAndSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lockupMsg := lockuptypes.MsgLockTokens{
		Owner:    msg.Sender,
		Duration: server.keeper.sk.GetParams(ctx).UnbondingTime,
		Coins:    msg.Coins,
	}

	lockupRes, err := server.keeper.lms.LockTokens(goCtx, &lockupMsg)
	if err != nil {
		return &types.MsgLockAndSuperfluidDelegateResponse{}, err
	}

	superfluidDelegateMsg := types.MsgSuperfluidDelegate{
		Sender:  msg.Sender,
		LockId:  lockupRes.GetID(),
		ValAddr: msg.ValAddr,
	}

	_, err = server.SuperfluidDelegate(goCtx, &superfluidDelegateMsg)
	return &types.MsgLockAndSuperfluidDelegateResponse{
		ID: lockupRes.ID,
	}, err
}

func (server msgServer) UnPoolWhitelistedPool(goCtx context.Context, msg *types.MsgUnPoolWhitelistedPool) (*types.MsgUnPoolWhitelistedPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = server.keeper.checkUnpoolWhitelisted(ctx, msg.PoolId)
	if err != nil {
		return nil, err
	}

	// We get all the lockIDs to unpool
	lpShareDenom := gammtypes.GetPoolShareDenom(msg.PoolId)
	minimalDuration := time.Millisecond
	unpoolLocks := server.keeper.lk.GetAccountLockedLongerDurationDenom(ctx, sender, lpShareDenom, minimalDuration)

	allExitedLockIDs := []uint64{}
	for _, lock := range unpoolLocks {
		exitedLockIDs, err := server.keeper.UnpoolAllowedPools(ctx, sender, msg.PoolId, lock.ID)
		if err != nil {
			return nil, err
		}
		allExitedLockIDs = append(allExitedLockIDs, exitedLockIDs...)
	}

	allExitedLockIDsSerialized, _ := json.Marshal(allExitedLockIDs)
	events.EmitUnpoolIdEvent(ctx, msg.Sender, lpShareDenom, allExitedLockIDsSerialized)

	return &types.MsgUnPoolWhitelistedPoolResponse{ExitedLockIds: allExitedLockIDs}, nil
}

func (server msgServer) UnlockAndMigrateSharesToFullRangeConcentratedPosition(goCtx context.Context, msg *types.MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition) (*types.MsgUnlockAndMigrateSharesToFullRangeConcentratedPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	amount0, amount1, liquidity, poolIdLeaving, poolIdEntering, newLockId, freezeDuration, err := server.keeper.UnlockAndMigrate(ctx, sender, msg.LockId, msg.SharesToMigrate)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUnlockAndMigrateShares,
			sdk.NewAttribute(types.AttributeKeyPoolIdEntering, strconv.FormatUint(poolIdEntering, 10)),
			sdk.NewAttribute(types.AttributeKeyPoolIdLeaving, strconv.FormatUint(poolIdLeaving, 10)),
			sdk.NewAttribute(types.AttributeNewLockId, strconv.FormatUint(newLockId, 10)),
			sdk.NewAttribute(types.AttributeKeyPoolIdLeaving, freezeDuration.String()),
		),
	})

	return &types.MsgUnlockAndMigrateSharesToFullRangeConcentratedPositionResponse{Amount0: amount0, Amount1: amount1, LiquidityCreated: liquidity}, err
}
