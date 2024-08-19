package keeper

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/incentives/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// msgServer provides a way to reference keeper pointer in the message server interface.
type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer for the provided keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// CreateGauge creates a gauge and sends coins to the gauge.
// Emits create gauge event and returns the create gauge response.
func (server msgServer) CreateGauge(goCtx context.Context, msg *types.MsgCreateGauge) (*types.MsgCreateGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := server.keeper.chargeFeeIfSufficientFeeDenomBalance(ctx, owner, types.CreateGaugeFee, msg.Coins); err != nil {
		return nil, err
	}

	gaugeID, err := server.keeper.CreateGauge(ctx, msg.IsPerpetual, owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochsPaidOver, msg.PoolId)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateGauge,
			sdk.NewAttribute(types.AttributeGaugeID, osmoutils.Uint64ToString(gaugeID)),
		),
	})

	return &types.MsgCreateGaugeResponse{}, nil
}

// AddToGauge adds coins to gauge.
// Emits add to gauge event and returns the add to gauge response.
func (server msgServer) AddToGauge(goCtx context.Context, msg *types.MsgAddToGauge) (*types.MsgAddToGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := server.keeper.chargeFeeIfSufficientFeeDenomBalance(ctx, owner, types.AddToGaugeFee, msg.Rewards); err != nil {
		return nil, err
	}
	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, msg.GaugeId)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToGauge,
			sdk.NewAttribute(types.AttributeGaugeID, osmoutils.Uint64ToString(msg.GaugeId)),
		),
	})

	return &types.MsgAddToGaugeResponse{}, nil
}

func (server msgServer) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	groupID, err := server.keeper.CreateGroup(ctx, msg.Coins, msg.NumEpochsPaidOver, owner, msg.PoolIds)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateGroup,
			sdk.NewAttribute(types.AttributeGroupID, osmoutils.Uint64ToString(groupID)),
		),
	})

	return &types.MsgCreateGroupResponse{GroupId: groupID}, nil
}

// Gov messages

func (server msgServer) CreateGroups(goCtx context.Context, msg *types.MsgCreateGroups) (*types.MsgCreateGroupsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.ak.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	groupIds := make([]uint64, len(msg.CreateGroups))
	for i, group := range msg.CreateGroups {
		incentivesModuleAddress := server.keeper.ak.GetModuleAddress(types.ModuleName)
		// N.B: We force internal gauge creation here only because we don't have a straightforward
		// way to escrow the funds from the prop creator to be used at time of prop execution (or returned if the prop fails).
		// Once we have a way to do this, we can change the CreateGroups proto to allow for coins and numEpochsPaidOver and
		// then modify it here as well.
		// Note: do not replace with CreateGroupAsIncentivesModuleAcc as that implementation does not attempt to sync weights
		// We still want to sync the weights here to ensure that the pools are valid and have the associated volume at group creation time.
		groupId, err := server.keeper.CreateGroup(ctx, sdk.Coins{}, types.PerpetualNumEpochsPaidOver, incentivesModuleAddress, group.PoolIds)
		if err != nil {
			return nil, err
		}

		groupIds[i] = groupId
	}

	return &types.MsgCreateGroupsResponse{GroupId: groupIds}, nil
}

func (server msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.ak.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	server.keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
