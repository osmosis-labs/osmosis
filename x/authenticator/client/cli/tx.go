package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewAddAuthentiactorCmd)
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
	}, &types.MsgAddAuthenticator{}
}
