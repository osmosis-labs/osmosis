package cli

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/client/queryproto"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetCmdPools)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetUserPositions)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetClaimableFees)
	osmocli.AddQueryCmd(cmd, queryproto.NewQueryClient, GetClaimableIncentives)
	cmd.AddCommand(
		osmocli.GetParams[*queryproto.ParamsRequest](
			types.ModuleName, queryproto.NewQueryClient),
	)
	return cmd
}

func GetUserPositions() (*osmocli.QueryDescriptor, *queryproto.UserPositionsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "user-positions [address]",
			Short: "Query user's positions",
			Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} user-positions osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`,
			Flags:               osmocli.FlagDesc{OptionalFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
			CustomFlagOverrides: poolIdFlagOverride},
		&queryproto.UserPositionsRequest{}
}

func GetCmdPools() (*osmocli.QueryDescriptor, *queryproto.PoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`}, &queryproto.PoolsRequest{}
}

func GetClaimableFees() (*osmocli.QueryDescriptor, *queryproto.ClaimableFeesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "claimable-fees [positionID]",
		Short: "Query claimable fees",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} claimable-fees 53`}, &queryproto.ClaimableFeesRequest{}
}

func GetClaimableIncentives() (*osmocli.QueryDescriptor, *queryproto.ClaimableIncentivesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "claimable-incentives [positionID]",
		Short: "Query claimable incentives",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} claimable-fees 53`}, &queryproto.ClaimableIncentivesRequest{}
}
