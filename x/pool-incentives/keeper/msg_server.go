package keeper

import (
	"context"
	"fmt"

	"github.com/osmosis-labs/osmosis/v25/x/pool-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// Gov messages

func (server msgServer) CreateGroups(goCtx context.Context, msg *types.MsgUpdateDistrRecords) (*types.MsgUpdateDistrRecordsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	err := server.keeper.UpdateDistrRecords(ctx, msg.Records...)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateDistrRecordsResponse{}, nil
}

func (server msgServer) ReplaceGroups(goCtx context.Context, msg *types.MsgReplaceDistrRecords) (*types.MsgReplaceDistrRecordsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	err := server.keeper.ReplaceDistrRecords(ctx, msg.Records...)
	if err != nil {
		return nil, err
	}

	return &types.MsgReplaceDistrRecordsResponse{}, nil
}

// Gov messages

func (server msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	server.keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
