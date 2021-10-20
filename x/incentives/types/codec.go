package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/incentives interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateGauge{}, "osmosis/incentives/create-gauge", nil)
	cdc.RegisterConcrete(&MsgAddToGauge{}, "osmosis/incentives/add-to-gauge", nil)
	cdc.RegisterConcrete(&MsgClaimLockReward{}, "osmosis/incentives/claim-lock-reward", nil)
	cdc.RegisterConcrete(&MsgClaimLockRewardAll{}, "osmosis/incentives/claim-lock-reward-all", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateGauge{},
		&MsgAddToGauge{},
		&MsgClaimLockReward{},
		&MsgClaimLockRewardAll{},
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
