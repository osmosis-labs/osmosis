// Package keeper TODO: upgrade the signatures validation process
package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (m msgServer) InboundTransfer(
	goCtx context.Context,
	msg *types.MsgInboundTransfer,
) (*types.MsgInboundTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.k.validateSenderIsSigner(ctx, msg.Sender) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "Sender is not part of the signer set")
	}

	err := m.k.InboundTransfer(ctx, msg.DestAddr, msg.Asset, msg.Amount)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&types.EventInboundTransfer{
		Sender:   msg.Sender,
		DestAddr: msg.DestAddr,
		Asset:    msg.Asset,
		Amount:   msg.Amount,
	})
	if err != nil {
		return nil, err
	}

	return new(types.MsgInboundTransferResponse), nil
}

func (m msgServer) OutboundTransfer(
	goCtx context.Context,
	msg *types.MsgOutboundTransfer,
) (*types.MsgOutboundTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Don't need to check the signature here since every user could be the sender

	err := m.k.OutboundTransfer(ctx, msg.Sender, msg.Asset, msg.Amount)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&types.EventOutboundTransfer{
		Sender:   msg.Sender,
		DestAddr: msg.DestAddr,
		Asset:    msg.Asset,
		Amount:   msg.Amount,
	})
	if err != nil {
		return nil, err
	}

	// TODO: How to pass the outbound tx to the TSS valset?

	return new(types.MsgOutboundTransferResponse), nil
}

func (m msgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Sender != m.k.govModuleAddr {
		return nil, errorsmod.Wrapf(sdkerrors.ErrorInvalidSigner, "Only the gov module can update params")
	}

	result, err := m.k.UpdateParams(ctx, msg.NewParams)
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventUpdateParams{
		NewSigners:     msg.NewParams.Signers,
		CreatedSigners: result.signersToCreate,
		DeletedSigners: result.signersToDelete,
		NewAssets:      msg.NewParams.Assets,
		CreatedAssets:  result.assetsToCreate,
		DeletedAssets:  result.assetsToDelete,
	})
	if err != nil {
		return nil, err
	}

	return new(types.MsgUpdateParamsResponse), nil
}

func (m msgServer) ChangeAssetStatus(
	goCtx context.Context,
	msg *types.MsgChangeAssetStatus,
) (*types.MsgChangeAssetStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	result, err := m.k.ChangeAssetStatus(ctx, msg.Asset, msg.NewAssetStatus)
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventChangeAssetStatus{
		Sender:         msg.Sender,
		Asset:          msg.Asset,
		OldAssetStatus: result.OldStatus,
		NewAssetStatus: result.NewStatus,
	})
	if err != nil {
		return nil, err
	}

	return new(types.MsgChangeAssetStatusResponse), nil
}
