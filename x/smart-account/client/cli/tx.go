package cli

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewAddAuthentiactorCmd)
	osmocli.AddTxCmd(txCmd, NewRemoveAuthentiactorCmd)
	return txCmd
}

func NewAddAuthentiactorCmd() (*osmocli.TxCliDesc, *types.MsgAddAuthenticator) {
	return &osmocli.TxCliDesc{
		Use:   "add-authenticator",
		Short: "add an authenticator for an address",
		Long:  "",
		Example: `
			osmosisd tx authenticator add-authenticator SigVerification <pubkey> --from val \
			--chain-id osmosis-1 -b sync --keyring-backend test \
			--fees 1000uosmo
		`,
		ParseAndBuildMsg: BuildAddAuthenticatorMsg,
	}, &types.MsgAddAuthenticator{}
}

func NewRemoveAuthentiactorCmd() (*osmocli.TxCliDesc, *types.MsgRemoveAuthenticator) {
	return &osmocli.TxCliDesc{
		Use:   "remove-authenticator",
		Short: "add an authenticator for an address",
		Long:  "",
		Example: `
			osmosisd tx authenticator remove-authenticator 1 --from val \
			--chain-id osmosis-1 -b sync --keyring-backend test \
			--fees 1000uosmo
		`,
	}, &types.MsgRemoveAuthenticator{}
}

func BuildAddAuthenticatorMsg(
	clientCtx client.Context,
	args []string,
	flags *pflag.FlagSet,
) (sdk.Msg, error) {
	authenticatorType := args[0]
	pubKeyEncoded := args[1]

	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyEncoded)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddAuthenticator{
		AuthenticatorType: authenticatorType,
		Data:              pubKeyBytes,
		Sender:            clientCtx.GetFromAddress().String(),
	}, nil
}
