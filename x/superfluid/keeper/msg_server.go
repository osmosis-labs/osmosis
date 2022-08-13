package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v8constants "github.com/osmosis-labs/osmosis/v11/app/upgrades/v8/constants"
	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"

	"github.com/osmosis-labs/osmosis/v11/x/superfluid/types"
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

func (server msgServer) SuperfluidDelegate(goCtx context.Context, msg *types.MsgSuperfluidDelegate) (*types.MsgSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidDelegate(ctx, msg.Sender, msg.LockId, msg.ValAddr)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidDelegate,
			sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", msg.LockId)),
			sdk.NewAttribute(types.AttributeValidator, msg.ValAddr),
		))
	}
	return &types.MsgSuperfluidDelegateResponse{}, err
}

func (server msgServer) SuperfluidUndelegate(goCtx context.Context, msg *types.MsgSuperfluidUndelegate) (*types.MsgSuperfluidUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUndelegate(ctx, msg.Sender, msg.LockId)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidUndelegate,
			sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", msg.LockId)),
		))
	}
	return &types.MsgSuperfluidUndelegateResponse{}, err
}

// func (server msgServer) SuperfluidRedelegate(goCtx context.Context, msg *types.MsgSuperfluidRedelegate) (*types.MsgSuperfluidRedelegateResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	err := server.keeper.SuperfluidRedelegate(ctx, msg.Sender, msg.LockId, msg.NewValAddr)
// 	return &types.MsgSuperfluidRedelegateResponse{}, err
// }

func (server msgServer) SuperfluidUnbondLock(goCtx context.Context, msg *types.MsgSuperfluidUnbondLock) (
	*types.MsgSuperfluidUnbondLockResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUnbondLock(ctx, msg.LockId, msg.Sender)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidUnbondLock,
			sdk.NewAttribute(types.AttributeLockId, fmt.Sprintf("%d", msg.LockId)),
		))
	}
	return &types.MsgSuperfluidUnbondLockResponse{}, err
}

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

	if ctx.BlockHeight() < v8constants.UpgradeHeight {
		return nil, errors.New("message not activated")
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
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
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUnpoolId,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeDenom, lpShareDenom),
			sdk.NewAttribute(types.AttributeNewLockIds, string(allExitedLockIDsSerialized)),
		),
	})

	return &types.MsgUnPoolWhitelistedPoolResponse{ExitedLockIds: allExitedLockIDs}, nil
}
