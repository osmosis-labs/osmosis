package model

import (
	"encoding/json"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
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
	codeId uint64,
	sender sdk.AccAddress,
	instantiateMsg []byte,
) MsgCreateCosmWasmPool {
	return MsgCreateCosmWasmPool{
		CodeId:         codeId,
		Sender:         sender.String(),
		InstantiateMsg: instantiateMsg,
	}
}

func (msg MsgCreateCosmWasmPool) Route() string { return types.RouterKey }
func (msg MsgCreateCosmWasmPool) Type() string  { return TypeMsgCreateCosmWasmPool }
func (msg MsgCreateCosmWasmPool) ValidateBasic() error {
	if msg.CodeId == 0 {
		return errors.New("CodeId cannot be 0")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return fmt.Errorf("Invalid sender address (%s)", err)
	}
	if !json.Valid(msg.InstantiateMsg) {
		return fmt.Errorf("InstantiateMsg is not a valid json")
	}
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
	poolI := NewCosmWasmPool(poolID, msg.CodeId, msg.InstantiateMsg)
	return poolI, nil
}

func (msg MsgCreateCosmWasmPool) GetPoolType() poolmanagertypes.PoolType {
	return poolmanagertypes.CosmWasm
}
