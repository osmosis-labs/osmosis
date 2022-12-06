package osmocli

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type TxCliTestCase[M sdk.Msg] struct {
	Cmd         string
	ExpectedMsg M
	ExpectedErr bool
}

func RunTxTestCases[M sdk.Msg](t *testing.T, desc *TxCliDesc, testcases map[string]TxCliTestCase[M]) {
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			RunTxTestCase(t, desc, &tc)
		})
	}
}

func RunTxTestCase[M sdk.Msg](t *testing.T, desc *TxCliDesc, tc *TxCliTestCase[M]) {
	cmd := BuildTxCli[M](desc)

	args := strings.Split(tc.Cmd, " ")
	pflag.CommandLine.AddFlagSet(cmd.Flags())
	err := pflag.CommandLine.Parse(args)
	require.NoError(t, err, "error in pflag.CommandLine.Parse(args)")
	clientCtx := newClientContextWithFrom(t, cmd.Flags())

	msg, err := desc.ParseAndBuildMsg(clientCtx, args, cmd.Flags())
	if tc.ExpectedErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err, "error in desc.ParseAndBuildMsg")
	require.Equal(t, tc.ExpectedMsg, msg)
}

// This logic is copied from the SDK, it should've just been publicly exposed.
// But instead its buried within a mega-method.
func newClientContextWithFrom(t *testing.T, fs *pflag.FlagSet) client.Context {
	clientCtx := client.Context{}
	from, _ := fs.GetString(flags.FlagFrom)
	fromAddr, fromName, _, err := client.GetFromFields(nil, from, true)
	require.NoError(t, err)

	clientCtx = clientCtx.WithFrom(from).WithFromAddress(fromAddr).WithFromName(fromName)
	return clientCtx
}
