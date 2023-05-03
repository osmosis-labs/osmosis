package model

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v15/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// constants.
const (
	TypeMsgCreateCosmWasmPool = "create_cosmwasm_pool"
)

var (
	_ sdk.Msg                        = &MsgCreateCosmWasmPool{}
	_ poolmanagertypes.CreatePoolMsg = &MsgCreateCosmWasmPool{}
)

func NewMsgCreateCosmWasmPool(
	sender sdk.AccAddress,
) MsgCreateCosmWasmPool {
	return MsgCreateCosmWasmPool{
		Sender: sender.String(),
	}
}

func (msg MsgCreateCosmWasmPool) Route() string { return types.RouterKey }
func (msg MsgCreateCosmWasmPool) Type() string  { return TypeMsgCreateCosmWasmPool }
func (msg MsgCreateCosmWasmPool) ValidateBasic() error {
	// TODO: add more validation.

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

func (msg MsgCreateCosmWasmPool) GetSignBytes() []byte {
	// TODO: uncomment once merging state-breaks.
	// return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
	return nil
}

func (msg MsgCreateCosmWasmPool) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// Implement the CreatePoolMsg interface
func (msg MsgCreateCosmWasmPool) PoolCreator() sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return sender
}

func (msg MsgCreateCosmWasmPool) Validate(ctx sdk.Context) error {
	return msg.ValidateBasic()
}

func (msg MsgCreateCosmWasmPool) InitialLiquidity() sdk.Coins {
	return sdk.Coins{}
}

func (msg MsgCreateCosmWasmPool) CreatePool(ctx sdk.Context, poolID uint64) (poolmanagertypes.PoolI, error) {
	// TODO: uncomment once merging state-breaks.
	// poolI := NewCosmWasmPool(poolID, msg.CodeId, msg.InstantiateMsg)
	// return &poolI, nil
	return nil, nil
}

func (msg MsgCreateCosmWasmPool) GetPoolType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}
