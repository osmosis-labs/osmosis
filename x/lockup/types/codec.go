package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLockTokens{}, "osmosis/lockup/lock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlockingAll{}, "osmosis/lockup/begin-unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgUnlockTokens{}, "osmosis/lockup/unlock-tokens", nil)
	cdc.RegisterConcrete(&MsgBeginUnlocking{}, "osmosis/lockup/begin-unlock-period-lock", nil)
	cdc.RegisterConcrete(&MsgUnlockPeriodLock{}, "osmosis/lockup/unlock-period-lock", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLockTokens{},
		&MsgBeginUnlockingAll{},
		&MsgUnlockTokens{},
		&MsgBeginUnlocking{},
		&MsgUnlockPeriodLock{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
