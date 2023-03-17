package cli

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/query"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, query.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, query.NewQueryClient, GetCmdPools)
	osmocli.AddQueryCmd(cmd, query.NewQueryClient, GetUserPositions)
	osmocli.AddQueryCmd(cmd, query.NewQueryClient, GetClaimableFees)
	cmd.AddCommand(
		osmocli.GetParams[*query.QueryParamsRequest](
			types.ModuleName, query.NewQueryClient),
	)
	return cmd
}

func GetUserPositions() (*osmocli.QueryDescriptor, *query.QueryUserPositionsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "user-positions [address]",
			Short: "Query user's positions",
			Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} user-positions osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`,
			Flags:               osmocli.FlagDesc{OptionalFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
			CustomFlagOverrides: poolIdFlagOverride},
		&query.QueryUserPositionsRequest{}
}

func GetCmdPool() (*osmocli.QueryDescriptor, *query.QueryPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool [poolID]",
		Short: "Query pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`}, &query.QueryPoolRequest{}
}

func GetCmdPools() (*osmocli.QueryDescriptor, *query.QueryPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`}, &query.QueryPoolsRequest{}
}

func GetClaimableFees() (*osmocli.QueryDescriptor, *query.QueryClaimableFeesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "claimable-fees [poolID] [address] [lowerTick] [upperTick]",
		Short: "Query claimable fees",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} claimable-fees 1 osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj [-100] 100`}, &query.QueryClaimableFeesRequest{}
}
