package keeper

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		storeKey   sdk.StoreKey
		paramSpace paramtypes.Subspace

		channelKeeper  types.ChannelKeeper
		ContractKeeper *wasmkeeper.PermissionedKeeper
	}
)

// NewKeeper returns a new instance of the x/ibchooks keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	channelKeeper types.ChannelKeeper,
	contractKeeper *wasmkeeper.PermissionedKeeper,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		channelKeeper:  channelKeeper,
		ContractKeeper: contractKeeper,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of the module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the module's parameters with the provided parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

func GetPacketCallbackKey(channel string, packetSequence uint64) []byte {
	return []byte(fmt.Sprintf("%s::%d", channel, packetSequence))
}

func GetPacketAckKey(channel string, packetSequence uint64) []byte {
	return []byte(fmt.Sprintf("%s::%d::ack", channel, packetSequence))
}

// StorePacketCallback stores which contract will be listening for the ack or timeout of a packet
func (k Keeper) StorePacketCallback(ctx sdk.Context, channel string, packetSequence uint64, contract string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetPacketCallbackKey(channel, packetSequence), []byte(contract))
}

// GetPacketCallback returns the bech32 addr of the contract that is expecting a callback from a packet
func (k Keeper) GetPacketCallback(ctx sdk.Context, channel string, packetSequence uint64) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(GetPacketCallbackKey(channel, packetSequence)))
}

// IsInAllowList checks the params to see if the contract is in the KeyAsyncAckAllowList param
func (k Keeper) IsInAllowList(ctx sdk.Context, contract string) bool {
	var allowList []string
	k.paramSpace.GetIfExists(ctx, types.KeyAsyncAckAllowList, &allowList)
	for _, addr := range allowList {
		if addr == contract {
			return true
		}
	}
	return false
}

// DeletePacketCallback deletes the callback from storage once it has been processed
func (k Keeper) DeletePacketCallback(ctx sdk.Context, channel string, packetSequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetPacketCallbackKey(channel, packetSequence))
}

// StorePacketAckActor stores which contract is allowed to send an ack for the packet
func (k Keeper) StorePacketAckActor(ctx sdk.Context, channel string, packetSequence uint64, contract string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetPacketAckKey(channel, packetSequence), []byte(contract))
}

// GetPacketAckActor returns the bech32 addr of the contract that is allowed to send an ack for the packet
func (k Keeper) GetPacketAckActor(ctx sdk.Context, channel string, packetSequence uint64) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(GetPacketAckKey(channel, packetSequence)))
}

// DeletePacketAckActor deletes the ack actor from storage once it has been used
func (k Keeper) DeletePacketAckActor(ctx sdk.Context, channel string, packetSequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetPacketAckKey(channel, packetSequence))
}

// DeriveIntermediateSender derives the sender address to be used when calling wasm hooks
func DeriveIntermediateSender(channel, originalSender, bech32Prefix string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderHash32 := address.Hash(types.SenderPrefix, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:])
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}

// EmitIBCAck emits an event that the IBC packet has been acknowledged
func (k Keeper) EmitIBCAck(ctx sdk.Context, sender, channel string, packetSequence uint64) ([]byte, error) {
	contract := k.GetPacketAckActor(ctx, channel, packetSequence)
	if contract == "" {
		return nil, fmt.Errorf("no ack actor set for channel %s packet %d", channel, packetSequence)
	}
	// Only the contract itself can request for the ack to be emitted. This will generally happen as a callback
	// when the result of other IBC actions has finished, but it could be exposed directly by the contract if the
	// proper checks are made
	if sender != contract {
		return nil, fmt.Errorf("sender %s is not allowed to send an ack for channel %s packet %d", sender, channel, packetSequence)
	}

	// Write the acknowledgement
	_, cap, err := k.channelKeeper.LookupModuleByChannel(ctx, "transfer", channel)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Calling the contract. This could be made generic by using an interface if we want
	// to support other types of AckActors, but keeping it here for now for simplicity.
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse contract address")
	}

	msg := types.IBCAsync{
		RequestAck: types.RequestAck{RequestAckI: types.RequestAckI{
			PacketSequence: packetSequence,
			SourceChannel:  channel,
		}},
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not marshal message")
	}
	bz, err := k.ContractKeeper.Sudo(ctx, contractAddr, msgBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not execute contract")
	}

	ack, err := types.UnmarshalIBCAck(bz)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not unmarshal into IBCAckResponse or IBCAckError")

	}
	var newAck channeltypes.Acknowledgement
	var packet exported.PacketI

	switch ack.Type {
	case "ack_response":
		jsonAck, err := json.Marshal(ack.AckResponse)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "could not marshal acknowledgement")
		}
		packet = ack.AckResponse.Packet
		newAck = channeltypes.NewResultAcknowledgement(jsonAck)
	case "ack_error":
		packet = ack.AckError.Packet
		newAck = osmoutils.NewEmitErrorAcknowledgement(ctx, types.ErrAckFromContract, ack.AckError.ContractError)
	default:
		return nil, sdkerrors.Wrap(err, "could not unmarshal into IBCAckResponse or IBCAckError")
	}

	err = k.channelKeeper.WriteAcknowledgement(ctx, cap, packet, newAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not write acknowledgement")
	}

	response, err := json.Marshal(newAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not marshal acknowledgement")
	}
	return response, nil
}
