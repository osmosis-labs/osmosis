package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

// Bech32HrpToSourceChannelMap defines the contract that must be fulfilled by a bech32 prefix to source
// channel mapper
// The x/bech32ibc keeper is a reference implementation and is expected to satisfy this interface
type Bech32HrpToSourceChannelMap interface {
	GetHrpSourceChannel(ctx sdk.Context, hrp string) (sourceChannel string, err error)
	GetNativeHrp(ctx sdk.Context) (hrp string, err error)
}

// ICS20TransferMsgServer defines the contract that must be fulfilled by an ICS20 msg server
type ICS20TransferMsgServer interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}

type TransferKeeper interface {
	// GetPort returns the portID for the transfer module. Used in ExportGenesis
	GetPort(ctx sdk.Context) string
}
