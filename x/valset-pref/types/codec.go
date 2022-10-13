package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetValidatorSetPreference{}, "osmosis/validator-set-preference/set-validator-set-preference", nil)
	cdc.RegisterConcrete(&MsgDelegateToValidatorSet{}, "osmosis/validator-set-preference/delegate-to-validator-set", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromValidatorSet{}, "osmosis/validator-set-preference/undelegate-from-validator-set", nil)
	cdc.RegisterConcrete(&MsgWithdrawDelegationRewards{}, "osmosis/validator-set-preference/withdraw-delegation-rewards", nil)
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
