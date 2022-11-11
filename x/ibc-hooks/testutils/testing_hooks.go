package testutils

import (
	// external libraries
	sdk "github.com/cosmos/cosmos-sdk/types"

	// ibc-go
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	ibchooks "github.com/osmosis-labs/osmosis/v12/x/ibc-hooks"
)

var (
	_ ibchooks.Hooks = TestRecvOverrideHooks{}
	_ ibchooks.Hooks = TestRecvBeforeAfterHooks{}
)

type Status struct {
	OverrideRan bool
	BeforeRan   bool
	AfterRan    bool
}

// Recv
type TestRecvOverrideHooks struct{ Status *Status }

func (t TestRecvOverrideHooks) OnRecvPacketOverride(im ibchooks.IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	t.Status.OverrideRan = true
	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	return ack
}

type TestRecvBeforeAfterHooks struct{ Status *Status }

func (t TestRecvBeforeAfterHooks) OnRecvPacketBeforeHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) {
	t.Status.BeforeRan = true
}

func (t TestRecvBeforeAfterHooks) OnRecvPacketAfterHook(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, ack ibcexported.Acknowledgement) {
	t.Status.AfterRan = true
}
