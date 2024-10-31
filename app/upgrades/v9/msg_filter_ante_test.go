package v9_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/osmosis-labs/osmosis/v27/app"
	v8 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v8"
	v9 "github.com/osmosis-labs/osmosis/v27/app/upgrades/v9"
)

func noOpAnteDecorator() sdk.AnteHandler {
	return func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
}

func TestMsgFilterDecorator(t *testing.T) {
	handler := v9.MsgFilterDecorator{}
	txCfg := app.MakeEncodingConfig().TxConfig

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	testCases := []struct {
		name      string
		ctx       sdk.Context
		msgs      []sdk.Msg
		expectErr bool
	}{
		{
			name: "valid tx",
			ctx:  sdk.Context{}.WithBlockHeight(v8.UpgradeHeight - 1),
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("foo", 5))),
			},
			expectErr: false,
		},
		{
			name: "invalid tx",
			ctx:  sdk.Context{}.WithBlockHeight(v8.UpgradeHeight),
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewInt64Coin("foo", 5))),
				&ibcchanneltypes.MsgTimeoutOnClose{},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			txBuilder := txCfg.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msgs...))

			_, err := handler.AnteHandle(tc.ctx, txBuilder.GetTx(), false, noOpAnteDecorator())
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
