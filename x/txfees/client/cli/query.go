package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdFeeTokens(),
		osmocli.GetParams[*queryproto.ParamsRequest](
			types.ModuleName, queryproto.NewQueryClient),
	)

	return cmd
}

func GetCmdFeeTokens() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryFeeTokensRequest](
		"fee-tokens",
		"Query the list of non-basedenom fee tokens and their associated pool ids",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} fee-tokens
`,
		types.ModuleName, types.NewQueryClient,
	)
}
