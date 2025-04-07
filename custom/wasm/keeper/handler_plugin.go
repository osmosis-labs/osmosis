package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/osmosis-labs/osmosis/v27/ante"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// msgEncoder is an extension point to customize encodings
type msgEncoder interface {
	// Encode converts wasmvm message to n cosmos message types
	Encode(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Msg, error)
}

// MessageRouter ADR 031 request type routing
type MessageRouter interface {
	Handler(msg sdk.Msg) baseapp.MsgServiceHandler
}

// SDKMessageHandler can handles messages that can be encoded into sdk.Message types and routed.
type SDKMessageHandler struct {
	router         MessageRouter
	encoders       msgEncoder
	treasuryKeeper ante.TreasuryKeeper
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankKeeper.Keeper
}

var _ wasmkeeper.Messenger = SDKMessageHandler{}

func NewMessageHandler(
	router MessageRouter,
	ics4Wrapper wasmtypes.ICS4Wrapper,
	channelKeeper wasmtypes.ChannelKeeper,
	capabilityKeeper wasmtypes.CapabilityKeeper,
	wasmKeeper wasmtypes.IBCContractKeeper,
	bankKeeper bankKeeper.Keeper,
	treasuryKeeper ante.TreasuryKeeper,
	accountKeeper authkeeper.AccountKeeper,
	unpacker codectypes.AnyUnpacker,
	portSource wasmtypes.ICS20TransferPortSource,
	customEncoders ...*wasmkeeper.MessageEncoders,
) wasmkeeper.Messenger {
	encoders := wasmkeeper.DefaultEncoders(unpacker, portSource)
	for _, e := range customEncoders {
		encoders = encoders.Merge(e)
	}
	return wasmkeeper.NewMessageHandlerChain(
		NewSDKMessageHandler(router, encoders, treasuryKeeper, accountKeeper, bankKeeper),
		wasmkeeper.NewIBCRawPacketHandler(ics4Wrapper, wasmKeeper, channelKeeper, capabilityKeeper),
		wasmkeeper.NewBurnCoinMessageHandler(bankKeeper),
	)
}

func NewSDKMessageHandler(router MessageRouter, encoders msgEncoder, treasuryKeeper ante.TreasuryKeeper, accountKeeper authkeeper.AccountKeeper, bankKeeper bankKeeper.Keeper) SDKMessageHandler {
	return SDKMessageHandler{
		router:         router,
		encoders:       encoders,
		treasuryKeeper: treasuryKeeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

func (h SDKMessageHandler) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	contractIBCPortID string,
	msg wasmvmtypes.CosmosMsg,
) (
	events []sdk.Event,
	data [][]byte,
	msgResponses [][]*codectypes.Any,
	err error,
) {
	sdkMsgs, err := h.encoders.Encode(ctx, contractAddr, contractIBCPortID, msg)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, sdkMsg := range sdkMsgs {
		// Charge tax on result msg
		taxes := ante.FilterMsgAndComputeTax(ctx, h.treasuryKeeper, sdkMsg)
		if !taxes.IsZero() {
			eventManager := sdk.NewEventManager()
			contractAcc := h.accountKeeper.GetAccount(ctx, contractAddr)
			if err := cosmosante.DeductFees(h.bankKeeper, ctx.WithEventManager(eventManager), contractAcc, taxes); err != nil {
				return nil, nil, nil, err
			}

			events = eventManager.Events()
		}

		res, err := h.handleSdkMessage(ctx, contractAddr, sdkMsg)
		if err != nil {
			return nil, nil, nil, err
		}
		// append data
		data = append(data, res.Data)
		// append events
		sdkEvents := make([]sdk.Event, len(res.Events))
		for i := range res.Events {
			sdkEvents[i] = sdk.Event(res.Events[i])
		}
		events = append(events, sdkEvents...)
	}
	return events, data, nil, nil
}

func (h SDKMessageHandler) handleSdkMessage(ctx sdk.Context, contractAddr sdk.Address, msg sdk.Msg) (*sdk.Result, error) {
	if hasValidateBasic, ok := msg.(sdk.HasValidateBasic); ok {
		if err := hasValidateBasic.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	// find the handler and execute it
	if handler := h.router.Handler(msg); handler != nil {
		// ADR 031 request type routing
		msgResult, err := handler(ctx, msg)
		return msgResult, err
	}
	// legacy sdk.Msg routing
	// Assuming that the app developer has migrated all their Msgs to
	// proto messages and has registered all `Msg services`, then this
	// path should never be called, because all those Msgs should be
	// registered within the `msgServiceRouter` already.
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
}
