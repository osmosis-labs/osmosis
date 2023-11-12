package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type (
	Keeper struct {
		storeKey   storetypes.StoreKey
		paramSpace paramtypes.Subspace

		channelKeeper  types.ChannelKeeper
		ContractKeeper *wasmkeeper.PermissionedKeeper
	}
)

// NewKeeper returns a new instance of the x/ibchooks keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
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

// SetParam sets a specific ibc-hooks module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
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

func GeneratePacketAckValue(packet channeltypes.Packet, contract string) ([]byte, error) {
	if _, err := sdk.AccAddressFromBech32(contract); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidContractAddr, contract)
	}

	packetHash, err := hashPacket(packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not hash packet")
	}

	return []byte(fmt.Sprintf("%s::%s", contract, packetHash)), nil
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
func (k Keeper) StorePacketAckActor(ctx sdk.Context, packet channeltypes.Packet, contract string) {
	store := ctx.KVStore(k.storeKey)
	channel := packet.GetSourceChannel()
	packetSequence := packet.GetSequence()

	val, err := GeneratePacketAckValue(packet, contract)
	if err != nil {
		panic(err)
	}
	store.Set(GetPacketAckKey(channel, packetSequence), val)
}

// GetPacketAckActor returns the bech32 addr  of the contract that is allowed to send an ack for the packet and the packet hash
func (k Keeper) GetPacketAckActor(ctx sdk.Context, channel string, packetSequence uint64) (string, string) {
	store := ctx.KVStore(k.storeKey)
	rawData := store.Get(GetPacketAckKey(channel, packetSequence))
	if rawData == nil {
		return "", ""
	}
	data := strings.Split(string(rawData), "::")
	if len(data) != 2 {
		return "", ""
	}
	// validate that the contract is a valid bech32 addr
	if _, err := sdk.AccAddressFromBech32(data[0]); err != nil {
		return "", ""
	}
	// validate that the hash is a valid sha256sum hash
	if _, err := hex.DecodeString(data[1]); err != nil {
		return "", ""
	}

	return data[0], data[1]
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
	contract, packetHash := k.GetPacketAckActor(ctx, channel, packetSequence)
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
	var packet channeltypes.Packet

	switch ack.Type {
	case "ack_response":
		jsonAck, err := json.Marshal(ack.AckResponse.ContractAck)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "could not marshal acknowledgement")
		}
		packet = ack.AckResponse.Packet
		newAck = channeltypes.NewResultAcknowledgement(jsonAck)
	case "ack_error":
		packet = ack.AckError.Packet
		newAck = osmoutils.NewSuccessAckRepresentingAnError(ctx, types.ErrAckFromContract, []byte(ack.AckError.ErrorResponse), ack.AckError.ErrorDescription)
	default:
		return nil, sdkerrors.Wrap(err, "could not unmarshal into IBCAckResponse or IBCAckError")
	}

	// Validate that the packet returned by the contract matches the one we stored when sending
	receivedPacketHash, err := hashPacket(packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not hash packet")
	}
	if receivedPacketHash != packetHash {
		return nil, sdkerrors.Wrap(types.ErrAckPacketMismatch, fmt.Sprintf("packet hash mismatch. Expected %s, got %s", packetHash, receivedPacketHash))
	}

	// Now we can write the acknowledgement
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

func hashPacket(packet channeltypes.Packet) (string, error) {
	// ignore the data here. We only care about the channel information
	packet.Data = nil
	bz, err := json.Marshal(packet)
	if err != nil {
		return "", sdkerrors.Wrap(err, "could not marshal packet")
	}
	packetHash := tmhash.Sum(bz)
	return hex.EncodeToString(packetHash), nil
}
