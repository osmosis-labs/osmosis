package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
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

func (server msgServer) CreateDenom(goCtx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.Keeper.CreateDenom(ctx, msg.Sender, msg.Nonce)

	if err != nil {
		return nil, err
	}

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgCreateDenomResponse{}, nil
}

func (server msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	server.Keeper.mintTo(ctx, msg.Amount, msg.MintToAddress)

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgMintResponse{}, nil
}

func (server msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	server.Keeper.burnFrom(ctx, msg.Amount, msg.GetBurnFromAddress())

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgBurnResponse{}, nil
}

func (server msgServer) ForceTransfer(goCtx context.Context, msg *types.MsgForceTransfer) (*types.MsgForceTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Amount.GetDenom())
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	server.Keeper.forceTransfer(ctx, msg.Amount, msg.TransferFromAddress, msg.TransferToAddress)

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgForceTransferResponse{}, nil
}

func (server msgServer) ChangeAdmin(goCtx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityMetadata, err := server.Keeper.GetAuthorityMetadata(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	if msg.Sender != authorityMetadata.GetAdmin() {
		return nil, types.ErrUnauthorized
	}

	err = server.Keeper.setAdmin(ctx, msg.Denom, msg.NewAdmin)
	if err != nil {
		return nil, err
	}

	// TODO: events
	// ctx.EventManager().EmitEvents(sdk.Events{})

	return &types.MsgChangeAdminResponse{}, nil
}

var _ types.MsgServer = msgServer{}
