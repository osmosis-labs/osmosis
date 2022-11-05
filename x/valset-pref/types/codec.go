package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetValidatorSetPreference{}, "osmosis/valset-pref/MsgSetValidatorSetPreference", nil)
	cdc.RegisterConcrete(&MsgDelegateToValidatorSet{}, "osmosis/valset-pref/MsgDelegateToValidatorSet", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromValidatorSet{}, "osmosis/valset-pref/MsgUndelegateFromValidatorSet", nil)
	cdc.RegisterConcrete(&MsgWithdrawDelegationRewards{}, "osmosis/valset-pref/MsgWithdrawDelegationRewards", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetValidatorSetPreference{},
		&MsgDelegateToValidatorSet{},
		&MsgUndelegateFromValidatorSet{},
		&MsgWithdrawDelegationRewards{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
