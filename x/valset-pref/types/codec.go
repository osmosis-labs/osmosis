package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetValidatorSetPreference{}, "symphony/MsgSetValidatorSetPreference", nil)
	cdc.RegisterConcrete(&MsgDelegateToValidatorSet{}, "symphony/MsgDelegateToValidatorSet", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromValidatorSet{}, "symphony/MsgUndelegateFromValidatorSet", nil)
	cdc.RegisterConcrete(&MsgUndelegateFromRebalancedValidatorSet{}, "symphony/MsgUndelegateFromRebalValset", nil)
	cdc.RegisterConcrete(&MsgWithdrawDelegationRewards{}, "symphony/MsgWithdrawDelegationRewards", nil)
	cdc.RegisterConcrete(&MsgRedelegateValidatorSet{}, "symphony/MsgRedelegateValidatorSet", nil)
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

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterCodec(amino)

	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterCodec(authzcodec.Amino)

	amino.Seal()
}
