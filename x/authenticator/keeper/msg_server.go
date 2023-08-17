package keeper

import (
	"context"
	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) MsgAddAuthenticator(goCtx context.Context, request *types.MsgAddAuthenticatorRequest) (*types.MsgAddAuthenticatorResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	//TODO implement me
	panic("implement me")
}

func (m msgServer) MsgRemoveAuthenticator(goCtx context.Context, request *types.MsgRemoveAuthenticatorRequest) (*types.MsgRemoveAuthenticatorResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)

	//TODO implement me
	panic("implement me")
}
