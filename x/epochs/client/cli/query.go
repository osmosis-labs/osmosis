package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		GetCmdEpochsInfos(),
		GetCmdCurrentEpoch(),
	)

	return cmd
}

func GetCmdEpochsInfos() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryEpochsInfoRequest](
		"epoch-infos",
		"Query running epochInfos",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} epoch-infos
`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdCurrentEpoch() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryCurrentEpochRequest](
		"current-epoch [identifier]",
		"Query current epoch by specified identifier",
		`{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} current-epoch day
`,
		types.ModuleName, types.NewQueryClient,
	)
}
