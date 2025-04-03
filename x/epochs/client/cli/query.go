package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdEpochInfos)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdCurrentEpoch)

	return cmd
}

func GetCmdEpochInfos() (*osmocli.QueryDescriptor, *types.QueryEpochsInfoRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "epoch-infos",
		Short: "Query running epoch infos.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}}`,
		QueryFnName: "EpochInfos",
	}, &types.QueryEpochsInfoRequest{}
}

func GetCmdCurrentEpoch() (*osmocli.QueryDescriptor, *types.QueryCurrentEpochRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "current-epoch",
		Short: "Query current epoch by specified identifier.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} day`,
	}, &types.QueryCurrentEpochRequest{}
}
