package streamswap

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/osmosis-labs/osmosis/v10/x/streamswap/types"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&types.MsgCreateSale{},
		&types.MsgExitSale{},
		&types.MsgFinalizeSale{},
		&types.MsgSubscribe{},
		&types.MsgWithdraw{},
	)

	// registry.RegisterInterface()

	msgservice.RegisterMsgServiceDesc(registry, types.MsgServiceDesc())
}
