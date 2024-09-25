package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&UpdateFeeTokenProposal{}, "osmosis/UpdateFeeTokenProposal", nil)
	legacy.RegisterAminoMsg(cdc, &MsgSetFeeTokens{}, "osmosis/txfees/set-fee-tokens")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypesv1.Content)(nil),
		&UpdateFeeTokenProposal{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSetFeeTokens{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
