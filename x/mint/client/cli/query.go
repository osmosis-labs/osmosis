package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/mint/types"
)

// GetQueryCmd returns the cli query commands for the minting module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdQueryParams)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdQueryEpochProvisions)

	return cmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams() (*osmocli.QueryDescriptor, *types.QueryParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query the current minting parameters",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} params`,
	}, &types.QueryParamsRequest{}
}

// GetCmdQueryEpochProvisions implements a command to return the current minting
// epoch provisions value.
func GetCmdQueryEpochProvisions() (*osmocli.QueryDescriptor, *types.QueryEpochProvisionsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "epoch-provisions",
		Short: "Query the current minting epoch provisions value",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} epoch-provisions`,
	}, &types.QueryEpochProvisionsRequest{}
}
