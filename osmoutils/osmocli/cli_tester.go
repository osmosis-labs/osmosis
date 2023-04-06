package osmocli

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type TxCliTestCase[M sdk.Msg] struct {
	Cmd                    string
	ExpectedMsg            M
	ExpectedErr            bool
	OnlyCheckValidateBasic bool
}

type QueryCliTestCase[Q proto.Message] struct {
	Cmd           string
	ExpectedQuery Q
	ExpectedErr   bool
}

func RunTxTestCases[M sdk.Msg](t *testing.T, desc *TxCliDesc, testcases map[string]TxCliTestCase[M]) {
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			RunTxTestCase(t, desc, &tc)
		})
	}
}

func RunQueryTestCases[Q proto.Message](t *testing.T, desc *QueryDescriptor, testcases map[string]QueryCliTestCase[Q]) {
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			RunQueryTestCase(t, desc, &tc)
		})
	}
}

func RunTxTestCase[M sdk.Msg](t *testing.T, desc *TxCliDesc, tc *TxCliTestCase[M]) {
	cmd := BuildTxCli[M](desc)
	err := resetCommandFlagValues(cmd)
	require.NoError(t, err, "error in resetCommandFlagValues")
	args := strings.Split(tc.Cmd, " ")
	err = cmd.Flags().Parse(args)
	require.NoError(t, err, "error in cmd.Flags().Parse(args)")
	clientCtx := newClientContextWithFrom(t, cmd.Flags())

	msg, err := desc.ParseAndBuildMsg(clientCtx, args, cmd.Flags())
	if tc.ExpectedErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err, "error in desc.ParseAndBuildMsg")
	if tc.OnlyCheckValidateBasic {
		require.NoError(t, msg.ValidateBasic())
		return
	}

	require.Equal(t, tc.ExpectedMsg, msg)
}

func RunQueryTestCase[Q proto.Message](t *testing.T, desc *QueryDescriptor, tc *QueryCliTestCase[Q]) {
	cmd := BuildQueryCli[Q, int](desc, nil)
	err := resetCommandFlagValues(cmd)
	require.NoError(t, err, "error in resetCommandFlagValues")
	args := strings.Split(tc.Cmd, " ")
	err = cmd.Flags().Parse(args)
	require.NoError(t, err, "error in cmd.Flags().Parse(args)")

	req, err := desc.ParseQuery(args, cmd.Flags())
	if tc.ExpectedErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err, "error in desc.ParseQuery")
	require.Equal(t, tc.ExpectedQuery, req)
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

// taken from https://github.com/golang/debug/pull/8,
// due to no cobra command for resetting flag value
func resetCommandFlagValues(cmd *cobra.Command) error {
	var retErr error = nil
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			err := f.Value.Set(f.DefValue)
			if err != nil {
				retErr = err
			}
			f.Changed = false
		}
	})
	return retErr
}
