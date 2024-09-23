package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetValidatorSetPreference{}, "osmosis/MsgSetValidatorSetPreference", nil)
	cdc.RegisterConcrete(&MsgDelegateToValidatorSet{}, "osmosis/MsgDelegateToValidatorSet", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromValidatorSet{}, "osmosis/MsgUndelegateFromValidatorSet", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromRebalancedValidatorSet{}, "osmosis/MsgUndelegateFromRebalValset", nil)
	cdc.RegisterConcrete(&MsgWithdrawDelegationRewards{}, "osmosis/MsgWithdrawDelegationRewards", nil)
	cdc.RegisterConcrete(&MsgRedelegateValidatorSet{}, "osmosis/MsgRedelegateValidatorSet", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSetValidatorSetPreference{},
		&MsgDelegateToValidatorSet{},
		&MsgUndelegateFromValidatorSet{},
		&MsgUndelegateFromRebalancedValidatorSet{},
		&MsgWithdrawDelegationRewards{},
		&MsgRedelegateValidatorSet{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
