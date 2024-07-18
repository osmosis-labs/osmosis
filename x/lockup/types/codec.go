package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLockTokens{}, "symphony/lockup/lock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlockingAll{}, "symphony/lockup/begin-unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlocking{}, "symphony/lockup/begin-unlock-period-lock", nil)
	cdc.RegisterConcrete(&MsgExtendLockup{}, "symphony/lockup/extend-lockup", nil)
	cdc.RegisterConcrete(&MsgForceUnlock{}, "symphony/lockup/force-unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgSetRewardReceiverAddress{}, "symphony/lockup/set-reward-receiver-address", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLockTokens{},
		&MsgBeginUnlockingAll{},
		&MsgBeginUnlocking{},
		&MsgExtendLockup{},
		&MsgForceUnlock{},
		&MsgSetRewardReceiverAddress{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)
	RegisterCodec(authzcodec.Amino)

	amino.Seal()
}
