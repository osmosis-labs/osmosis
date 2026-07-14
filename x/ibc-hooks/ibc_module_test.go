package ibc_hooks_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	ibchooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"
)

type mockApp struct {
	porttypes.IBCModule

	negotiatedVersion string
}

func (m mockApp) OnChanOpenInit(
	_ sdk.Context,
	_ channeltypes.Order,
	_ []string,
	_, _ string,
	_ *capabilitytypes.Capability,
	_ channeltypes.Counterparty,
	_ string,
) (string, error) {
	return m.negotiatedVersion, nil
}

func TestOnChanOpenInitReturnsNegotiatedVersion(t *testing.T) {
	const negotiated = "ics20-1"
	ics4 := ibchooks.NewICS4Middleware(nil, nil)
	middleware := ibchooks.NewIBCMiddleware(mockApp{negotiatedVersion: negotiated}, &ics4)
	version, err := middleware.OnChanOpenInit(
		sdk.Context{},
		channeltypes.UNORDERED,
		[]string{"connection-0"},
		"transfer",
		"channel-0",
		nil,
		channeltypes.Counterparty{},
		"",
	)
	require.NoError(t, err)
	require.Equal(t, negotiated, version)
}
