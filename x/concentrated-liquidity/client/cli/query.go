package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPools)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdTickInfo)
	cmd.AddCommand(
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)
	return cmd
}

func GetCmdPool() (*osmocli.QueryDescriptor, *types.QueryPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool [poolID]",
		Short: "Query pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`}, &types.QueryPoolRequest{}
}

func GetCmdPools() (*osmocli.QueryDescriptor, *types.QueryPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`}, &types.QueryPoolsRequest{}
}

func GetCmdTickInfo() (*osmocli.QueryDescriptor, *types.QueryTickInfoRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "tick-info",
		Short: "Query a tick for the specified pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`}, &types.QueryTickInfoRequest{}
}
