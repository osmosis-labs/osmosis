package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPools)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetUserPositions)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetClaimableFees)
	cmd.AddCommand(
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)
	return cmd
}

func GetUserPositions() (*osmocli.QueryDescriptor, *types.QueryUserPositionsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "user-positions [address]",
			Short: "Query user's positions",
			Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} user-positions osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`,
			CustomFlagOverrides: poolIdFlagOverride},
		&types.QueryUserPositionsRequest{}
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

func GetClaimableFees() (*osmocli.QueryDescriptor, *types.QueryClaimableFeesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "claimable-fees [poolID] [address] [lowerTick] [upperTick]",
		Short: "Query claimable fees",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} claimable-fees 1 osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj [-100] 100`}, &types.QueryClaimableFeesRequest{}
}
