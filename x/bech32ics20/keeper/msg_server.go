package keeper

import (
	"context"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) Send(goCtx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.bk.SendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	prefix, _, err := bech32.DecodeAndConvert(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	nativePrefix, err := k.hrpToChannelMapper.GetNativeHrp(ctx)
	if err != nil {
		return nil, err
	}

	if prefix == nativePrefix {

		to, err := sdk.AccAddressFromBech32(msg.ToAddress)
		if err != nil {
			return nil, err
		}

		if k.bk.BlockedAddr(to) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.ToAddress)
		}

		err = k.bk.SendCoins(ctx, from, to, msg.Amount)
		if err != nil {
			return nil, err
		}

		defer func() {
			for _, a := range msg.Amount {
				if a.Amount.IsInt64() {
					telemetry.SetGaugeWithLabels(
						[]string{"tx", "msg", "send"},
						float32(a.Amount.Int64()),
						[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
					)
				}
			}
		}()

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			),
		)

		return &types.MsgSendResponse{}, nil
	}

	sourceChannel, err := k.hrpToChannelMapper.GetHrpSourceChannel(ctx, prefix)
	if err != nil {
		return nil, err
	}

	if msg.Amount.Len() == 0 {
		return nil, sdkerrors.Wrap(types.ErrNoInputs, "invalid send amount")
	}
	if msg.Amount.Len() > 1 {
		return nil, sdkerrors.Wrap(ibctransfertypes.ErrInvalidAmount, "cannot send multiple denoms via IBC")
	}

	ibcTransferMsg := ibctransfertypes.NewMsgTransfer(
		k.tk.GetPort(ctx),
		sourceChannel,
		msg.Amount[0],
		from,
		msg.ToAddress,
		clienttypes.ZeroHeight(), 0, // Use no timeouts for now.  Can add this in future.
	)

	_, err = k.ics20TransferMsgServer.Transfer(sdk.WrapSDKContext(ctx), ibcTransferMsg)

	return &types.MsgSendResponse{}, err
}

func (k msgServer) MultiSend(goCtx context.Context, msg *types.MsgMultiSend) (*types.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// NOTE: totalIn == totalOut should already have been checked
	for _, in := range msg.Inputs {
		if err := k.bk.SendEnabledCoins(ctx, in.Coins...); err != nil {
			return nil, err
		}
	}

	for _, out := range msg.Outputs {
		accAddr, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			panic(err)
		}
		if k.bk.BlockedAddr(accAddr) {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive transactions", out.Address)
		}
	}

	err := k.bk.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &types.MsgMultiSendResponse{}, nil
}
