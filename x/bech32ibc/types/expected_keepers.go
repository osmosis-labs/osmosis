package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (_ channeltypes.Channel, found bool)
}

type TransferKeeper interface {
	// GetPort returns the portID for the transfer module. Used in ExportGenesis
	GetPort(ctx sdk.Context) string
}
